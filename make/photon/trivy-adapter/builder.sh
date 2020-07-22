#!/bin/bash

set +e

if [ -z $1 ]; then
  error "Please set the 'version' variable"
  exit 1
fi

VERSION="$1"

set -e

cd $(dirname $0)
cur=$PWD

# The temporary directory to clone Trivy adapter source code
TEMP=$(mktemp -d ${TMPDIR-/tmp}/trivy-adapter.XXXXXX)
git clone https://github.com/aquasecurity/harbor-scanner-trivy.git $TEMP
cd $TEMP; git checkout $VERSION; cd -

echo "Building Trivy adapter binary based on golang:1.14.5..."
cp Dockerfile.binary $TEMP
docker build -f $TEMP/Dockerfile.binary -t trivy-adapter-golang $TEMP

echo "Copying Trivy adapter binary from the container to the local directory..."
ID=$(docker create trivy-adapter-golang)
docker cp $ID:/go/src/github.com/aquasecurity/harbor-scanner-trivy/scanner-trivy binary

docker rm -f $ID
docker rmi -f trivy-adapter-golang

echo "Building Trivy adapter binary finished successfully"
cd $cur
rm -rf $TEMP
