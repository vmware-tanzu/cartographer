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


echo "# inclusiveness analysis"

function cleanup {
  echo "# removing temp dir"
  rm -rf ./temp-inclusiveness
  echo "# completed: inclusiveness analysis"
}

trap cleanup EXIT

mkdir -p ./temp-inclusiveness

echo "# downloading sensitive terms"
curl https://s3.amazonaws.com/srp-cli/api-rules/regex > ./temp-inclusiveness/regex

echo "# selecting files to scan"
find . \( -name "*.go" -o -name "*.yaml" -o -name "*.md" -o -name "*.sh" \) | grep -vE "/node_modules/|/target/|/dist/" > ./temp-inclusiveness/files
echo -e "NOTICE\nMakefile\n" >> ./temp-inclusiveness/files


echo "# scanning files"
IFS=$'\n'
for file in $(cat ./temp-inclusiveness/files); do
  grep -iHnP -f ./temp-inclusiveness/regex $file >> ./temp-inclusiveness/result || true;
done
unset IFS

echo "# excluding false positives and other exclusions"
cat .inclusive-exclusions | \
  sed '/^[[:blank:]]*#/d;s/\s*#.*//' | \
  grep -v "^$" | \
  grep -v -f - ./temp-inclusiveness/result \
  > ./temp-inclusiveness/filtered-result

echo "# checking result"
if [ -s ./temp-inclusiveness/filtered-result ]
then
  echo -e "\nERROR: there are issues with sensitive terms\n"
  cat ./temp-inclusiveness/filtered-result
  echo
  exit 1
fi
