#!/bin/sh

set -e

if ! grep -q "Photon" /etc/lsb-release; then
    echo "Current OS is not Photon, skip appending ca bundle"
    exit 0
fi

if [ ! -f ~/ca-bundle.crt.original ]; then
    cp /etc/pki/tls/certs/ca-bundle.crt ~/ca-bundle.crt.original
fi

cp ~/ca-bundle.crt.original /etc/pki/tls/certs/ca-bundle.crt

# Install /etc/harbor/ssl/{component}/ca.crt to trust CA.
echo "Appending internal tls trust CA to ca-bundle ..."
for caFile in `find /etc/harbor/ssl -maxdepth 2 -name ca.crt`; do
    cat $caFile >> /etc/pki/tls/certs/ca-bundle.crt
    echo "Internal tls trust CA $caFile appended ..."
done
echo "Internal tls trust CA appending is Done."

if [[ -d /harbor_cust_cert && -n "$(ls -A /harbor_cust_cert)" ]]; then
    echo "Appending trust CA to ca-bundle ..."
    for z in /harbor_cust_cert/*; do
        case ${z} in
            *.crt | *.ca | *.ca-bundle | *.pem)
                if [ -d "$z" ]; then
                    echo "$z is dirictory, skip it ..."
                else
                    cat $z >> /etc/pki/tls/certs/ca-bundle.crt
                    echo " $z Appended ..."
                fi
                ;;
            *) echo "$z is Not ca file ..." ;;
        esac
    done
    echo "CA appending is Done."
fi
