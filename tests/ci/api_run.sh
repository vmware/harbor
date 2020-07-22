#!/bin/bash
set -x

#source gskey.sh

sudo gsutil version -l

harbor_logs_bucket="harbor-ci-logs"
# GC credentials
#keyfile="/home/travis/harbor-ci-logs.key"
#botofile="/home/travis/.boto"
#echo -en $GS_PRIVATE_KEY > $keyfile
#sudo chmod 400 $keyfile
#echo "[Credentials]" >> $botofile
#echo "gs_service_key_file = $keyfile" >> $botofile
#echo "gs_service_client_id = $GS_CLIENT_EMAIL" >> $botofile
#echo "[GSUtil]" >> $botofile
#echo "content_language = en" >> $botofile
#echo "default_project_id = $GS_PROJECT_ID" >> $botofile
DIR="$(cd "$(dirname "$0")" && pwd)"

# GS util
function uploader {
   sudo gsutil cp $1 gs://$2/$1
   sudo gsutil acl ch -u AllUsers:R gs://$2/$1
}

set +e

docker ps
# run db auth api cases
if [ "$1" = 'DB' ]; then
    pybot -X -v ip:$2 -v HARBOR_PASSWORD:Harbor12345 $DIR/../../tests/robot-cases/Group0-BAT/API_DB.robot
elif [ "$1" = 'LDAP' ]; then
    # run ldap api cases
    python $DIR/../../tests/configharbor.py -H $IP -u $HARBOR_ADMIN -p $HARBOR_ADMIN_PASSWD -c auth_mode=ldap_auth \
                                  ldap_url=ldap://$IP \
                                  ldap_search_dn=cn=admin,dc=example,dc=com \
                                  ldap_search_password=admin \
                                  ldap_base_dn=dc=example,dc=com \
                                  ldap_uid=cn
    pybot -X -v ip:$2 -v HARBOR_PASSWORD:Harbor12345 $DIR/../../tests/robot-cases/Group0-BAT/API_LDAP.robot
else
    rc=999
fi
rc=$?
## --------------------------------------------- Upload Harbor CI Logs -------------------------------------------
timestamp=$(date +%s)
outfile="integration_logs_$timestamp$TRAVIS_COMMIT.tar.gz"
sudo tar -zcvf $outfile output.xml log.html /var/log/harbor/*
if [ -f "$outfile" ]; then
   uploader $outfile $harbor_logs_bucket
   echo "----------------------------------------------"
   echo "Download test logs:"
   echo "https://storage.googleapis.com/harbor-ci-logs/$outfile"
   echo "----------------------------------------------"
else
   echo "No log output file to upload"
fi

exit $rc
