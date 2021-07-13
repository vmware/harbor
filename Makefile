# Makefile for Harbor project
#
# Targets:
#
# all:			prepare env, compile binaries, build images and install images
# prepare: 		prepare env
# compile: 		compile core and jobservice code
#
# compile_golangimage:
#			compile from golang image
#			for example: make compile_golangimage -e GOBUILDIMAGE= \
#							golang:1.16.5
# compile_core, compile_jobservice: compile specific binary
#
# build:	build Harbor docker images from photon baseimage
#
# install:		include compile binaries, build images, prepare specific \
#				version composefile and startup Harbor instance
#
# start:		startup Harbor instance
#
# down:			shutdown Harbor instance
#
# package_online:
#				prepare online install package
#			for example: make package_online -e DEVFLAG=false\
#							REGISTRYSERVER=reg-bj.goharbor.io \
#							REGISTRYPROJECTNAME=harborrelease
#
# package_offline:
#				prepare offline install package
#
# pushimage:	push Harbor images to specific registry server
#			for example: make pushimage -e DEVFLAG=false REGISTRYUSER=admin \
#							REGISTRYPASSWORD=***** \
#							REGISTRYSERVER=reg-bj.goharbor.io/ \
#							REGISTRYPROJECTNAME=harborrelease
#				note**: need add "/" on end of REGISTRYSERVER. If not setting \
#						this value will push images directly to dockerhub.
#						 make pushimage -e DEVFLAG=false REGISTRYUSER=goharbor \
#							REGISTRYPASSWORD=***** \
#							REGISTRYPROJECTNAME=goharbor
#
# clean:        remove binary, Harbor images, specific version docker-compose \
#               file, specific version tag and online/offline install package
# cleanbinary:	remove core and jobservice binary
# cleanbaseimage:
#               remove the base images of Harbor images
# cleanimage: 	remove Harbor images
# cleandockercomposefile:
#				remove specific version docker-compose
# cleanversiontag:
#				cleanpackageremove specific version tag
# cleanpackage: remove online/offline install package
#
# other example:
#	clean specific version binaries and images:
#				make clean -e VERSIONTAG=[TAG]
#				note**: If commit new code to github, the git commit TAG will \
#				change. Better use this command clean previous images and \
#				files with specific TAG.
#   By default DEVFLAG=true, if you want to release new version of Harbor, \
#		should setting the flag to false.
#				make XXXX -e DEVFLAG=false

SHELL := /bin/bash
BUILDPATH=$(CURDIR)
MAKEPATH=$(BUILDPATH)/make
MAKE_PREPARE_PATH=$(MAKEPATH)/photon/prepare
SRCPATH=./src
TOOLSPATH=$(BUILDPATH)/tools
CORE_PATH=$(BUILDPATH)/src/core
PORTAL_PATH=$(BUILDPATH)/src/portal
CHECKENVCMD=checkenv.sh

# parameters
# default is true
BUILD_PG96=true
REGISTRYSERVER=
REGISTRYPROJECTNAME=goharbor
DEVFLAG=true
NOTARYFLAG=false
TRIVYFLAG=false
HTTPPROXY=
BUILDBIN=false
NPM_REGISTRY=https://registry.npmjs.org
# enable/disable chart repo supporting
CHARTFLAG=false
BUILDTARGET=build
GEN_TLS=

# version prepare
# for docker image tag
VERSIONTAG=dev
# for base docker image tag
BUILD_BASE=true
PUSHBASEIMAGE=false
BASEIMAGETAG=dev
BUILDBASETARGET=chartserver trivy-adapter core db jobservice log nginx notary-server notary-signer portal prepare redis registry registryctl exporter
BASEIMAGENAMESPACE=goharbor
# #input true/false only
PULL_BASE_FROM_DOCKERHUB=true

# for harbor package name
PKGVERSIONTAG=dev

PREPARE_VERSION_NAME=versions

#versions
REGISTRYVERSION=v2.7.1-patch-2819-2553-redis
NOTARYVERSION=v0.6.1
NOTARYMIGRATEVERSION=v4.11.0
TRIVYVERSION=v0.17.2
TRIVYADAPTERVERSION=v0.19.0

# version of chartmuseum for pulling the source code
CHARTMUSEUM_SRC_TAG=v0.13.1

