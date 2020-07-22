#!/bin/bash
# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
set -x
gsutil version -l
set +x

docker-compose version

## -------------------------------------------- Pre-condition --------------------------------------------
if [[ $DRONE_REPO != "goharbor/harbor" ]]; then
    echo "Only run tests again Harbor Repo."
    exit 1
fi
# It won't package an new harbor build against tag, just pick up a build which passed CI and push to release.
if [[ $DRONE_BUILD_EVENT == "tag" || $DRONE_BUILD_EVENT == "pull_request" ]]; then
    echo "We do nothing against 'tag' and 'pull request'."
    exit 0
fi

## --------------------------------------------- Init Env -------------------------------------------------
dpkg -l > package.list
# Start Xvfb for Chrome headlesss
Xvfb -ac :99 -screen 0 1280x1024x16 & export DISPLAY=:99

export DRONE_SERVER=$DRONE_SERVER
export DRONE_TOKEN=$DRONE_TOKEN

upload_build=false
nightly_run=false
upload_latest_build=false
upload_bundle_success=false
latest_build_file='latest.build'
publish_npm=true

harbor_offline_build_bundle=""
harbor_online_build_bundle=""
harbor_logs_bucket="harbor-ci-logs"
harbor_builds_bucket="harbor-builds"
harbor_releases_bucket="harbor-releases"
harbor_ci_pipeline_store_bucket="harbor-ci-pipeline-store/latest"
harbor_target_bucket=""
if [[ $DRONE_BRANCH == "master" ]]; then
  harbor_target_bucket=$harbor_builds_bucket
else
  harbor_target_bucket=$harbor_releases_bucket/$DRONE_BRANCH
fi

# GC credentials
keyfile="/root/harbor-ci-logs.key"
botofile="/root/.boto"
echo -en $GS_PRIVATE_KEY > $keyfile
chmod 400 $keyfile
echo "[Credentials]" >> $botofile
echo "gs_service_key_file = $keyfile" >> $botofile
echo "gs_service_client_id = $GS_CLIENT_EMAIL" >> $botofile
echo "[GSUtil]" >> $botofile
echo "content_language = en" >> $botofile
echo "default_project_id = $GS_PROJECT_ID" >> $botofile
container_ip=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
echo $container_ip

## --------------------------------------------- Init Version -----------------------------------------------
buildinfo=$(drone build info goharbor/harbor $DRONE_BUILD_NUMBER)
echo $buildinfo
#  the target release version is the version of next release(RC or GA). It needs to be updated on creating new release branch.
target_release_version=$(cat ./VERSION)
#  the harbor package version is for both online and offline installer.
#  harbor-offline-installer-v1.5.2-build.8.tgz
Harbor_Package_Version=$target_release_version-'build.'$DRONE_BUILD_NUMBER
#  the harbor assets version is for tag of harbor images:
# 1, On master branch, it's same as package version.
# 2, On release branch(others), it would set to the target realese version so that we can rename the latest passed CI build to publish.
if [[ $DRONE_BRANCH == "master" ]]; then
  Harbor_Assets_Version=$Harbor_Package_Version
else
  Harbor_Assets_Version=$target_release_version
fi
export Harbor_Assets_Version=$Harbor_Assets_Version
#  the env is for online and offline package.
export Harbor_Package_Version=$Harbor_Package_Version
export NPM_REGISTRY=$NPM_REGISTRY

# release branch must have their own base image with branch name, master and others will use the dev as base.
if [[ $DRONE_BRANCH == "release-"* ]]; then
  Harbor_Build_Base_Tag=$target_release_version
else
  Harbor_Build_Base_Tag=dev
fi
export Harbor_Build_Base_Tag=$Harbor_Build_Base_Tag

echo "--------------------------------------------------"
echo "Harbor Package version: $Harbor_Package_Version"
echo "Harbor Assets version: $Harbor_Assets_Version"
echo "Harbor Build Base tag: $Harbor_Build_Base_Tag"
echo "--------------------------------------------------"

# GS util
function uploader {
    gsutil cp $1 gs://$2/$1
    gsutil -D setacl public-read gs://$2/$1 &> /dev/null
}

function package_installer {
    echo "Package Harbor offline installer."
    pybot --removekeywords TAG:secret --include Bundle tests/robot-cases/Group0-Distro-Harbor
    harbor_offline_build_bundle=$(basename harbor-offline-installer-*.tgz)
    harbor_online_build_bundle=$(basename harbor-online-installer-*.tgz)
    upload_build=true
    echo "Package name is: $harbor_offline_build_bundle"
    du -ks $harbor_offline_build_bundle | awk '{print $1 / 1024}' | { read x; echo $x MB; }
}

