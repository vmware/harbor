---
title: Configure the Harbor YML File
weight: 35
---

You set system level parameters for Harbor in the `harbor.yml` file that is contained in the installer package. These parameters take effect when you run the `install.sh` script to install or reconfigure Harbor.

After the initial deployment and after you have started Harbor, you perform additional configuration in the Harbor Web Portal.

## Required Parameters

The table below lists the parameters that must be set when you deploy Harbor. By default, all of the required parameters are uncommented in the `harbor.yml` file. The optional parameters are commented with `#`. You do not necessarily need to change the values of the required parameters from the defaults that are provided, but these parameters must remain uncommented. At the very least, you must update the `hostname` parameter.

**IMPORTANT**: Harbor does not ship with any certificates. In versions up to and including 1.9.x, by default Harbor uses HTTP to serve registry requests. This is acceptable only in air-gapped test or development environments. In production environments, always use HTTPS. If you enable Content Trust with Notary to properly sign all images, you must use HTTPS.

You can use certificates that are signed by a trusted third-party CA, or you can use self-signed certificates. For information about how to create a CA, and how to use a CA to sign a server certificate and a client certificate, see [Configuring Harbor with HTTPS Access](configure-https.md).

<table width="100%" border="0">
  <caption>
    Required Parameters for Harbor Deployment
  </caption>
  <tr>
    <th scope="col">Parameter</th>
    <th scope="col">Sub-parameters</th>
    <th scope="col">Description and Additional Parameters </th>
  </tr>
  <tr>
    <td valign="top"><code>hostname</code></td>
    <td valign="top">None</td>
    <td valign="top">Specify the IP address or the fully qualified domain name (FQDN) of the target host on which to deploy Harbor. This is the address at which you access the Harbor Portal and the registry service. For example, <code>192.168.1.10</code> or <code>reg.yourdomain.com</code>. The registry service must be accessible to external clients, so do not specify <code>localhost</code>, <code>127.0.0.1</code>, or <code>0.0.0.0</code> as the hostname.</td>
  </tr>
  <tr>
    <td valign="top"><code>http</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Do not use HTTP in production environments. Using HTTP is acceptable only in air-gapped test or development environments that do not have a connection to the external internet. Using HTTP in environments that are not air-gapped exposes you to man-in-the-middle attacks.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>port</code></td>
    <td valign="top">Port number for HTTP, for both Harbor portal and Docker commands. The default is 80.</td>
  </tr>
  <tr>
    <td valign="top"><code>https</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Use HTTPS to access the Harbor Portal and the token/notification service. Always use HTTPS in production environments and environments that are not air-gapped.
      </td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>port</code></td>
    <td valign="top">The port number for HTTPS, for both Harbor portal and Docker commands. The default is 443.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>certificate</code></td>
    <td valign="top">The path to the SSL certificate.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>private_key</code></td>
    <td valign="top">The path to the SSL key.</td>
  </tr>

  <tr>
    <td valign="top"><code>internal_tls</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top"> Use HTTPS to communicate between harbor components</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>enabled</code></td>
    <td valign="top">Set this flag to <code>true</code> means internal tls is enabled</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>certificate</code></td>
    <td valign="top">The path to the directory that contains internal certs and keys</td>
  </tr>

  <tr>
    <td valign="top"><code>harbor_admin_password</code></td>
    <td valign="top">None</td>
    <td valign="top">Set an initial password for the Harbor system administrator. This password is only used on the first time that Harbor starts. On subsequent logins, this setting is ignored and the administrator's password is set in the Harbor Portal. The default username and password are <code>admin</code> and <code>Harbor12345</code>.</td>
  </tr>
  <tr>
    <td valign="top"><code>database</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Use a local PostgreSQL database. You can optionally configure an external database, in which case you can disable this option.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>password</code></td>
    <td valign="top">Set the root password for the local database. You must change this password for production deployments.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>max_idle_conns</code></td>
    <td valign="top">The maximum number of connections in the idle connection pool. If set to &lt;=0 no idle connections are retained. The default value is 50. If it is not configured the value is 2.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>max_open_conns</code></td>
    <td valign="top">The maximum number of open connections to the database. If &lt;= 0 there is no limit on the number of open connections. The default value is 100 for the max connections to the Harbor database. If it is not configured the value is 0.</td>
  </tr>
  <tr>
    <td valign="top"><code>data_volume</code></td>
    <td valign="top">None</td>
    <td valign="top">The location on the target host in which to store Harbor's data. This data remains unchanged even when Harbor's containers are removed and/or recreated. You can optionally configure external storage, in which case disable this option and enable <code>storage_service</code>. The default is <code>/data</code>.</td>
  </tr>
  <tr>
    <td valign="top"><code>clair</code></td>
    <td valign="top"><code>updaters_interval</code></td>
    <td valign="top">Set an interval for Clair updates, in hours. Set to 0 to disable the updates. The default is 12 hours.</td>
  </tr>
  <tr>
    <td valign="top"><code>trivy</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure Trivy scanner.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>ignore_unfixed</code></td>
    <td valign="top">Set the flag to <code>true</code> to display only fixed vulnerabilities. The default value is <code>false</code></td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>skip_update</code></td>
    <td valign="top">You might want to enable this flag in test or CI/CD environments to avoid GitHub rate limiting issues. If the flag is enabled you have to manually download the `trivy.db` file and mount it in the <code>/home/scanner/.cache/trivy/db/trivy.db</code> path in container. The default value is <code>false</code></td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>insecure</code></td>
    <td valign="top">Set the flag to <code>true</code> to skip verifying registry certificate. The default value is <code>false</code></td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>github_token</code></td>
    <td valign="top">Set the GitHub access token to download Trivy DB. Trivy DB is downloaded by Trivy from the GitHub release page. Anonymous downloads from GitHub are subject to the limit of 60 requests per hour. Normally such rate limit is enough for production operations. If, for any reason, it's not enough, you could increase the rate limit to 5000 requests per hour by specifying the GitHub access token. For more details on GitHub rate limiting please consult https://developer.github.com/v3/#rate-limiting .You can create a GitHub token by following the instuctions in https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line</td>
  </tr>
  <tr>
    <td valign="top"><code>jobservice</code></td>
    <td valign="top"><code>max_job_workers</code></td>
    <td valign="top">The maximum number of replication workers in the job service. For each image replication job, a worker synchronizes all tags of a repository to the remote destination. Increasing this number allows more concurrent replication jobs in the system. However, since each worker consumes a certain amount of network/CPU/IO resources, set the value of this attribute based on the hardware resource of the host. The default is 10.</td>
  </tr>
