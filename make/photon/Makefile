# Makefile for a harbor project
#
# Targets:
#
# build: 	build harbor photon images
# clean:	clean core and jobservice harbor images

# common
SHELL := /bin/bash
BUILDPATH=$(CURDIR)
MAKEPATH=$(BUILDPATH)/make
SRCPATH=./src
TOOLSPATH=$(CURDIR)/tools
SEDCMD=$(shell which sed)
WGET=$(shell which wget)
CURL=$(shell which curl)
TIMESTAMP=$(shell date +"%Y%m%d")

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build --no-cache
DOCKERBUILD_WITH_PULL_PARA=$(DOCKERBUILD) --pull=$(PULL_BASE_FROM_DOCKERHUB)
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERIMAGES=$(DOCKERCMD) images
BASEIMAGENAMESPACE=goharbor

# pushimage
PUSHSCRIPTPATH=$(MAKEPATH)
PUSHSCRIPTNAME=pushimage.sh

# binary
CORE_SOURCECODE=$(SRCPATH)/core
CORE_BINARYNAME=harbor_core
JOBSERVICESOURCECODE=$(SRCPATH)/jobservice
JOBSERVICEBINARYNAME=harbor_jobservice

# photon dockerfile
DOCKERFILEPATH=$(MAKEPATH)/photon

PREPARE=prepare
DOCKERFILEPATH_PREPARE=$(DOCKERFILEPATH)/$(PREPARE)
DOCKERFILENAME_PREPARE=Dockerfile
DOCKERIMAGENAME_PREPARE=$(BASEIMAGENAMESPACE)/$(PREPARE)

PORTAL=portal
DOCKERFILEPATH_PORTAL=$(DOCKERFILEPATH)/$(PORTAL)
DOCKERFILENAME_PORTAL=Dockerfile
DOCKERIMAGENAME_PORTAL=$(BASEIMAGENAMESPACE)/harbor-$(PORTAL)

CORE=core
DOCKERFILEPATH_CORE=$(DOCKERFILEPATH)/$(CORE)
DOCKERFILENAME_CORE=Dockerfile
DOCKERIMAGENAME_CORE=$(BASEIMAGENAMESPACE)/harbor-$(CORE)

JOBSERVICE=jobservice
DOCKERFILEPATH_JOBSERVICE=$(DOCKERFILEPATH)/$(JOBSERVICE)
DOCKERFILENAME_JOBSERVICE=Dockerfile
DOCKERIMAGENAME_JOBSERVICE=$(BASEIMAGENAMESPACE)/harbor-$(JOBSERVICE)

LOG=log
DOCKERFILEPATH_LOG=$(DOCKERFILEPATH)/$(LOG)
DOCKERFILENAME_LOG=Dockerfile
DOCKERIMAGENAME_LOG=$(BASEIMAGENAMESPACE)/harbor-$(LOG)

DB=db
DOCKERFILEPATH_DB=$(DOCKERFILEPATH)/$(DB)
DOCKERFILENAME_DB=Dockerfile
DOCKERIMAGENAME_DB=$(BASEIMAGENAMESPACE)/harbor-$(DB)

POSTGRESQL=postgresql
DOCKERFILEPATH_POSTGRESQL=$(DOCKERFILEPATH)/$(POSTGRESQL)
DOCKERFILENAME_POSTGRESQL=Dockerfile
DOCKERIMAGENAME_POSTGRESQL=$(BASEIMAGENAMESPACE)/$(POSTGRESQL)-photon

TRIVY_ADAPTER=trivy-adapter
DOCKERFILEPATH_TRIVY_ADAPTER=$(DOCKERFILEPATH)/$(TRIVY_ADAPTER)
DOCKERFILENAME_TRIVY_ADAPTER=Dockerfile
DOCKERIMAGENAME_TRIVY_ADAPTER=$(BASEIMAGENAMESPACE)/$(TRIVY_ADAPTER)-photon

NGINX=nginx
DOCKERFILEPATH_NGINX=$(DOCKERFILEPATH)/$(NGINX)
DOCKERFILENAME_NGINX=Dockerfile
DOCKERIMAGENAME_NGINX=$(BASEIMAGENAMESPACE)/$(NGINX)-photon

REGISTRY=registry
DOCKERFILEPATH_REG=$(DOCKERFILEPATH)/$(REGISTRY)
DOCKERFILENAME_REG=Dockerfile
DOCKERIMAGENAME_REG=$(BASEIMAGENAMESPACE)/$(REGISTRY)-photon

REGISTRYCTL=registryctl
DOCKERFILEPATH_REGISTRYCTL=$(DOCKERFILEPATH)/$(REGISTRYCTL)
DOCKERFILENAME_REGISTRYCTL=Dockerfile
DOCKERIMAGENAME_REGISTRYCTL=$(BASEIMAGENAMESPACE)/harbor-$(REGISTRYCTL)

