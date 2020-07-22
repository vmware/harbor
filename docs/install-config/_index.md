---
title: Harbor Installation and Configuration
weight: 5
---

This section describes how to perform a new installation of Harbor.

If you are upgrading from a previous version of Harbor, you might need to update the configuration file and migrate your data to fit the database schema of the later version. For information about upgrading, see [Upgrading Harbor](../administration/upgrade/_index.md).

Before you install Harbor, you can test the latest version of Harbor on a demo environment maintained by the Harbor team. For information, see [Test Harbor with the Demo Server](demo-server.md).

Harbor supports integration with different 3rd-party replication adapters for replicating data, OIDC adapters for authN/authZ, and scanner adapters for vulnerability scanning of container images. For information about the supported adapters, see the [Harbor Compatibility List](harbor-compatibility-list.md).

## Installation Process

The standard Harbor installation process involves the following stages:

1. Make sure that your target host meets the [Harbor Installation Prerequisites](installation-prereqs.md).
1. [Download the Harbor Installer](download-installer.md)
1. [Configure HTTPS Access to Harbor](configure-https.md)
1. [Configure the Harbor YML File](configure-yml-file.md)
1. [Configure Enabling Internal TLS](configure-internal-tls.md)
1. [Run the Installer Script](run-installer-script.md)

If installation fails, see [Troubleshooting Harbor Installation](troubleshoot-installation.md).

## Quick Installation

You can run a script that deploys Harbor to Ubuntu 18.04 with a single command. For information, see [Deploy Harbor with the Quick Installation Script](quick-install-script.md).

## Deploy Harbor on Kubernetes

You can also use Helm to install Harbor on a Kubernetes cluster, to make Harbor highly available. For information about installing Harbor with Helm on a Kubernetes cluster, see [Deploying Harbor with High Availability via Helm](harbor-ha-helm.md).

## Post-Installation Configuration

For information about how to manage your deployed Harbor instance, see [Reconfigure Harbor and Manage the Harbor Lifecycle](reconfigure-manage-lifecycle.md).

By default, Harbor uses its own private key and certificate to authenticate with Docker. For information about how to optionally customize your configuration to use your own key and certificate, see [Customize the Harbor Token Service](customize-token-service.md).

After installation, log into your Harbor via the web console to configure the instance under 'configuration'.  Harbor also provides a command line interface (CLI) that allows you to [Configure Harbor User Settings at the Command Line](configure-user-settings-cli.md).

## Harbor Components

The table below lists the some of the key components that are deployed when you deploy Harbor.

|Component|Version|
|---|---|
|Postgresql|9.6.10-1.ph2|
|Redis|4.0.10-1.ph2|
|Clair|2.0.8|
|Beego|1.9.0|
|Chartmuseum|0.9.0|
|Docker/distribution|2.7.1|
|Docker/notary|0.6.1|
|Helm|2.9.1|
|Swagger-ui|3.22.1|

