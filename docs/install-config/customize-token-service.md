---
title: Customize the Harbor Token Service
weight: 60
---

By default, Harbor uses its own private key and certificate to authenticate with Docker clients. This topic describes how to optionally customize your configuration to use your own key and certificate.

Harbor requires the Docker client to access the Harbor registry with a token. The procedure to generate a token is like [Docker Registry v2 authentication](https://github.com/docker/distribution/blob/master/docs/spec/auth/token.md). Firstly, you make a request to the token service for a token. The token is signed by the private key. After that, you make a new request with the token to the Harbor registry, Harbor registry verifies the token with the public key in the root cert bundle. Then Harbor registry authorizes the Docker client to push and pull images.

- If you do not already have a certificate, follow the instructions in [Generate a Root Certificate](#gen-cert) to generate a root certificate by using openSSL.
- If you already have a certificate, go to [Provide the Certificate to Harbor](#provide-cert).

## Generate a Root Certificate {#gen-cert}
  
1. Generate a private key.

   ```sh
   openssl genrsa -out private_key.pem 4096    
   ```
   
1. Generate a certificate.  

   ```sh
   openssl req -new -x509 -key private_key.pem -out root.crt -days 3650
   ```   

1. Enter information to include in your certificate request.

   What you are about to enter is what is called a Distinguished Name or a DN. There are quite a few fields but you can leave some of them blank. For some fields there is a default value. If you enter `.`, the field is left blank.

   - Country Name (2 letter code) [AU]:
   - State or Province Name (full name) [Some-State]:
   - Locality Name (eg, city) []:
   - Organization Name (eg, company) [Internet Widgits Pty Ltd]:
   - Organizational Unit Name (eg, section) []:
   - Common Name (eg,  server FQDN or YOUR name) []:
   - Email Address []:

   After you run these commands, the files `private_key.pem` and `root.crt` are created in the current directory.

## Provide the Certificate to Harbor {#provide-cert}

See [Run the Installer Script](run-installer-script.md) or [Reconfigure Harbor and Manage the Harbor Lifecycle](reconfigure-manage-lifecycle.md) to install or reconfigure Harbor. After you run `./install` or `./prepare`, Harbor generates several configuration files. You need to replace the original private key and certificate with your own key and certificate.

1. Replace the default key and certificate. 

   Assuming that the key and certificate are in `/root/cert`, run the following commands:

   ```sh
   cd config/ui
   cp /root/cert/private_key.pem private_key.pem
   cp /root/cert/root.crt ../registry/root.crt
   ```

1. Go back to the `make` directory, and start Harbor by using following command:

   ```sh
   docker-compose up -d
   ```

1. Push and pull images to and from Harbor to check that your own certificate works. 