NOTARYSERVER=notary-server
NOTARYSIGNER=notary-signer
DOCKERFILEPATH_NOTARY=$(DOCKERFILEPATH)/notary
DOCKERFILEPATH_NOTARYSERVER=$(DOCKERFILEPATH)/$(NOTARYSERVER)
DOCKERFILENAME_NOTARYSERVER=Dockerfile
DOCKERIMAGENAME_NOTARYSERVER=$(BASEIMAGENAMESPACE)/$(NOTARYSERVER)-photon
DOCKERFILEPATH_NOTARYSIGNER=$(DOCKERFILEPATH)/$(NOTARYSIGNER)
DOCKERFILENAME_NOTARYSIGNER=Dockerfile
DOCKERIMAGENAME_NOTARYSIGNER=$(BASEIMAGENAMESPACE)/$(NOTARYSIGNER)-photon

REDIS=redis
DOCKERFILEPATH_REDIS=$(DOCKERFILEPATH)/$(REDIS)
DOCKERFILENAME_REDIS=Dockerfile
DOCKERIMAGENAME_REDIS=$(BASEIMAGENAMESPACE)/$(REDIS)-photon

DOCKERFILEPATH_STANDALONE_DB_MIGRATOR=$(DOCKERFILEPATH)/standalone-db-migrator
DOCKERFILENAME_STANDALONE_DB_MIGRATOR=Dockerfile
DOCKERIMAGENAME_STANDALONE_DB_MIGRATOR=$(BASEIMAGENAMESPACE)/standalone-db-migrator

EXPORTER=exporter
DOCKERFILEPATH_EXPORTER=$(DOCKERFILEPATH)/$(EXPORTER)
DOCKERFILENAME_EXPORTER=Dockerfile
DOCKERIMAGENAME_EXPORTER=$(BASEIMAGENAMESPACE)/harbor-$(EXPORTER)

# for chart server (chartmuseum)
CHARTSERVER=chartserver
DOCKERFILEPATH_CHART_SERVER=$(DOCKERFILEPATH)/$(CHARTSERVER)
DOCKERFILENAME_CHART_SERVER=Dockerfile
CHART_SERVER_CODE_BASE=https://github.com/helm/chartmuseum.git
CHART_SERVER_MAIN_PATH=cmd/chartmuseum
CHART_SERVER_BIN_NAME=chartm

_build_prepare:
	@$(call _build_base,$(PREPARE),$(DOCKERFILEPATH_PREPARE))
	@echo "building prepare container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_PREPARE)/$(DOCKERFILENAME_PREPARE) -t $(DOCKERIMAGENAME_PREPARE):$(VERSIONTAG) .
	@echo "Done."

_build_db:
	@$(call _build_base,$(DB),$(DOCKERFILEPATH_DB))
	@echo "building db container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_DB)/$(DOCKERFILENAME_DB) -t $(DOCKERIMAGENAME_DB):$(VERSIONTAG) .
	@echo "Done."

_build_portal:
	@$(call _build_base,$(PORTAL),$(DOCKERFILEPATH_PORTAL))
	@echo "building portal container for photon..."
	$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) --build-arg npm_registry=$(NPM_REGISTRY) -f $(DOCKERFILEPATH_PORTAL)/$(DOCKERFILENAME_PORTAL) -t $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) .
	@echo "Done."

_build_core:
	@$(call _build_base,$(CORE),$(DOCKERFILEPATH_CORE))
	@echo "building core container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_CORE)/$(DOCKERFILENAME_CORE) -t $(DOCKERIMAGENAME_CORE):$(VERSIONTAG) .
	@echo "Done."

_build_jobservice:
	@$(call _build_base,$(JOBSERVICE),$(DOCKERFILEPATH_JOBSERVICE))
	@echo "building jobservice container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_JOBSERVICE)/$(DOCKERFILENAME_JOBSERVICE) -t $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) .
	@echo "Done."

_build_log:
	@$(call _build_base,$(LOG),$(DOCKERFILEPATH_LOG))
	@echo "building log container for photon..."
	$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_LOG)/$(DOCKERFILENAME_LOG) -t $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) .
	@echo "Done."

