---
title: Troubleshooting Harbor Installation
weight: 50
---

The following sections help you to solve problems when installing Harbor.

## Access Harbor Logs

By default, registry data is persisted in the host's `/data/` directory. This data remains unchanged even when Harbor's containers are removed and/or recreated, you can edit the `data_volume` in `harbor.yml` file to change this directory.

In addition, Harbor uses `rsyslog` to collect the logs of each container. By default, these log files are stored in the directory `/var/log/harbor/` on the target host for troubleshooting, also you can change the log directory in `harbor.yml`.

## Harbor Does Not Start or Functions Incorrectly

If Harbor does not start or functions incorrectly, run the following command to check whether all of Harbor's containers are in the `Up` state.

```
sudo docker-compose ps
        Name                     Command               State                    Ports
  -----------------------------------------------------------------------------------------------------------------------------
  harbor-core         /harbor/start.sh                 Up
  harbor-db           /entrypoint.sh postgres          Up      5432/tcp
  harbor-jobservice   /harbor/start.sh                 Up
  harbor-log          /bin/sh -c /usr/local/bin/ ...   Up      127.0.0.1:1514->10514/tcp
  harbor-portal       nginx -g daemon off;             Up      80/tcp
  nginx               nginx -g daemon off;             Up      0.0.0.0:443->443/tcp, 0.0.0.0:4443->4443/tcp, 0.0.0.0:80->80/tcp
  redis               docker-entrypoint.sh redis ...   Up      6379/tcp
  registry            /entrypoint.sh /etc/regist ...   Up      5000/tcp
  registryctl         /harbor/start.sh                 Up
```

If a container is not in the `Up` state, check the log file for that container in `/var/log/harbor`. For example, if the `harbor-core` container is not running, look at the `core.log` log file.

## Using `nginx` or Load Balancing

If Harbor is running behind an `nginx` proxy or elastic load balancing, open the file `common/config/nginx/nginx.conf` and search for the following line.

```
proxy_set_header X-Forwarded-Proto $scheme;
```

If the proxy already has similar settings, remove it from the sections `location /`, `location /v2/` and `location /service/` and redeploy Harbor. For instructions about how to redeploy Harbor, see [Reconfigure Harbor and Manage the Harbor Lifecycle](reconfigure-manage-lifecycle.md).

## Troubleshoot HTTPS Connections {#https}

If you use an intermediate certificate from a certificate issuer, merge the intermediate certificate with your own certificate to create a certificate bundle. Run the following command.

```
cat intermediate-certificate.pem >> yourdomain.com.crt
```
When the Docker daemon runs on certain operating systems, you might need to trust the certificate at the OS level. For example, run the following commands.

- Ubuntu:

    ```sh
    cp yourdomain.com.crt /usr/local/share/ca-certificates/yourdomain.com.crt 
    update-ca-certificates
    ```

- Red Hat (CentOS etc):

    ```sh
    cp yourdomain.com.crt /etc/pki/ca-trust/source/anchors/yourdomain.com.crt
    update-ca-trust
    ```