# version of chartmuseum
CHARTMUSEUMVERSION=$(CHARTMUSEUM_SRC_TAG)-redis

# version of registry for pulling the source code
REGISTRY_SRC_TAG=v2.7.1

# dependency binaries
CHARTURL=https://storage.googleapis.com/harbor-builds/bin/chartmuseum/release-${CHARTMUSEUMVERSION}/chartm
NOTARYURL=https://storage.googleapis.com/harbor-builds/bin/notary/release-${NOTARYVERSION}/binary-bundle.tgz
REGISTRYURL=https://storage.googleapis.com/harbor-builds/bin/registry/release-${REGISTRYVERSION}/registry
TRIVY_DOWNLOAD_URL=https://github.com/aquasecurity/trivy/releases/download/$(TRIVYVERSION)/trivy_$(TRIVYVERSION:v%=%)_Linux-64bit.tar.gz
TRIVY_ADAPTER_DOWNLOAD_URL=https://github.com/aquasecurity/harbor-scanner-trivy/releases/download/$(TRIVYADAPTERVERSION)/harbor-scanner-trivy_$(TRIVYADAPTERVERSION:v%=%)_Linux_x86_64.tar.gz

define VERSIONS_FOR_PREPARE
VERSION_TAG: $(VERSIONTAG)
REGISTRY_VERSION: $(REGISTRYVERSION)
NOTARY_VERSION: $(NOTARYVERSION)
TRIVY_VERSION: $(TRIVYVERSION)
TRIVY_ADAPTER_VERSION: $(TRIVYADAPTERVERSION)
CHARTMUSEUM_VERSION: $(CHARTMUSEUMVERSION)
endef

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERPULL=$(DOCKERCMD) pull
DOCKERIMAGES=$(DOCKERCMD) images
DOCKERSAVE=$(DOCKERCMD) save
DOCKERCOMPOSECMD=$(shell which docker-compose)
DOCKERTAG=$(DOCKERCMD) tag

# go parameters
GOCMD=$(shell which go)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w
GOBUILDIMAGE=golang:1.16.5
GOBUILDPATHINCONTAINER=/harbor

# go build
PKG_PATH=github.com/goharbor/harbor/src/pkg
GITCOMMIT := $(shell git rev-parse --short=8 HEAD)
RELEASEVERSION := $(shell cat VERSION)
GOFLAGS=
GOTAGS=$(if $(GOBUILDTAGS),-tags "$(GOBUILDTAGS)",)
GOLDFLAGS=$(if $(GOBUILDLDFLAGS),--ldflags "-w -s $(GOBUILDLDFLAGS)",)
CORE_LDFLAGS=-X $(PKG_PATH)/version.GitCommit=$(GITCOMMIT) -X $(PKG_PATH)/version.ReleaseVersion=$(RELEASEVERSION)
ifneq ($(GOBUILDLDFLAGS),)
	CORE_LDFLAGS += $(GOBUILDLDFLAGS)
endif

# go build command
GOIMAGEBUILDCMD=/usr/local/go/bin/go build -mod vendor
GOIMAGEBUILD_COMMON=$(GOIMAGEBUILDCMD) $(GOFLAGS) ${GOTAGS} ${GOLDFLAGS}
GOIMAGEBUILD_CORE=$(GOIMAGEBUILDCMD) $(GOFLAGS) ${GOTAGS} --ldflags "-w -s $(CORE_LDFLAGS)"

GOBUILDPATH_CORE=$(GOBUILDPATHINCONTAINER)/src/core
GOBUILDPATH_JOBSERVICE=$(GOBUILDPATHINCONTAINER)/src/jobservice
GOBUILDPATH_REGISTRYCTL=$(GOBUILDPATHINCONTAINER)/src/registryctl
GOBUILDPATH_MIGRATEPATCH=$(GOBUILDPATHINCONTAINER)/src/cmd/migrate-patch
GOBUILDPATH_STANDALONE_DB_MIGRATOR=$(GOBUILDPATHINCONTAINER)/src/cmd/standalone-db-migrator
GOBUILDPATH_EXPORTER=$(GOBUILDPATHINCONTAINER)/src/cmd/exporter
GOBUILDMAKEPATH=make
GOBUILDMAKEPATH_CORE=$(GOBUILDMAKEPATH)/photon/core
GOBUILDMAKEPATH_JOBSERVICE=$(GOBUILDMAKEPATH)/photon/jobservice
GOBUILDMAKEPATH_REGISTRYCTL=$(GOBUILDMAKEPATH)/photon/registryctl
GOBUILDMAKEPATH_NOTARY=$(GOBUILDMAKEPATH)/photon/notary
GOBUILDMAKEPATH_STANDALONE_DB_MIGRATOR=$(GOBUILDMAKEPATH)/photon/standalone-db-migrator
GOBUILDMAKEPATH_EXPORTER=$(GOBUILDMAKEPATH)/photon/exporter

