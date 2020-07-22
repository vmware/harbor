#!/bin/bash
set -x

set -e

sudo make package_online GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-travis PKGVERSIONTAG=dev-travis UIVERSIONTAG=dev-travis GOBUILDIMAGE=golang:1.14.5 COMPILETAG=compile_golangimage NOTARYFLAG=true CLAIRFLAG=true CHARTFLAG=true TRIVYFLAG=true HTTPPROXY=
sudo make package_offline GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-travis PKGVERSIONTAG=dev-travis UIVERSIONTAG=dev-travis GOBUILDIMAGE=golang:1.14.5 COMPILETAG=compile_golangimage NOTARYFLAG=true CLAIRFLAG=true CHARTFLAG=true TRIVYFLAG=true HTTPPROXY=
