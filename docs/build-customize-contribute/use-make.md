---
title: Using Make
---

## Variables

Variable           | Description
-------------------|-------------
BASEIMAGE          | Container base image, default: photon
DEVFLAG            | Build model flag, default: dev
COMPILETAG         | Compile model flag, default: compile_normal (local golang build)
GOBUILDIMAGE       | Golang image to compile harbor go source code.
NOTARYFLAG         | Whether to enable notary in harbor, default:false
HTTPPROXY          | Clarity proxy to build UI.

## Targets

Target              | Description
--------------------|-------------
all                 | prepare env, compile binaries, build images and install images
prepare             | prepare env
compile             | compile ui and jobservice code
compile_portal      | compile portal code
compile_core        | compile core binary
compile_jobservice  | compile jobservice binary
build               | build Harbor docker images (default: using build_photon)
build_photon        | build Harbor docker images from Photon OS base image
install             | compile binaries, build images, prepare specific version of compose file and startup Harbor instance
start               | startup Harbor instance
down                | shutdown Harbor instance
package_online      | prepare online install package
package_offline     | prepare offline install package
pushimage           | push Harbor images to specific registry server
clean all           | remove binary, Harbor images, specific version docker-compose file, specific version tag and online/offline install package
cleanbinary         | remove ui and jobservice binary
cleanimage          | remove Harbor images
cleandockercomposefile  | remove specific version docker-compose
cleanversiontag     | remove specific version tag
cleanpackage        | remove online/offline install package
version				 | set harbor version

## Examples

### Build and run harbor from source code

```sh
make install GOBUILDIMAGE=golang:1.14.5 COMPILETAG=compile_golangimage NOTARYFLAG=true
```

### Package offline installer

```sh
make package_offline GOBUILDIMAGE=golang:1.14.5 COMPILETAG=compile_golangimage NOTARYFLAG=true
```

### Start harbor with notary

```sh
make -e NOTARYFLAG=true start
```

### Stop harbor with notary

```sh
make -e NOTARYFLAG=true down
```