_build_trivy_adapter:
	@if [ "$(TRIVYFLAG)" = "true" ] ; then \
		$(call _build_base,$(TRIVY_ADAPTER),$(DOCKERFILEPATH_TRIVY_ADAPTER)) ; \
		rm -rf $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary && mkdir -p $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary ; \
		echo "Downloading Trivy scanner $(TRIVYVERSION)..." ; \
		$(call _extract_archive, $(TRIVY_DOWNLOAD_URL), $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary/) ; \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			echo "Downloading Trivy adapter $(TRIVYADAPTERVERSION)..." ; \
			$(call _extract_archive, $(TRIVY_ADAPTER_DOWNLOAD_URL), $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary/) ; \
		else \
			echo "Building Trivy adapter $(TRIVYADAPTERVERSION) from sources..." ; \
			cd $(DOCKERFILEPATH_TRIVY_ADAPTER) && $(DOCKERFILEPATH_TRIVY_ADAPTER)/builder.sh $(TRIVYADAPTERVERSION) && cd - ; \
		fi ; \
		echo "Building Trivy adapter container for photon..." ; \
		$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) \
			--build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) \
			--build-arg trivy_version=$(TRIVYVERSION) \
			-f $(DOCKERFILEPATH_TRIVY_ADAPTER)/$(DOCKERFILENAME_TRIVY_ADAPTER) \
			-t $(DOCKERIMAGENAME_TRIVY_ADAPTER):$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_TRIVY_ADAPTER)/binary; \
		echo "Done." ; \
	fi

_build_chart_server:
	@if [ "$(CHARTFLAG)" = "true" ] ; then \
		$(call _build_base,$(CHARTSERVER),$(DOCKERFILEPATH_CHART_SERVER)); \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_CHART_SERVER)/binary && mkdir -p $(DOCKERFILEPATH_CHART_SERVER)/binary && \
			$(call _get_binary, $(CHARTURL), $(DOCKERFILEPATH_CHART_SERVER)/binary/chartm); \
		else \
			cd $(DOCKERFILEPATH_CHART_SERVER) && $(DOCKERFILEPATH_CHART_SERVER)/builder $(GOBUILDIMAGE) $(CHART_SERVER_CODE_BASE) $(CHARTMUSEUM_SRC_TAG) $(CHART_SERVER_MAIN_PATH) $(CHART_SERVER_BIN_NAME) && cd - ; \
		fi ; \
		echo "building chartmuseum container for photon..." ; \
		$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_CHART_SERVER)/$(DOCKERFILENAME_CHART_SERVER) -t $(DOCKERIMAGENAME_CHART_SERVER):$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_CHART_SERVER)/binary; \
		echo "Done." ; \
	fi

_build_nginx:
	@$(call _build_base,$(NGINX),$(DOCKERFILEPATH_NGINX))
	@echo "building nginx container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_NGINX)/$(DOCKERFILENAME_NGINX) -t $(DOCKERIMAGENAME_NGINX):$(VERSIONTAG) .
	@echo "Done."

_build_notary:
	@if [ "$(NOTARYFLAG)" = "true" ] ; then \
		$(call _build_base,$(NOTARYSERVER),$(DOCKERFILEPATH_NOTARYSERVER)) ; \
		$(call _build_base,$(NOTARYSIGNER),$(DOCKERFILEPATH_NOTARYSIGNER)) ; \
		if [ "$(BUILDBIN)" != "true" ] ; then \
			rm -rf $(DOCKERFILEPATH_NOTARY)/binary && mkdir -p $(DOCKERFILEPATH_NOTARY)/binary && \
			$(call _get_binary, $(NOTARYURL), $(DOCKERFILEPATH_NOTARY)/binary-bundle.tgz); \
			cd $(DOCKERFILEPATH_NOTARY) && tar -zvxf binary-bundle.tgz && cd - ; \
		else \
			cd $(DOCKERFILEPATH_NOTARY) && $(DOCKERFILEPATH_NOTARY)/builder $(NOTARYVERSION) $(NOTARYMIGRATEVERSION) && cd - ; \
		fi ; \
		echo "building notary container for photon..."; \
		chmod 655 $(DOCKERFILEPATH_NOTARY)/binary/notary-signer && $(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_NOTARYSIGNER)/$(DOCKERFILENAME_NOTARYSIGNER) -t $(DOCKERIMAGENAME_NOTARYSIGNER):$(VERSIONTAG) . ; \
		chmod 655 $(DOCKERFILEPATH_NOTARY)/binary/notary-server && $(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_NOTARYSERVER)/$(DOCKERFILENAME_NOTARYSERVER) -t $(DOCKERIMAGENAME_NOTARYSERVER):$(VERSIONTAG) . ; \
		rm -rf $(DOCKERFILEPATH_NOTARY)/binary; \
		echo "Done."; \
	fi

_build_registry:
	@$(call _build_base,$(REGISTRY),$(DOCKERFILEPATH_REG))
	@if [ "$(BUILDBIN)" != "true" ] ; then \
		rm -rf $(DOCKERFILEPATH_REG)/binary && mkdir -p $(DOCKERFILEPATH_REG)/binary && \
		$(call _get_binary, $(REGISTRYURL), $(DOCKERFILEPATH_REG)/binary/registry); \
	else \
		cd $(DOCKERFILEPATH_REG) && $(DOCKERFILEPATH_REG)/builder $(REGISTRY_SRC_TAG) && cd - ; \
	fi
	@echo "building registry container for photon..."
	@chmod 655 $(DOCKERFILEPATH_REG)/binary/registry && $(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REG)/$(DOCKERFILENAME_REG) -t $(DOCKERIMAGENAME_REG):$(VERSIONTAG) .
	@echo "Done."

