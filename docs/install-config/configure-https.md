---
title: Configure HTTPS Access to Harbor
weight: 30
---

By default, Harbor does not ship with certificates. It is possible to deploy Harbor without security, so that you can connect to it over HTTP. However, using HTTP is acceptable only in air-gapped test or development environments that do not have a connection to the external internet. Using HTTP in environments that are not air-gapped exposes you to man-in-the-middle attacks. In production environments, always use HTTPS. If you enable Content Trust with Notary to properly sign all images, you must use HTTPS. 

To configure HTTPS, you must create SSL certificates. You can use certificates that are signed by a trusted third-party CA, or you can use self-signed certificates. This section describes how to use [OpenSSL](https://www.openssl.org/) to create a CA, and how to use your CA to sign a server certificate and a client certificate. You can use other CA providers, for example [Let's Encrypt](https://letsencrypt.org/).

The procedures below assume that your Harbor registry's hostname is `yourdomain.com`, and that its DNS record points to the host on which you are running Harbor. 

## Generate a Certificate Authority Certificate

In a production environment, you should obtain a certificate from a CA. In a test or development environment, you can generate your own CA. To generate a CA certficate, run the following commands. 

1. Generate a CA certificate private key.

    ```sh
    openssl genrsa -out ca.key 4096
    ```

1. Generate the CA certificate.

   Adapt the values in the `-subj` option to reflect your organization. If you use an FQDN to connect your Harbor host, you must specify it as the common name (`CN`) attribute.
   
    ```sh
    openssl req -x509 -new -nodes -sha512 -days 3650 \
     -subj "/C=CN/ST=Beijing/L=Beijing/O=example/OU=Personal/CN=yourdomain.com" \
     -key ca.key \
     -out ca.crt
    ```

## Generate a Server Certificate

The certificate usually contains a `.crt` file and a `.key` file, for example, `yourdomain.com.crt` and `yourdomain.com.key`.

1. Generate a private key.

    ```sh
    openssl genrsa -out yourdomain.com.key 4096
    ```

1. Generate a certificate signing request (CSR).

   Adapt the values in the `-subj` option to reflect your organization. If you use an FQDN to connect your Harbor host, you must specify it as the common name (`CN`) attribute and use it in the key and CSR filenames.

    ```sh
    openssl req -sha512 -new \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=example/OU=Personal/CN=yourdomain.com" \
        -key yourdomain.com.key \
        -out yourdomain.com.csr
    ```

1. Generate an x509 v3 extension file.

    Regardless of whether you're using either an FQDN or an IP address to connect to your Harbor host, you must create this file so that you can generate a certificate for your Harbor host that complies with the Subject Alternative Name (SAN) and x509 v3 extension requirements. Replace the `DNS` entries to reflect your domain.

    ```sh
    cat > v3.ext <<-EOF
    authorityKeyIdentifier=keyid,issuer
    basicConstraints=CA:FALSE
    keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
    extendedKeyUsage = serverAuth
    subjectAltName = @alt_names

    [alt_names]
    DNS.1=yourdomain.com
    DNS.2=yourdomain
    DNS.3=hostname
    EOF
    ```

1. Use the `v3.ext` file to generate a certificate for your Harbor host.
   
    Replace the `yourdomain.com` in the CRS and CRT file names with the Harbor host name.

    ```sh
    openssl x509 -req -sha512 -days 3650 \
        -extfile v3.ext \
        -CA ca.crt -CAkey ca.key -CAcreateserial \
        -in yourdomain.com.csr \
        -out yourdomain.com.crt
    ```

## Provide the Certificates to Harbor and Docker

After generating the `ca.crt`, `yourdomain.com.crt`, and `yourdomain.com.key` files, you must provide them to Harbor and to Docker, and reconfigure Harbor to use them.

1. Copy the server certificate and key into the certficates folder on your Harbor host.

    ```sh
    cp yourdomain.com.crt /data/cert/
    cp yourdomain.com.key /data/cert/
    ```

1. Convert `yourdomain.com.crt` to `yourdomain.com.cert`, for use by Docker.

    The Docker daemon interprets `.crt` files as CA certificates and `.cert` files as client certificates.

    ```sh
    openssl x509 -inform PEM -in yourdomain.com.crt -out yourdomain.com.cert
    ```

1. Copy the server certificate, key and CA files into the Docker certificates folder on the Harbor host. You must create the appropriate folders first.

    ```sh
    cp yourdomain.com.cert /etc/docker/certs.d/yourdomain.com/
    cp yourdomain.com.key /etc/docker/certs.d/yourdomain.com/
    cp ca.crt /etc/docker/certs.d/yourdomain.com/
    ```

    If you mapped the default `nginx` port 443 to a different port, create the folder `/etc/docker/certs.d/yourdomain.com:port`, or `/etc/docker/certs.d/harbor_IP:port`.

1. Restart Docker Engine.

    ```sh
    systemctl restart docker
    ```

You might also need to trust the certificate at the OS level. See [Troubleshooting Harbor Installation](troubleshoot-installation.md#https) for more information.

The following example illustrates a configuration that uses custom certificates.

```
/etc/docker/certs.d/
    └── yourdomain.com:port
       ├── yourdomain.com.cert  <-- Server certificate signed by CA
       ├── yourdomain.com.key   <-- Server key signed by CA
       └── ca.crt               <-- Certificate authority that signed the registry certificate
```

## Deploy or Reconfigure Harbor

If you have not yet deployed Harbor, see [Configure the Harbor YML File](configure-yml-file.md) for information about how to configure Harbor to use the certificates by specifying the `hostname` and `https` attributes in `harbor.yml`.

If you already deployed Harbor with HTTP and want to reconfigure it to use HTTPS, perform the following steps.

1. Run the `prepare` script to enable HTTPS.

    Harbor uses an `nginx` instance as a reverse proxy for all services. You use the `prepare` script to configure `nginx` to use HTTPS. The `prepare` is in the Harbor installer bundle, at the same level as the `install.sh` script.

    ```sh
    ./prepare
    ```

1. If Harbor is running, stop and remove the existing instance. 

    Your image data remains in the file system, so no data is lost.

    ```sh
    docker-compose down -v
    ```

1. Restart Harbor:

    ```sh
    docker-compose up -d
    ```

## Verify the HTTPS Connection

After setting up HTTPS for Harbor, you can verify the HTTPS connection by performing the following steps.

* Open a browser and enter https://yourdomain.com. It should display the Harbor interface.

    Some browsers might show a warning stating that the Certificate Authority (CA) is unknown. This happens when using a self-signed CA that is not from a trusted third-party CA. You can import the CA to the browser to remove the warning.

* On a machine that runs the Docker daemon, check the `/etc/docker/daemon.json` file to make sure that the `-insecure-registry` option is not set for https://yourdomain.com.

* Log into Harbor from the Docker client.

    ```sh
    docker login yourdomain.com
    ```

    If you've mapped `nginx` 443 port to a different port,add the port in the `login` command.

    ```sh
    docker login yourdomain.com:port
    ```
   
## What to Do Next

- If the verification succeeds, see [Harbor Administration](../administration) for information about using Harbor.
- If installation fails, see [Troubleshooting Harbor Installation](troubleshoot-installation.md).