<tr>
    <td valign="top"><code>notification</code></td>
    <td valign="top"><code>webhook_job_max_retry</code></td>
    <td valign="top">Set the maximum number of retries for web hook jobs. The default is 10.</td>
  </tr>
  <tr>
    <td valign="top"><code>chart</code></td>
    <td valign="top"><code>absolute_url</code></td>
    <td valign="top">Set to <code>enabled</code> for Chart to use an absolute URL. Set to <code>disabled</code> for Chart to use a relative URL.</td>
  </tr>
  <tr>
    <td valign="top"><code>log</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure logging. Harbor uses `rsyslog` to collect the logs for each container.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>level</code></td>
    <td valign="top">Set the logging level to <code>debug</code>, <code>info</code>, <code>warning</code>, <code>error</code>, or <code>fatal</code>. The default is <code>info</code>.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>local</code></td>
    <td valign="top">Set the log retention parameters:<ul>
          <li><code>rotate_count</code>: Log files are rotated <code>rotate_count</code> times before being removed. If count is 0, old versions are removed rather than rotated. The default is 50.</li>
          <li><code>rotate_size</code>: Log files are rotated only if they grow bigger than <code>rotate_size</code> bytes. Use <code>k</code> for kilobytes, <code>M</code> for megabytes, and <code>G</code> for gigabytes.  <code>100</code>, <code>100k</code>, <code>100M</code> and <code>100G</code> are all valid values. The default is 200M.</li>
          <li><code>location</code>: Set the directory in which to store the logs. The default is <code>/var/log/harbor</code>.</li>
        </ul></td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>external_endpoint</code></td>
    <td valign="top">Enable this option to forward logs to a syslog server.
      <ul>
        <li><code>protocol</code>: Transport protocol for the syslog server. Default is TCP.</li>
        <li><code>host</code>: The URL of the syslog server.</li>
        <li><code>port</code>: The port on which the syslog server listens</li>
    </ul>    </td>
  </tr>
  <tr>
    <td valign="top"><code>proxy</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure proxies to be used by Clair, the replication jobservice, and Harbor. Leave blank if no proxies are required.</td>
  </tr>
    <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>http_proxy</code></td>
    <td valign="top">Configure an HTTP proxy, for example,  <code>http://my.proxy.com:3128</code>.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>https_proxy</code></td>
    <td valign="top">Configure an HTTPS proxy, for example,  <code>http://my.proxy.com:3128</code>.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>no_proxy</code></td>
    <td valign="top">Configure when not to use a proxy, for example, <code>127.0.0.1,localhost,core,registry</code>.</td>
  </tr>
</table>

## Optional Parameters

The following table lists the additional, optional parameters that you can set to configure your Harbor deployment beyond the minimum required settings. To enable a setting, you must uncomment it in `harbor.yml` by deleting the leading `#` character.

