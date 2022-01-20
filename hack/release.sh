#!/usr/bin/env bash
# Copyright 2021 VMware
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(cd "$(dirname $0)"/.. && pwd)
readonly ROOT

readonly SCRATCH=${SCRATCH:-$(mktemp -d)}
readonly REGISTRY=${REGISTRY:-"$($ROOT/hack/ip.py):5000"}
readonly BUNDLE=${BUNDLE:-$REGISTRY/cartographer-bundle}
readonly RELEASE_DATE=${RELEASE_DATE:-$(TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")}

readonly YTT_VERSION=0.38.0
readonly YTT_CHECKSUM=2ca800c561464e0b252e5ee5cacff6aa53831e65e2fb9a09cf388d764013c40d

main() {
        readonly RELEASE_VERSION="v0.0.0-dev"
        readonly PREVIOUS_VERSION=${PREVIOUS_VERSION:-$(git_previous_version $RELEASE_VERSION)}

        show_vars
        cd $ROOT

        download_ytt_to_kodata
        create_imgpkg_bundle
        create_carvel_packaging_objects

        populate_release_directory
        create_release_notes
}

show_vars() {
        echo "Vars:

	BUNDLE:	       		$BUNDLE
	REGISTRY:		      $REGISTRY
	RELEASE_DATE:		  $RELEASE_DATE
	RELEASE_VERSION:	$RELEASE_VERSION
	PREVIOUS_VERSION:	$PREVIOUS_VERSION
	ROOT:	       		  $ROOT
	SCRATCH:       		$SCRATCH
	YTT_VERSION		    $YTT_VERSION
	"
}

download_ytt_to_kodata() {
        local url=https://github.com/vmware-tanzu/carvel-ytt/releases/download/v${YTT_VERSION}/ytt-linux-amd64
        local fname=ytt-linux-amd64

        local dest
        dest=$(realpath ./cmd/cartographer/kodata/$fname)

        test -x $dest && echo "${YTT_CHECKSUM}  $dest" | sha256sum -c && {
                echo "ytt already found in kodata."
                return
        }

        pushd "$(mktemp -d)"
        curl -sSOL $url
        echo "${YTT_CHECKSUM}  $fname" | sha256sum -c
        install -m 0755 $fname $dest
        popd
}

# creates, in a scratch location, an imgpkg bundle following the convention that
# is expected of bundles for Packages (see ref):
#
#
# 	$scratch
# 	├── bundle
# 	│   ├── .imgpkg
# 	│   │   └── images.yml			absolute image references to
# 	│   │					images in this bundle
# 	│   │
# 	│   └── config
# 	│       ├── cartographer.yaml		everything from the release
# 	│       │
# 	│       ├── objects/			extra objects to include in the
# 	│       │				bundle to aid the installation
# 	│       │
# 	│       └── overlays/			overlays to tweak properties
# 	│                       		from the release according to
# 	│                       		packaging configuration
# 	│
# 	├── bundle.tar				tarball of the imgpkg bundle
# 	│
# 	│
# 	└── bundle.lock.yaml			exact image reference to this
# 						bundle
#
#
# ref: https://carvel.dev/kapp-controller/docs/latest/packaging-artifact-formats/#package-contents-bundle
#
create_imgpkg_bundle() {
        mkdir -p $SCRATCH/bundle/{.imgpkg,config}

        cp -r ./packaging/{objects,overlays} $SCRATCH/bundle/config

        ytt --ignore-unknown-comments -f ./config |
                KO_DOCKER_REPO=$REGISTRY ko resolve -B -f- > \
                        $SCRATCH/bundle/config/cartographer.yaml

        kbld -f $SCRATCH/bundle/config/cartographer.yaml \
                --imgpkg-lock-output $SCRATCH/bundle/.imgpkg/images.yml \
                >/dev/null

        imgpkg push -f $SCRATCH/bundle \
                --bundle $BUNDLE \
                --lock-output $SCRATCH/bundle.lock.yaml

        imgpkg copy \
                --bundle "$(image_from_lockfile $SCRATCH/bundle.lock.yaml)" \
                --to-tar $SCRATCH/bundle.tar
}

create_carvel_packaging_objects() {
        mkdir -p $SCRATCH/package

        for package_fpath in ./packaging/package*.yaml; do
                ytt --ignore-unknown-comments \
                        -f ./packaging/values.yaml \
                        -f $package_fpath \
                        --data-value image="$(image_from_lockfile $SCRATCH/bundle.lock.yaml)" \
                        --data-value releasedAt=$RELEASE_DATE \
                        --data-value version=${RELEASE_VERSION#v} > \
                        $SCRATCH/package/"$(basename $package_fpath)"
        done

}

create_release_notes() {
        local changeset
        changeset="$(git_changeset $RELEASE_VERSION $PREVIOUS_VERSION)"

        local assets_checksums
        assets_checksums=$(checksums ./release)

        release_body "$changeset" "$assets_checksums" >./release/CHANGELOG.md
}

# generates the final release directory containing the files that are meant to
# be used during installation.
#
#
# 	release
# 	│
# 	├── package
# 	│   └── package-install.yaml
# 	│   └── package-metadata.yaml
# 	│   └── package.yaml
# 	│
# 	├── cartographer.yaml
# 	└── bundle.tar
#
populate_release_directory() {
        rm -rf ./release
        mkdir -p ./release/package

        cp $SCRATCH/bundle.tar ./release
        cp $SCRATCH/bundle/config/cartographer.yaml ./release
        cp -r $SCRATCH/package ./release
}

image_from_lockfile() {
        local lockfile=$1

        awk -F"image: " '{if ($2) print $2;}' $lockfile
}

checksums() {
        local assets_directory=$1

        pushd $assets_directory &>/dev/null
        find . -name "*" -type f -exec sha256sum {} +
        popd &>/dev/null
}

git_changeset() {
        local current_version=$1
        local previous_version=$2

        [[ $current_version != v* ]] && current_version=v$current_version
        [[ $previous_version != v* ]] && previous_version=v$previous_version
        [[ $(git tag -l $current_version) == "" ]] && current_version=HEAD

        git -c log.showSignature=false \
                log \
                --pretty=oneline \
                --abbrev-commit \
                --no-decorate \
                --no-color \
                "${previous_version}..${current_version}"
}

git_previous_version() {
        local current_version=$1

        local version_filter
        version_filter=$(printf '^%s$' $current_version)

        [[ $(git tag -l $current_version) == "" ]] && version_filter='.'

        git tag --sort=-v:refname -l |
                grep -A30 $version_filter |
                tail -n +2 |
                grep -E '^v[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+$' |
                head -n1
}

release_body() {
        local changeset="$1"
        local checksums="$2"

        readonly fmt='
# Change Set

%s


# Installation

See https://github.com/vmware-tanzu/cartographer#installation.


# Checksums

```
%s
```
  '
        printf "$fmt" "$changeset" "$checksums"
}

main "$@"