# binary
CORE_BINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_CORE)
CORE_BINARYNAME=harbor_core
JOBSERVICEBINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_JOBSERVICE)
JOBSERVICEBINARYNAME=harbor_jobservice
REGISTRYCTLBINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_REGISTRYCTL)
REGISTRYCTLBINARYNAME=harbor_registryctl
MIGRATEPATCHBINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_NOTARY)
MIGRATEPATCHBINARYNAME=migrate-patch
STANDALONE_DB_MIGRATOR_BINARYPATH=$(BUILDPATH)/$(GOBUILDMAKEPATH_STANDALONE_DB_MIGRATOR)
STANDALONE_DB_MIGRATOR_BINARYNAME=migrate

# configfile
CONFIGPATH=$(MAKEPATH)
INSIDE_CONFIGPATH=/compose_location
CONFIGFILE=harbor.yml

# prepare parameters
PREPAREPATH=$(TOOLSPATH)
PREPARECMD=prepare
PREPARECMD_PARA=--conf $(INSIDE_CONFIGPATH)/$(CONFIGFILE)
ifeq ($(NOTARYFLAG), true)
	PREPARECMD_PARA+= --with-notary
endif
ifeq ($(TRIVYFLAG), true)
	PREPARECMD_PARA+= --with-trivy
endif
# append chartmuseum parameters if set
ifeq ($(CHARTFLAG), true)
    PREPARECMD_PARA+= --with-chartmuseum
endif

# makefile
MAKEFILEPATH_PHOTON=$(MAKEPATH)/photon

# common dockerfile
DOCKERFILEPATH_COMMON=$(MAKEPATH)/common

# docker image name
DOCKER_IMAGE_NAME_PREPARE=$(BASEIMAGENAMESPACE)/prepare
DOCKERIMAGENAME_PORTAL=$(BASEIMAGENAMESPACE)/harbor-portal
DOCKERIMAGENAME_CORE=$(BASEIMAGENAMESPACE)/harbor-core
DOCKERIMAGENAME_JOBSERVICE=$(BASEIMAGENAMESPACE)/harbor-jobservice
DOCKERIMAGENAME_LOG=$(BASEIMAGENAMESPACE)/harbor-log
DOCKERIMAGENAME_DB=$(BASEIMAGENAMESPACE)/harbor-db
DOCKERIMAGENAME_CHART_SERVER=$(BASEIMAGENAMESPACE)/chartmuseum-photon
DOCKERIMAGENAME_REGCTL=$(BASEIMAGENAMESPACE)/harbor-registryctl
DOCKERIMAGENAME_EXPORTER=$(BASEIMAGENAMESPACE)/harbor-exporter

# docker-compose files
DOCKERCOMPOSEFILEPATH=$(MAKEPATH)
DOCKERCOMPOSEFILENAME=docker-compose.yml

SEDCMD=$(shell which sed)
SEDCMDI=$(SEDCMD) -i
ifeq ($(shell uname),Darwin)
    SEDCMDI=$(SEDCMD) -i ''
endif

# package
TARCMD=$(shell which tar)
ZIPCMD=$(shell which gzip)
DOCKERIMGFILE=harbor
HARBORPKG=harbor

# pull/push image
PUSHSCRIPTPATH=$(MAKEPATH)
PUSHSCRIPTNAME=pushimage.sh
REGISTRYUSER=
REGISTRYPASSWORD=

# cmds
DOCKERSAVE_PARA=$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_CORE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_REGCTL):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_EXPORTER):$(VERSIONTAG) \
		$(BASEIMAGENAMESPACE)/redis-photon:$(VERSIONTAG) \
		$(BASEIMAGENAMESPACE)/nginx-photon:$(VERSIONTAG) \
		$(BASEIMAGENAMESPACE)/registry-photon:$(VERSIONTAG)