# publish images to Docker Hub
function publishImage {
    echo "Publishing images to Docker Hub..."
    echo "The images on the host:"
    # for master, will use 'dev' as the tag name
    # for release-*, will use 'release-*-dev' as the tag name, like release-v1.8.0-dev
    if [[ $DRONE_BRANCH == "master" ]]; then
      image_tag=dev
    fi
    if [[ $DRONE_BRANCH == "release-"* ]]; then
      image_tag=$Harbor_Assets_Version-dev
    fi
    # rename the images with tag "dev" and push to Docker Hub
    docker images
    docker login -u $DOCKER_HUB_USERNAME -p $DOCKER_HUB_PASSWORD
    docker images | grep goharbor | grep -v "\-base" | sed -n "s|\(goharbor/[-._a-z0-9]*\)\s*\(.*$Harbor_Assets_Version\).*|docker tag \1:\2 \1:$image_tag;docker push \1:$image_tag|p" | bash
    echo "Images are published successfully"
    docker images
}

echo "--------------------------------------------------"
echo "Running CI for $DRONE_BUILD_EVENT on $DRONE_BRANCH"
echo "--------------------------------------------------"

##
# Any merge code(PUSH) on branch master, release-* will trigger package offline installer.
#
# Put code here is because that it needs clean code to build installer.
##
if [[ $DRONE_BRANCH == "master" || $DRONE_BRANCH == *"refs/tags"* || $DRONE_BRANCH == "release-"* ]]; then
    if [[ $DRONE_BUILD_EVENT == "push" ]]; then
        package_installer
        upload_latest_build=true
        echo -en "$HARBOR_SIGN_KEY" | gpg --import
        gpg -v -ab -u $HARBOR_SIGN_KEY_ID $harbor_offline_build_bundle
        gpg -v -ab -u $HARBOR_SIGN_KEY_ID $harbor_online_build_bundle
    fi
fi

## --------------------------------------------- Upload Harbor Bundle File ---------------------------------------
#
# Build storage structure:
#
# 1(master), harbor-builds/harbor-offline-installer-*.tgz
#                         latest.build
#                         harbor-offline-installer-latest.tgz

# 2(others), harbor-releases/${branch}/harbor-offline-installer-*.tgz
#                                     latest.build
#                                     harbor-offline-installer-latest.tgz
#
set -e
if [ $upload_build == true ]; then
    cp ${harbor_offline_build_bundle}     harbor-offline-installer-latest.tgz
    cp ${harbor_offline_build_bundle}.asc harbor-offline-installer-latest.tgz.asc
    uploader ${harbor_offline_build_bundle}     $harbor_target_bucket
    uploader ${harbor_offline_build_bundle}.asc $harbor_target_bucket
    uploader ${harbor_online_build_bundle}     $harbor_target_bucket
    uploader ${harbor_online_build_bundle}.asc $harbor_target_bucket
    uploader harbor-offline-installer-latest.tgz     $harbor_target_bucket
    uploader harbor-offline-installer-latest.tgz.asc $harbor_target_bucket
    upload_bundle_success=true
fi


## --------------------------------------------- Upload Harbor Dev Images ---------------------------------------
#
# Any merge code(PUSH) on branch master, release-* will trigger push dev images.
#
##
if [[ $DRONE_BRANCH == "master" || $DRONE_BRANCH == "release-"* ]]; then
    if [[ $DRONE_BUILD_EVENT == "push" ]]; then
        publishImage
    fi
fi

## --------------------------------------------- Upload Harbor Latest Build File ----------------------------------
#
# latest.build file holds the latest offline installer url, it must be sure that the installer has been uploaded successfull.
#
if [ $upload_latest_build == true ] && [ $upload_bundle_success == true ]; then
    echo 'https://storage.googleapis.com/'$harbor_target_bucket/$harbor_offline_build_bundle > $latest_build_file
    uploader $latest_build_file $harbor_target_bucket
fi

## --------------------------------------------- Upload securego results ------------------------------------------
#if [ $DRONE_BUILD_EVENT == "push" ]; then
#    go get github.com/securego/gosec/cmd/gosec
#    go get github.com/dghubble/sling
#    make gosec -e GOSECRESULTS=harbor-gosec-results-latest.json
#    echo $git_commit > ./harbor-gosec-results-latest-version
#    uploader harbor-gosec-results-latest.json $harbor_target_bucket
#    uploader harbor-gosec-results-latest-version $harbor_target_bucket
#fi

## ------------------------------------------------ Tear Down -----------------------------------------------------
if [ -f "$keyfile" ]; then
  rm -f $keyfile
fi

