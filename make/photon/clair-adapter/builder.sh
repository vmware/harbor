#!/bin/bash

set +e

if [ -z $1 ]; then
  error "Please set the 'version' variable"
  exit 1
fi

VERSION="$1"

set -e

# the temp folder to store binary file...
mkdir -p binary
rm -rf binary/harbor-scanner-clair || true

cd $(dirname $0)
cur=$PWD

# The temporary directory to clone Clair adapter source code
TEMP=$(mktemp -d ${TMPDIR-/tmp}/clair-adapter.XXXXXX)
git clone https://github.com/goharbor/harbor-scanner-clair.git $TEMP
cd $TEMP; git checkout $VERSION; export COMMIT=$(git rev-list -1 HEAD); cd -

echo "Building Clair adapter binary based on golang:1.14.5..."
cp Dockerfile.binary $TEMP
docker build --build-arg VERSION=${VERSION} --build-arg COMMIT=${COMMIT} -f $TEMP/Dockerfile.binary -t clair-adapter-golang $TEMP

echo "Copying Clair adapter binary from the container to the local directory..."
ID=$(docker create clair-adapter-golang)
docker cp $ID:/go/src/github.com/goharbor/harbor-scanner-clair/harbor-scanner-clair binary

docker rm -f $ID
docker rmi -f clair-adapter-golang

echo "Building Clair adapter binary finished successfully"
cd $cur
rm -rf $TEMP