PACKAGE_OFFLINE_PARA=-zcvf harbor-offline-installer-$(PKGVERSIONTAG).tgz \
					$(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz \
					$(HARBORPKG)/prepare \
					$(HARBORPKG)/LICENSE $(HARBORPKG)/install.sh \
					$(HARBORPKG)/common.sh \
					$(HARBORPKG)/harbor.yml.tmpl

PACKAGE_ONLINE_PARA=-zcvf harbor-online-installer-$(PKGVERSIONTAG).tgz \
					$(HARBORPKG)/prepare \
					$(HARBORPKG)/LICENSE \
					$(HARBORPKG)/install.sh \
					$(HARBORPKG)/common.sh \
					$(HARBORPKG)/harbor.yml.tmpl

DOCKERCOMPOSE_FILE_OPT=-f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)

ifeq ($(NOTARYFLAG), true)
	DOCKERSAVE_PARA+= $(BASEIMAGENAMESPACE)/notary-server-photon:$(VERSIONTAG) $(BASEIMAGENAMESPACE)/notary-signer-photon:$(VERSIONTAG)
endif
ifeq ($(TRIVYFLAG), true)
	DOCKERSAVE_PARA+= $(BASEIMAGENAMESPACE)/trivy-adapter-photon:$(VERSIONTAG)
endif
# append chartmuseum parameters if set
ifeq ($(CHARTFLAG), true)
	DOCKERSAVE_PARA+= $(DOCKERIMAGENAME_CHART_SERVER):$(VERSIONTAG)
endif


RUNCONTAINER=$(DOCKERCMD) run --rm -u $(shell id -u):$(shell id -g) -v $(BUILDPATH):$(BUILDPATH) -w $(BUILDPATH)

# $1 the name of the docker image
# $2 the tag of the docker image
# $3 the command to build the docker image
define prepare_docker_image
	@if [ "$(shell ${DOCKERIMAGES} -q $(1):$(2) 2> /dev/null)" == "" ]; then \
		$(3) && echo "build $(1):$(2) successfully" || (echo "build $(1):$(2) failed" && exit 1) ; \
	fi
endef

# lint swagger doc
SPECTRAL_IMAGENAME=$(BASEIMAGENAMESPACE)/spectral
SPECTRAL_VERSION=v5.9.1
SPECTRAL_IMAGE_BUILD_CMD=${DOCKERBUILD} -f ${TOOLSPATH}/spectral/Dockerfile --build-arg GOLANG=${GOBUILDIMAGE} --build-arg SPECTRAL_VERSION=${SPECTRAL_VERSION} -t ${SPECTRAL_IMAGENAME}:$(SPECTRAL_VERSION) .
SPECTRAL=$(RUNCONTAINER) $(SPECTRAL_IMAGENAME):$(SPECTRAL_VERSION)

lint_apis:
	$(call prepare_docker_image,${SPECTRAL_IMAGENAME},${SPECTRAL_VERSION},${SPECTRAL_IMAGE_BUILD_CMD})
	$(SPECTRAL) lint ./api/v2.0/swagger.yaml

SWAGGER_IMAGENAME=$(BASEIMAGENAMESPACE)/swagger
SWAGGER_VERSION=v0.25.0
SWAGGER=$(RUNCONTAINER) ${SWAGGER_IMAGENAME}:${SWAGGER_VERSION}
SWAGGER_GENERATE_SERVER=${SWAGGER} generate server --template-dir=$(TOOLSPATH)/swagger/templates --exclude-main --additional-initialism=CVE --additional-initialism=GC --additional-initialism=OIDC
SWAGGER_IMAGE_BUILD_CMD=${DOCKERBUILD} -f ${TOOLSPATH}/swagger/Dockerfile --build-arg GOLANG=${GOBUILDIMAGE} --build-arg SWAGGER_VERSION=${SWAGGER_VERSION} -t ${SWAGGER_IMAGENAME}:$(SWAGGER_VERSION) .

# $1 the path of swagger spec
# $2 the path of base directory for generating the files
# $3 the name of the application
define swagger_generate_server
	@echo "generate all the files for API from $(1)"
	@rm -rf $(2)/{models,restapi}
	@mkdir -p $(2)
	@$(SWAGGER_GENERATE_SERVER) -f $(1) -A $(3) --target $(2)
endef