<table width="100%" border="0">
  <caption>
    Optional Parameters for Harbor
  </caption>
  <tr>
    <th scope="col">Parameter</th>
    <th scope="col">Sub-Parameters</th>
    <th scope="col">Description and Additional Parameters </th>
  </tr>
  <tr>
    <td valign="top"><code>external_url</code></td>
    <td valign="top">None</td>
    <td valign="top">Enable this option to use an external proxy. When  enabled, the hostname is no longer used.</td>
  </tr>
  <tr>
  <tr>
    <td valign="top"><code>storage_service</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">By default, Harbor stores images and charts on your local filesystem. In a production environment, you might want to use another storage backend instead of the local filesystem. The parameters listed below are the configurations for the registry. See *Configuring Storage Backend* below for more information about how to configure a different backend.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>ca_bundle</code></td>
    <td valign="top">The path to the custom root CA certificate, which is injected into the trust store of registry and chart repository containers. This is usually needed if internal storage uses a self signed certificate.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>filesystem</code></td>
    <td valign="top">The default is <code>filesystem</code>, but you can set <code>azure</code>, <code>gcs</code>, <code>s3</code>, <code>swift</code> and <code>oss</code>. For information about how to configure other backends, see <a href="#backend">Configuring a Storage Backend</a> below. Set <code>maxthreads</code> to limit the number of threads to the external provider. The default is 100.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>redirect</code></td>
    <td valign="top">Set <code>disable</code> to <code>true</code> when you want to disable registry redirect</td>
  </tr>
  <tr>
    <td valign="top"><code>external_database</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure external database settings, if you disable the local database option. Currently, Harbor only supports PostgreSQL database. You must create four databases for Harbor core, Clair, Notary server, and Notary signer. The tables are generated automatically when Harbor starts up.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>harbor</code></td>
    <td valign="top"><p>Configure an external database for Harbor data.</p>
      <ul>
        <li><code>host</code>: Hostname of the Harbor database.</li>
        <li><code>port</code>: Database port.</li>
        <li><code>db_name</code>: Database name.</li>
        <li><code>username</code>: Username to connect to the core Harbor database.</li>
        <li><code>password</code>: Password for the account you set in <code>username</code>.</li>
        <li><code>ssl_mode</code>: Enable SSL mode.</li>
        <li><code>max_idle_conns</code>: The maximum number of connections in the idle connection pool. If &lt;=0 no idle connections are retained. The default value is 2.</li>
        <li><code>max_open_conns</code>: The maximum number of open connections to the database. If &lt;= 0 there is no limit on the number of open connections. The default value is 0.</li>
    </ul>      </td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>clair</code></td>
    <td valign="top">Configure an external database for Clair.
      <ul>
        <li><code>host</code>: Hostname of the Clair database</li>
        <li><code>port</code>: Database port.</li>
        <li><code>db_name</code>: Database name.</li>
        <li><code>username</code>: Username to connect to the Clair database.</li>
        <li><code>password</code>: Password for the account you set in <code>username</code>.</li>
        <li><code>ssl_mode</code>: Enable SSL mode.</li>
      </ul>    </td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>notary_signer</code></td>
    <td valign="top">Configure an external database for the Notary signer database
      <ul>
        <li><code>host</code>: Hostname of the Notary signer database</li>
        <li><code>port</code>: Database port.</li>
        <li><code>db_name</code>: Database name.</li>
        <li><code>username</code>: Username to connect to the Notary signer database.</li>
        <li><code>password</code>: Password for the account you set in <code>username</code>.</li>
        <li><code>ssl_mode</code>: Enable SSL mode.</li>
      </ul>    </td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>notary_server</code></td>
    <td valign="top"><ul>
      <li><code>host</code>: Hostname of the Notary server database.</li>
      <li><code>port</code>: Database port.</li>
      <li><code>db_name</code>: Database name.</li>
      <li><code>username</code>: Username to connect to the Notary server database.</li>
      <li><code>password</code>: Password for the account you set in <code>username</code>.</li>
      <li><code>ssl_mode</code>: Enable SSL mode.e</li>
    </ul>    </td>
  </tr>
  <tr>
    <td valign="top"><code>external_redis</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure an external Redis instance.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>host</code></td>
    <td valign="top">Hostname of the external Redis instance.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>port</code></td>
    <td valign="top">Redis instance port.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>password</code></td>
    <td valign="top">Password to connect to the external Redis instance.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>registry_db_index</code></td>
    <td valign="top">Database index for Harbor registry.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>jobservice_db_index</code></td>
    <td valign="top">Database index for jobservice.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>chartmuseum_db_index</code></td>
    <td valign="top">Database index for Chart museum.</td>
  </tr>
</table>

{{< note >}}
The `harbor.yml` file includes options to configure a UAA CA certificate. This authentication mode is not recommended and is not documented.
{{< /note >}}

## Configuring a Storage Backend {#backend}

By default Harbor uses local storage for the registry, but you can optionally configure the `storage_service` setting so that Harbor uses external storage. For information about how to configure the storage backend of a registry for different storage providers, see the [Registry Configuration Reference](https://docs.docker.com/registry/configuration/#storage) in the Docker documentation. For example, if you use Openstack Swift as your storage backend, the parameters might resemble the following:

``` yaml
storage_service:
  ca_bundle:
  swift:
    username: admin
    password: ADMIN_PASS
    authurl: http://keystone_addr:35357/v3/auth
    tenant: admin
    domain: default
    region: regionOne
    container: docker_images"
  redirect:
    disable: false
```

## What to Do Next

To install Harbor, [Run the Installer Script](run-installer-script.md).
