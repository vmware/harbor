#!/bin/bash
IP=$(hostname -I | awk '{print $1}')

#echo $IP
sudo sed "s/reg.mydomain.com/$IP/" make/harbor.yml.tmpl |sudo tee make/harbor.yml

# enable internal tls
echo "internal_tls:" >> make/harbor.yml
echo "  enabled: true" >> make/harbor.yml
echo "  dir: /etc/harbor/tls/internal" >> make/harbor.yml

# TODO: remove it when scanner adapter support internal access of harbor
echo "storage_service:" >> make/harbor.yml
echo "  ca_bundle: /data/cert/server.crt" >> make/harbor.yml

sed "s|/your/certificate/path|/data/cert/server.crt|g" -i make/harbor.yml
sed "s|/your/private/key/path|/data/cert/server.key|g" -i make/harbor.yml