gen_apis: lint_apis
	$(call prepare_docker_image,${SWAGGER_IMAGENAME},${SWAGGER_VERSION},${SWAGGER_IMAGE_BUILD_CMD})
	$(call swagger_generate_server,api/v2.0/swagger.yaml,src/server/v2.0,harbor)


MOCKERY_IMAGENAME=$(BASEIMAGENAMESPACE)/mockery
MOCKERY_VERSION=v2.1.0
MOCKERY=$(RUNCONTAINER) ${MOCKERY_IMAGENAME}:${MOCKERY_VERSION}
MOCKERY_IMAGE_BUILD_CMD=${DOCKERBUILD} -f ${TOOLSPATH}/mockery/Dockerfile --build-arg GOLANG=${GOBUILDIMAGE} --build-arg MOCKERY_VERSION=${MOCKERY_VERSION} -t ${MOCKERY_IMAGENAME}:$(MOCKERY_VERSION) .

gen_mocks:
	$(call prepare_docker_image,${MOCKERY_IMAGENAME},${MOCKERY_VERSION},${MOCKERY_IMAGE_BUILD_CMD})
	${MOCKERY} go generate ./...

mocks_check: gen_mocks
	@echo checking mocks...
	@res=$$(git status -s src/ | awk '{ printf("%s\n", $$2) }' | egrep .*.go); \
	if [ -n "$${res}" ]; then \
		echo mocks of the interface are out of date... ; \
		echo "$${res}"; \
		exit 1; \
	fi

export VERSIONS_FOR_PREPARE
versions_prepare:
	@echo "$$VERSIONS_FOR_PREPARE" > $(MAKE_PREPARE_PATH)/$(PREPARE_VERSION_NAME)

check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

compile_core: gen_apis
	@echo "compiling binary for core (golang image)..."
	@echo $(GOBUILDPATHINCONTAINER)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_CORE) $(GOBUILDIMAGE) $(GOIMAGEBUILD_CORE) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_CORE)/$(CORE_BINARYNAME)
	@echo "Done."

compile_jobservice:
	@echo "compiling binary for jobservice (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_JOBSERVICE) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_JOBSERVICE)/$(JOBSERVICEBINARYNAME)
	@echo "Done."

compile_registryctl:
	@echo "compiling binary for harbor registry controller (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_REGISTRYCTL) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_REGISTRYCTL)/$(REGISTRYCTLBINARYNAME)
	@echo "Done."

compile_notary_migrate_patch:
	@echo "compiling binary for migrate patch (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_MIGRATEPATCH) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_NOTARY)/$(MIGRATEPATCHBINARYNAME)
	@echo "Done."

compile_standalone_db_migrator:
	@echo "compiling binary for standalone db migrator (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(GOBUILDPATH_STANDALONE_DB_MIGRATOR) $(GOBUILDIMAGE) $(GOIMAGEBUILD_COMMON) -o $(GOBUILDPATHINCONTAINER)/$(GOBUILDMAKEPATH_STANDALONE_DB_MIGRATOR)/$(STANDALONE_DB_MIGRATOR_BINARYNAME)
	@echo "Done."

compile: check_environment versions_prepare compile_core compile_jobservice compile_registryctl compile_notary_migrate_patch

update_prepare_version:
	@echo "substitute the prepare version tag in prepare file..."
	@$(SEDCMDI) -e 's/goharbor\/prepare:.*[[:space:]]\+/goharbor\/prepare:$(VERSIONTAG) prepare /' $(MAKEPATH)/prepare ;

gen_tls:
	@$(DOCKERCMD) run --rm -v /:/hostfs:z $(BASEIMAGENAMESPACE)/prepare:$(VERSIONTAG) gencert -p /etc/harbor/tls/internal

prepare: update_prepare_version
	@echo "preparing..."
	@if [ -n "$(GEN_TLS)" ] ; then \
		$(DOCKERCMD) run --rm -v /:/hostfs:z $(BASEIMAGENAMESPACE)/prepare:$(VERSIONTAG) gencert -p /etc/harbor/tls/internal; \
	fi
	@$(MAKEPATH)/$(PREPARECMD) $(PREPARECMD_PARA)