_build_registryctl:
	@$(call _build_base,$(REGISTRYCTL),$(DOCKERFILEPATH_REGISTRYCTL))
	@echo "building registry controller for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REGISTRYCTL)/$(DOCKERFILENAME_REGISTRYCTL) -t $(DOCKERIMAGENAME_REGISTRYCTL):$(VERSIONTAG) .
	@rm -rf $(DOCKERFILEPATH_REG)/binary
	@echo "Done."

_build_redis:
	@$(call _build_base,$(REDIS),$(DOCKERFILEPATH_REDIS))
	@echo "building redis container for photon..."
	@$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_REDIS)/$(DOCKERFILENAME_REDIS) -t $(DOCKERIMAGENAME_REDIS):$(VERSIONTAG) .
	@echo "Done."

_build_standalone_db_migrator:
	@echo "building standalone db migrator image for photon..."
	$(DOCKERBUILD_WITH_PULL_PARA) --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) -f $(DOCKERFILEPATH_STANDALONE_DB_MIGRATOR)/$(DOCKERFILENAME_STANDALONE_DB_MIGRATOR) -t $(DOCKERIMAGENAME_STANDALONE_DB_MIGRATOR):$(VERSIONTAG) .
	@echo "Done."

_compile_and_build_exporter:
	@$(call _build_base,$(EXPORTER),$(DOCKERFILEPATH_EXPORTER))
	@echo "compiling and building image for exporter..."
	@$(DOCKERCMD) build --build-arg harbor_base_image_version=$(BASEIMAGETAG) --build-arg harbor_base_namespace=$(BASEIMAGENAMESPACE) --build-arg build_image=$(GOBUILDIMAGE) -f ${DOCKERFILEPATH_EXPORTER}/${DOCKERFILENAME_EXPORTER} -t $(DOCKERIMAGENAME_EXPORTER):$(VERSIONTAG) .
	@echo "Done."

define _extract_archive
	echo "download $1";\
	$(CURL) --connect-timeout 30 -f -k -L $1 | tar xvz -C $2 || exit 1
endef

define _get_binary
	echo "download $1";\
	$(CURL) --connect-timeout 30 -f -k -L $1 -o $2 || exit 1
endef

define _build_base
	if [ "$(BUILD_BASE)" = "true" ] ; then \
		echo "building base image for $(1)...";\
		if [ -n "$(REGISTRYUSER)"  ] && [ -n "$(REGISTRYPASSWORD)" ] ; then \
			docker login -u $(REGISTRYUSER) -p $(REGISTRYPASSWORD) ; \
		else \
			echo "No docker credentials provided, please be aware of priviledges to access docker hub!" ; \
		fi ;\
		if [ "$(1)" = "db" ] && [ "$(BUILD_PG96)" = "true" ]; then \
			echo "build pg96 rpm package." ; \
			cd $(DOCKERFILEPATH)/db && $(DOCKERFILEPATH)/db/rpm_builder.sh && cd - ; \
			$(DOCKERBUILD) --pull --no-cache -f $(DOCKERFILEPATH)/db/Dockerfile.pg96 -t $(BASEIMAGENAMESPACE)/harbor-db-base:$(BASEIMAGETAG) --label base-build-date=$(TIMESTAMP) . ; \
		else \
			$(DOCKERBUILD) -f $(2)/Dockerfile.base -t $(BASEIMAGENAMESPACE)/harbor-$(1)-base:$(BASEIMAGETAG) --label base-build-date=$(TIMESTAMP) . ;\
		fi ;\
		if [ "$(PUSHBASEIMAGE)" = "true" ] ; then \
			$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(BASEIMAGENAMESPACE)/harbor-$(1)-base:$(BASEIMAGETAG) $(REGISTRYUSER) $(REGISTRYPASSWORD) docker.io $(PULL_BASE_FROM_DOCKERHUB) || exit 1; \
		fi ; \
	fi
endef

build: _build_prepare _build_db _build_portal _build_core _build_jobservice _build_log _build_nginx _build_registry _build_registryctl _build_notary _build_trivy_adapter _build_redis _build_chart_server _compile_and_build_exporter
	@if [ -n "$(REGISTRYUSER)"  ] && [ -n "$(REGISTRYPASSWORD)" ] ; then \
		docker logout ; \
	fi
cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

.PHONY: clean
clean: cleanimage

