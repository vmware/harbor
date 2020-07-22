#!/bin/bash
set +e

usage(){
  echo "Usage: compile.sh <code path> <code tag> <main.go path> <binary name>"
  echo "e.g: compile.sh github.com/helm/chartmuseum v0.5.1 cmd/chartmuseum chartm"
  exit 1
}

if [ $# != 4 ]; then
  usage
fi

GIT_PATH="$1"
VERSION="$2"
MAIN_GO_PATH="$3"
BIN_NAME="$4"

#Get the source code
git clone $GIT_PATH src_code
ls
SRC_PATH=$(pwd)/src_code
set -e

#Checkout the released tag branch
cd $SRC_PATH
git checkout tags/$VERSION -b $VERSION

#Patch
for p in $(ls /go/bin/*.patch); do
  git apply $p || exit /b 1
done

#Compile
cd $SRC_PATH/$MAIN_GO_PATH && go build -a -o $BIN_NAME
mv $BIN_NAME /go/bin/