build:
# PUSHBASEIMAGE should not be true if BUILD_BASE is not true
	@if [ "$(PULL_BASE_FROM_DOCKERHUB)" != "true" ] && [ "$(PULL_BASE_FROM_DOCKERHUB)" != "false" ] ; then \
		echo set PULL_BASE_FROM_DOCKERHUB to true or false.; exit 1; \
	fi
	@if [ "$(BUILD_BASE)" != "true" ]  && [ "$(PUSHBASEIMAGE)" = "true" ] ; then \
		echo Do not push base images since no base images built. ; \
		exit 1; \
	fi
# PULL_BASE_FROM_DOCKERHUB should be true if BUILD_BASE is not true
	@if [ "$(BUILD_BASE)" != "true" ]  && [ "$(PULL_BASE_FROM_DOCKERHUB)" = "false" ] ; then \
		echo Should pull base images from registry in docker configuration since no base images built. ; \
		exit 1; \
	fi
	make -f $(MAKEFILEPATH_PHOTON)/Makefile $(BUILDTARGET) -e DEVFLAG=$(DEVFLAG) -e GOBUILDIMAGE=$(GOBUILDIMAGE) \
	 -e REGISTRYVERSION=$(REGISTRYVERSION) -e REGISTRY_SRC_TAG=$(REGISTRY_SRC_TAG) \
	 -e NOTARYVERSION=$(NOTARYVERSION) -e NOTARYMIGRATEVERSION=$(NOTARYMIGRATEVERSION) \
	 -e TRIVYVERSION=$(TRIVYVERSION) -e TRIVYADAPTERVERSION=$(TRIVYADAPTERVERSION) \
	 -e VERSIONTAG=$(VERSIONTAG) \
	 -e BUILDBIN=$(BUILDBIN) \
	 -e CHARTMUSEUMVERSION=$(CHARTMUSEUMVERSION) -e CHARTMUSEUM_SRC_TAG=$(CHARTMUSEUM_SRC_TAG) -e DOCKERIMAGENAME_CHART_SERVER=$(DOCKERIMAGENAME_CHART_SERVER) \
	 -e NPM_REGISTRY=$(NPM_REGISTRY) -e BASEIMAGETAG=$(BASEIMAGETAG) -e BASEIMAGENAMESPACE=$(BASEIMAGENAMESPACE) \
	 -e CHARTURL=$(CHARTURL) -e NOTARYURL=$(NOTARYURL) -e REGISTRYURL=$(REGISTRYURL) \
	 -e TRIVY_DOWNLOAD_URL=$(TRIVY_DOWNLOAD_URL) -e TRIVY_ADAPTER_DOWNLOAD_URL=$(TRIVY_ADAPTER_DOWNLOAD_URL) \
	 -e PULL_BASE_FROM_DOCKERHUB=$(PULL_BASE_FROM_DOCKERHUB) -e BUILD_BASE=$(BUILD_BASE) \
	 -e REGISTRYUSER=$(REGISTRYUSER) -e REGISTRYPASSWORD=$(REGISTRYPASSWORD) \
	 -e PUSHBASEIMAGE=$(PUSHBASEIMAGE) -e BUILD_PG96=$(BUILD_PG96)

build_standalone_db_migrator: compile_standalone_db_migrator
	make -f $(MAKEFILEPATH_PHOTON)/Makefile _build_standalone_db_migrator -e BASEIMAGETAG=$(BASEIMAGETAG) -e VERSIONTAG=$(VERSIONTAG)

build_base_docker:
	if [ -n "$(REGISTRYUSER)" ] && [ -n "$(REGISTRYPASSWORD)" ] ; then \
		docker login -u $(REGISTRYUSER) -p $(REGISTRYPASSWORD) ; \
	else \
		echo "No docker credentials provided, please make sure enough priviledges to access docker hub!" ; \
	fi
	@for name in $(BUILDBASETARGET); do \
		echo $$name ; \
		sleep 30 ; \
		if [ $$name == "db" ]; then \
		    make _build_base_db ; \
		else \
			$(DOCKERBUILD) --build-arg BUILD_PG96=$(BUILD_PG96) --pull --no-cache -f $(MAKEFILEPATH_PHOTON)/$$name/Dockerfile.base -t $(BASEIMAGENAMESPACE)/harbor-$$name-base:$(BASEIMAGETAG) --label base-build-date=$(date +"%Y%m%d") . ; \
		fi ; \
		if [ "$(PUSHBASEIMAGE)" != "false" ] ; then \
			$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(BASEIMAGENAMESPACE)/harbor-$$name-base:$(BASEIMAGETAG) $(REGISTRYUSER) $(REGISTRYPASSWORD) || exit 1; \
		fi ; \
	done

