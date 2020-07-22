#!/bin/bash
set -x

set +e
sudo rm -fr /data/*
sudo mkdir -p /data
DIR="$(cd "$(dirname "$0")" && pwd)"

set -e
if [ -z "$1" ]; then echo no ip specified; exit 1;fi
# prepare cert ...
sudo ./tests/generateCerts.sh $1
sudo mkdir -p /etc/docker/certs.d/$1 && sudo cp ./tests/harbor_ca.crt /etc/docker/certs.d/$1/ && rm -rf ~/.docker/ &&  mkdir -p ~/.docker/tls/$1:4443/ && sudo cp ./tests/harbor_ca.crt ~/.docker/tls/$1:4443/

sudo ./tests/hostcfg.sh

if [ "$2" = 'LDAP' ]; then
    cd tests && sudo ./ldapprepare.sh && cd ..
fi



# prepare a chart file for API_DB test...
sudo curl -o $DIR/../../tests/apitests/python/mariadb-4.3.1.tgz https://storage.googleapis.com/harbor-builds/bin/charts/mariadb-4.3.1.tgz

sudo apt-get update && sudo apt-get install -y --no-install-recommends python-dev openjdk-7-jdk libssl-dev && sudo apt-get autoremove -y && sudo rm -rf /var/lib/apt/lists/*
sudo wget https://bootstrap.pypa.io/get-pip.py && sudo python ./get-pip.py && sudo pip install --ignore-installed urllib3 chardet requests && sudo pip install robotframework==3.0.4 robotframework-httplibrary requests dbbot robotframework-pabot --upgrade
sudo make swagger_client
if [ $GITHUB_TOKEN ];
then
    sed "s/# github_token: xxx/github_token: $GITHUB_TOKEN/" -i make/harbor.yml
fi
sudo make compile build prepare COMPILETAG=compile_golangimage GOBUILDTAGS="include_oss include_gcs" NOTARYFLAG=true CLAIRFLAG=true TRIVYFLAG=true CHARTFLAG=true GEN_TLS=true

# set the debugging env
echo "GC_TIME_WINDOW_HOURS=0" | sudo tee -a ./make/common/config/core/env
sudo make start

# waiting 5 minutes to start
for((i=1;i<=30;i++)); do
  echo $i waiting 10 seconds...
  sleep 10
  curl -k -L -f 127.0.0.1/api/v2.0/systeminfo && break
  docker ps
done
