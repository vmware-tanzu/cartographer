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

readonly ROOT=$(cd $(dirname $0)/.. && pwd)
readonly SCRATCH=${SCRATCH:-$(mktemp -d)}
readonly REGISTRY=${REGISTRY:-"$($ROOT/hack/ip.py):5000"}
readonly BUNDLE=${BUNDLE:-$REGISTRY/cartographer-bundle}
readonly RELEASE_VERSION=${RELEASE_VERSION:-"0.0.0-dev"}
readonly RELEASE_DATE=${RELEASE_DATE:-$(date -Iseconds)}

readonly YTT_VERSION=0.36.0
readonly YTT_CHECKSUM=d81ecf6c47209f6ac527e503a6fd85e999c3c2f8369e972794047bddc7e5fbe2

main() {
        show_vars

        cd $ROOT
        download_ytt_to_kodata
        create_imgpkg_bundle
        generate_release
}

show_vars() {
        echo "Vars:

	BUNDLE:	       		$BUNDLE
	REGISTRY:		$REGISTRY
	RELEASE_DATE:		$RELEASE_DATE
	RELEASE_VERSION:	$RELEASE_VERSION
	ROOT:	       		$ROOT
	SCRATCH:       		$SCRATCH
	YTT_VERSION		$YTT_VERSION
	"
}

download_ytt_to_kodata() {
        local url=https://github.com/vmware-tanzu/carvel-ytt/releases/download/v${YTT_VERSION}/ytt-linux-amd64
        local fname=ytt-linux-amd64
        local dest=$(realpath ./cmd/cartographer/kodata/$fname)

        test -x $dest && echo "${YTT_CHECKSUM} $dest" | sha256sum -c && {
                echo "ytt already found in kodata."
                return
        }

        pushd $(mktemp -d)
        curl -sSOL $url
        echo "${YTT_CHECKSUM} $fname" | sha256sum -c
        install -m 0755 $fname $dest
        popd
}

# creates, in a scratch location, an imgpkg bundle following the convention that
# is expected of bundles for Packages (see ref):
#
#
# 	.
# 	├── bundle
# 	│   ├── .imgpkg
# 	│   │   └── images.yml			absolute image references to
# 	│   │					images in this bundle
# 	│   │
# 	│   └── config
# 	│       └── cartographer.yaml		everything from the release
# 	│       │
# 	│       └── overlays/			overlays to tweak properties
# 	│                       		from the release according to
# 	│                       		packaging configuration
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
                KO_DOCKER_REPO=$REGISTRY ko resolve -f- > \
                        $SCRATCH/bundle/config/cartographer.yaml

        kbld -f $SCRATCH/bundle/config/cartographer.yaml \
                --imgpkg-lock-output $SCRATCH/bundle/.imgpkg/images.yml \
                >/dev/null

        imgpkg push -f $SCRATCH/bundle \
                --bundle $BUNDLE \
                --lock-output $SCRATCH/bundle.lock.yaml

        imgpkg copy \
                --bundle $(image_from_lockfile $SCRATCH/bundle.lock.yaml) \
                --to-tar $SCRATCH/bundle.tar
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
# 	├── bundle.tar.gz
# 	│
# 	└── release.yaml
#
generate_release() {
        rm -rf ./release
        mkdir -p ./release/package

        cp $SCRATCH/bundle.tar ./release
        cp $SCRATCH/bundle/config/cartographer.yaml ./release
        for package_fpath in ./packaging/package*.yaml; do
                ytt --ignore-unknown-comments \
                        -f ./packaging/values.yaml \
                        -f $package_fpath \
                        --data-value image=$(image_from_lockfile $SCRATCH/bundle.lock.yaml) \
                        --data-value releasedAt=$RELEASE_DATE \
                        --data-value version=$RELEASE_VERSION > \
                        ./release/package/$(basename $package_fpath)
        done
}

image_from_lockfile() {
        local lockfile=$1

        awk -F"image: " '{if ($2) print $2;}' $lockfile
}

main "$@"
