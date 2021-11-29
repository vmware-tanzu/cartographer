#!/bin/bash
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

NEW_DOCS_VERSION=${1:-}
if [[ -z $NEW_DOCS_VERSION ]]; then
    echo "ERROR: new version not supplied"
    exit 1
fi

DOCS_DIRECTORY=content/docs
DATA_DOCS_DIRECTORY=data/docs
CONFIG_FILE=config.yaml
DEV_VERSION=development

# don't run if there's already a directory for the target docs version
if [[ -d $DOCS_DIRECTORY/$NEW_DOCS_VERSION ]]; then
    echo "ERROR: $DOCS_DIRECTORY/$NEW_DOCS_VERSION already exists"
    exit 1
fi

# make a copy of the previous versioned docs dir
echo "Creating copy of docs directory $DOCS_DIRECTORY/$DEV_VERSION in $DOCS_DIRECTORY/$NEW_DOCS_VERSION"
cp -r $DOCS_DIRECTORY/${DEV_VERSION}/ $DOCS_DIRECTORY/${NEW_DOCS_VERSION}/

# 'git add' the previous version's docs as-is so we get a useful diff when we copy the $DEV_VERSION docs in
echo "Running 'git add' for previous version's doc contents to use as a base for diff"
git add -f $DOCS_DIRECTORY/${NEW_DOCS_VERSION}

# now copy the contents of $DOCS_DIRECTORY/$DEV_VERSION into the same directory so we can get a nice
# git diff of what changed since previous version
echo "Copying $DOCS_DIRECTORY/$DEV_VERSION/ to $DOCS_DIRECTORY/${NEW_DOCS_VERSION}/"
rm -rf $DOCS_DIRECTORY/${NEW_DOCS_VERSION}/ && cp -r $DOCS_DIRECTORY/$DEV_VERSION/ $DOCS_DIRECTORY/${NEW_DOCS_VERSION}/

# make a copy of the previous versioned ToC
NEW_DOCS_TOC="$(echo ${NEW_DOCS_VERSION} | tr . -)-toc"
PREVIOUS_DOCS_TOC="$(echo ${DEV_VERSION} | tr . -)-toc"

echo "Creating copy of $DATA_DOCS_DIRECTORY/$PREVIOUS_DOCS_TOC.yml at $DATA_DOCS_DIRECTORY/$NEW_DOCS_TOC.yml"
cp $DATA_DOCS_DIRECTORY/$PREVIOUS_DOCS_TOC.yml $DATA_DOCS_DIRECTORY/$NEW_DOCS_TOC.yml

# 'git add' the previous version's ToC content as-is so we get a useful diff when we copy the $DEV_VERSION ToC in
echo "Running 'git add' for previous version's ToC to use as a base for diff"
git add $DATA_DOCS_DIRECTORY/$NEW_DOCS_TOC.yml

# now copy the $DEV_VERSION ToC so we can get a nice git diff of what changed since previous version
echo "Copying $DATA_DOCS_DIRECTORY/$DEV_VERSION-toc.yml to $DATA_DOCS_DIRECTORY/$NEW_DOCS_TOC.yml"
rm $DATA_DOCS_DIRECTORY/$NEW_DOCS_TOC.yml && cp $DATA_DOCS_DIRECTORY/$DEV_VERSION-toc.yml $DATA_DOCS_DIRECTORY/$NEW_DOCS_TOC.yml

# replace known version-specific links -- the sed syntax is slightly different in OS X and Linux,
# so check which OS we're running on.
if [[ $(uname) == "Darwin" ]]; then
    echo "[OS X] updating version-specific links"
    find $DOCS_DIRECTORY/${NEW_DOCS_VERSION} -type f -name "*.md" | xargs sed -i '' "s|https://github.com/vmware-tanzu/cartographer/releases/latest/download|https://github.com/vmware-tanzu/cartographer/releases/download/$NEW_DOCS_VERSION|g"
    find $DOCS_DIRECTORY/${NEW_DOCS_VERSION} -type f -name "*.md" | xargs sed -i '' "s|https://github.com/vmware-tanzu/cartographer/releases/latest|https://github.com/vmware-tanzu/cartographer/releases/tag/$NEW_DOCS_VERSION|g"
    find $DOCS_DIRECTORY/${NEW_DOCS_VERSION} -type f -name "_index.md" | xargs sed -i '' "s|version: $DEV_VERSION|version: $NEW_DOCS_VERSION|g"

    echo "[OS X] Updating latest version in $CONFIG_FILE"
    sed -i '' "s/docs_latest: .*/docs_latest: ${NEW_DOCS_VERSION}/" $CONFIG_FILE

    # newlines and lack of indentation are requirements for this sed syntax
    # which is doing an append
    echo "[OS X] Adding latest version to versions list in $CONFIG_FILE"
    sed -i '' "/- $DEV_VERSION/a\\
\ \ \ \ - ${NEW_DOCS_VERSION}
" $CONFIG_FILE

    echo "[OS X] Adding ToC mapping entry"
    sed -i '' "/$DEV_VERSION: $DEV_VERSION-toc/a\\
${NEW_DOCS_VERSION}: ${NEW_DOCS_TOC}
" $DATA_DOCS_DIRECTORY/toc-mapping.yml

else
    echo "[Linux] updating version-specific links"
    find $DOCS_DIRECTORY/${NEW_DOCS_VERSION} -type f -name "*.md" | xargs sed -i'' "s|https://github.com/vmware-tanzu/cartographer/releases/latest/download|https://github.com/vmware-tanzu/cartographer/releases/download/$NEW_DOCS_VERSION|g"
    find $DOCS_DIRECTORY/${NEW_DOCS_VERSION} -type f -name "*.md" | xargs sed -i'' "s|https://github.com/vmware-tanzu/cartographer/releases/latest|https://github.com/vmware-tanzu/cartographer/releases/tag/$NEW_DOCS_VERSION|g"
    find $DOCS_DIRECTORY/${NEW_DOCS_VERSION} -type f -name "_index.md" | xargs sed -i'' "s|version: $DEV_VERSION|version: $NEW_DOCS_VERSION|g"

    echo "[Linux] Updating latest version in $CONFIG_FILE"
    sed -i'' "s/docs_latest: .*/docs_latest: ${NEW_DOCS_VERSION}/" $CONFIG_FILE

    echo "[Linux] Adding latest version to versions list in $CONFIG_FILE"
    sed -i'' "/- $DEV_VERSION/a \ \ \ \ - ${NEW_DOCS_VERSION}" $CONFIG_FILE

    echo "[Linux] Adding ToC mapping entry"
    sed -i'' "/$DEV_VERSION: $DEV_VERSION-toc/a ${NEW_DOCS_VERSION}: ${NEW_DOCS_TOC}" $DATA_DOCS_DIRECTORY/toc-mapping.yml
fi

echo "Success! $DOCS_DIRECTORY/$NEW_DOCS_VERSION has been created."