_build_base_db:
	@if [ "$(BUILD_PG96)" = "true" ] ; then \
		echo "build pg96 rpm package." ; \
		cd $(MAKEFILEPATH_PHOTON)/db && $(MAKEFILEPATH_PHOTON)/db/rpm_builder.sh && cd - ; \
		$(DOCKERBUILD) --pull --no-cache -f $(MAKEFILEPATH_PHOTON)/db/Dockerfile.pg96 -t $(BASEIMAGENAMESPACE)/harbor-db-base:$(BASEIMAGETAG) --label base-build-date=$(date +"%Y%m%d") . ; \
  	else \
 		$(DOCKERBUILD) --pull --no-cache -f $(MAKEFILEPATH_PHOTON)/db/Dockerfile.base -t $(BASEIMAGENAMESPACE)/harbor-db-base:$(BASEIMAGETAG) --label base-build-date=$(date +"%Y%m%d") . ; \
  	fi

pull_base_docker:
	@for name in $(BUILDBASETARGET); do \
		echo $$name ; \
		$(DOCKERPULL) $(BASEIMAGENAMESPACE)/harbor-$$name-base:$(BASEIMAGETAG) ; \
	done

install: compile build prepare start

package_online: update_prepare_version
	@echo "packing online package ..."
	@cp -r make $(HARBORPKG)
	@if [ -n "$(REGISTRYSERVER)" ] ; then \
		$(SEDCMDI) -e 's/image\: $(BASEIMAGENAMESPACE)/image\: $(REGISTRYSERVER)\/$(REGISTRYPROJECTNAME)/' \
		$(HARBORPKG)/docker-compose.yml ; \
	fi
	@cp LICENSE $(HARBORPKG)/LICENSE

	@$(TARCMD) $(PACKAGE_ONLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."

package_offline: update_prepare_version compile build

	@echo "packing offline package ..."
	@cp -r make $(HARBORPKG)
	@cp LICENSE $(HARBORPKG)/LICENSE

	@echo "saving harbor docker image"
	@$(DOCKERSAVE) $(DOCKERSAVE_PARA) > $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar
	@gzip $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar

	@$(TARCMD) $(PACKAGE_OFFLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."

gosec:
	#go get github.com/securego/gosec/cmd/gosec
	#go get github.com/dghubble/sling
	@echo "run secure go scan ..."
	@if [ "$(GOSECRESULTS)" != "" ] ; then \
		$(GOPATH)/bin/gosec -fmt=json -out=$(GOSECRESULTS) -quiet ./... | true ; \
	else \
		$(GOPATH)/bin/gosec -fmt=json -out=harbor_gas_output.json -quiet ./... | true ; \
	fi

go_check: gen_apis mocks_check misspell gofmt commentfmt golint govet

gofmt:
	@echo checking gofmt...
	@res=$$(gofmt -d -e -s $$(find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print)); \
	if [ -n "$${res}" ]; then \
		echo checking gofmt fail... ; \
		echo "$${res}"; \
		exit 1; \
	fi

commentfmt:
	@echo checking comment format...
	@res=$$(find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print | xargs egrep '(^|\s)\/\/(\S)'|grep -v '//go:generate'); \
	if [ -n "$${res}" ]; then \
		echo checking comment format fail.. ; \
		echo missing whitespace between // and comment body;\
		echo "$${res}"; \
		exit 1; \
	fi

misspell:
	@echo checking misspell...
	@find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print | xargs misspell -error

golint:
	@echo checking golint...
	@go list ./... | grep -v -E 'vendor|test' | xargs fgt golint

govet:
	@echo checking govet...
	@cd src;go list ./... | grep -v -E 'vendor|test' | xargs go vet

pushimage:
	@echo "pushing harbor images ..."
	@$(DOCKERTAG) $(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKER_IMAGE_NAME_PREPARE):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_CORE):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_DB):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG)

start:
	@echo "loading harbor images..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_FILE_OPT) up -d
	@echo "Start complete. You can visit harbor now."

