#!/usr/bin/env bash

if [[ $1 == '-h' || $1 == '--help' ]];then
  echo 'Tag the current git branch based on the version in the VERSION file.'
  exit 0
fi

set -e

# make sure we're in the right directory
cd $(dirname $0)

echo "getting project version from VERSION file"
VERSION=$(head -n 1 VERSION)
TAGS=$(git tag -l)

echo "checking existing tags"
if [[ "$TAGS" != '' ]];then
  for tag in "$TAGS";do
    if [[ "$tag" == "$VERSION" ]];then
      echo "tag for version '$VERSION' already exists. the version must be bumped before pushing it as a new tag."
      exit 1
    fi
  done
fi

echo "tagging as '$VERSION'"
git tag -a "$VERSION" -m "$VERSION"
echo "pushing tags"
git push --tags