down:
	@while [ -z "$$CONTINUE" ]; do \
        read -r -p "Type anything but Y or y to exit. [Y/N]: " CONTINUE; \
    done ; \
    [ $$CONTINUE = "y" ] || [ $$CONTINUE = "Y" ] || (echo "Exiting."; exit 1;)
	@echo "stoping harbor instance..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_FILE_OPT) down -v
	@echo "Done."

restart: down prepare start

swagger_client:
	@echo "Generate swagger client"
	wget https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/4.3.1/openapi-generator-cli-4.3.1.jar -O openapi-generator-cli.jar
	rm -rf harborclient
	mkdir  -p harborclient/harbor_client
	mkdir  -p harborclient/harbor_swagger_client
	mkdir  -p harborclient/harbor_v2_swagger_client
	java -jar openapi-generator-cli.jar generate -i api/swagger.yaml -g python -o harborclient/harbor_client --package-name client
	java -jar openapi-generator-cli.jar generate -i api/v2.0/legacy_swagger.yaml -g python -o harborclient/harbor_swagger_client --package-name swagger_client
	java -jar openapi-generator-cli.jar generate -i api/v2.0/swagger.yaml -g python -o harborclient/harbor_v2_swagger_client --package-name v2_swagger_client
	cd harborclient/harbor_client; python ./setup.py install
	cd harborclient/harbor_swagger_client; python ./setup.py install
	cd harborclient/harbor_v2_swagger_client; python ./setup.py install
	pip install docker -q
	pip freeze

cleanbinary:
	@echo "cleaning binary..."
	if [ -f $(CORE_BINARYPATH)/$(CORE_BINARYNAME) ] ; then rm $(CORE_BINARYPATH)/$(CORE_BINARYNAME) ; fi
	if [ -f $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ] ; then rm $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ; fi
	if [ -f $(REGISTRYCTLBINARYPATH)/$(REGISTRYCTLBINARYNAME) ] ; then rm $(REGISTRYCTLBINARYPATH)/$(REGISTRYCTLBINARYNAME) ; fi
	if [ -f $(MIGRATEPATCHBINARYPATH)/$(MIGRATEPATCHBINARYNAME) ] ; then rm $(MIGRATEPATCHBINARYPATH)/$(MIGRATEPATCHBINARYNAME) ; fi
	rm -rf make/photon/*/binary/

cleanbaseimage:
	@echo "cleaning base image for photon..."
	@for name in $(BUILDBASETARGET); do \
		$(DOCKERRMIMAGE) -f $(BASEIMAGENAMESPACE)/harbor-$$name-base:$(BASEIMAGETAG) ; \
	done

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

cleandockercomposefile:
	@echo "cleaning docker-compose files in $(DOCKERCOMPOSEFILEPATH)"
	@find $(DOCKERCOMPOSEFILEPATH) -maxdepth 1 -name "docker-compose*.yml" -exec rm -f {} \;
	@find $(DOCKERCOMPOSEFILEPATH) -maxdepth 1 -name "docker-compose*.yml-e" -exec rm -f {} \;

cleanpackage:
	@echo "cleaning harbor install package"
	@if [ -d $(BUILDPATH)/harbor ] ; then rm -rf $(BUILDPATH)/harbor ; fi
	@if [ -f $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ; fi
	@if [ -f $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ; fi

cleanconfig:
	@echo "clean generated config files"
	rm -f $(BUILDPATH)/make/photon/prepare/versions
	rm -f $(BUILDPATH)/UIVERSION
	rm -rf $(BUILDPATH)/make/common
	rm -rf $(BUILDPATH)/harborclient
	rm -rf $(BUILDPATH)/src/portal/dist
	rm -rf $(BUILDPATH)/src/portal/lib/dist
	rm -f $(BUILDPATH)/src/portal/proxy.config.json

.PHONY: cleanall
cleanall: cleanbinary cleanimage cleanbaseimage cleandockercomposefile cleanconfig cleanpackage

clean:
	@echo "  make cleanall:		remove binary, Harbor images, specific version docker-compose"
	@echo "		file, specific version tag, online and offline install package"
	@echo "  make cleanbinary:		remove core and jobservice binary"
	@echo "  make cleanbaseimage:		remove base image of Harbor images"
	@echo "  make cleanimage:		remove Harbor images"
	@echo "  make cleandockercomposefile:	remove specific version docker-compose"
	@echo "  make cleanpackage:		remove online and offline install package"

all: install
