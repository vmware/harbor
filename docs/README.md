Harbor Documentation

This is the main table of contents for the Harbor 1.10.x documentation.

## Harbor Installation and Configuration

This section describes how to install Harbor and perform the required initial configurations. These day 1 operations are performed by the Harbor Administrator.

- [Introduction](install-config/_index.md) 
- [Test Harbor with the Demo Server](install-config/demo-server.md)
- [Harbor Compatibility List](install-config/harbor-compatibility-list.md)
- [Harbor Installation Prerequisites](install-config/installation-prereqs.md)
- [Download the Harbor Installer](install-config/download-installer.md)
- [Configure HTTPS Access to Harbor](install-config/configure-https.md)
- [Configure the Harbor YML File](install-config/configure-yml-file.md)
- [Run the Installer Script](install-config/run-installer-script.md)
- [Deploying Harbor with High Availability via Helm](install-config/harbor-ha-helm.md)
- [Deploy Harbor with the Quick Installation Script](install-config/quick-install-script.md)
- [Troubleshooting Harbor Installation](install-config/troubleshoot-installation.md)
- [Reconfigure Harbor and Manage the Harbor Lifecycle](install-config/reconfigure-manage-lifecycle.md)
- [Customize the Harbor Token Service](install-config/customize-token-service.md)
- [Configure Harbor User Settings at the Command Line](install-config/configure-user-settings-cli.md)
  
## Harbor Administration

This section describes how to use and maintain Harbor after deployment. These day 2 operations are performed by the Harbor Administrator.

- [Introduction](administration/_index.md)
- [Configuring Authentication](administration/configure-authentication/_index.md)
   - [Configure Database Authentication](administration/configure-authentication/db-auth.md)
   - [Configure LDAP/Active Directory Authentication](administration/configure-authentication/ldap-auth.md)
   - [Configure OIDC Provider Authentication](administration/configure-authentication/oidc-auth.md)
- [Managing Users](administration/managing-users/_index.md)
    - [User Permissions By Role](administration/managing-users/user-permissions-by-role.md)
    - [Create User Accounts in Database Mode](administration/managing-users/create-users-db.md)
- [Configure Global Settings](administration/general-settings/_index.md)
- [Configure Project Quotas](administration/configure-project-quotas/_index.md)
- [Configuring Replication](administration/configuring-replication/_index.md)
    - [Create Replication Endpoints](administration/configuring-replication/create-replication-endpoints.md)
    - [Create Replication Rules](administration/configuring-replication/create-replication-rules.md)
    - [Manage Replications](administration/configuring-replication/manage-replications.md) 
- [Vulnerability Scanning](administration/vulnerability-scanning/_index.md)
    - [Connect Harbor to Additional Vulnerability Scanners](administration/vulnerability-scanning/pluggable-scanners.md)
    - [Scan Individual Images](administration/vulnerability-scanning/scan-individual-image.md)
    - [Scan All Images](administration/vulnerability-scanning/scan-all-images.md)
    - [Schedule Scans](administration/vulnerability-scanning/schedule-scans.md)
    - [Import Vulnerability Data to an Offline Harbor instance](administration/vulnerability-scanning/import-vulnerability-data.md)
    - [Configure System-Wide CVE Allowlists](administration/vulnerability-scanning/configure-system-allowlist.md)
- [Garbage Collection](administration/garbage-collection/_index.md)
- [Upgrade Harbor and Migrate Data](administration/upgrade/upgrade-migrate-data.md)
  - [Upgrading Harbor Deployed with Helm](administration/upgrade/helm-upgrade.md)
  - [Roll Back an Upgrade](administration/upgrade/roll-back-upgrade.md)
  - [Test Harbor Upgrade](administration/upgrade/upgrade-test.md)

## Working with Harbor Projects

This section describes how users with the developer, master, and project administrator roles manage and participate in Harbor projects.

- [Introduction](working-with-projects/_index.md)
- [Create Projects](working-with-projects/create-projects/_index.md)
    - [Assign Users to a Project](working-with-projects/add-users.md)
- [Project Configuration](working-with-projects/project-configuration/_index.md)
    - [Access and Search Project Logs](working-with-projects/access-project-logs.md)
    - [Create Robot Accounts](working-with-projects/create-robot-accounts.md)
    - [Configure Webhook Notifications](working-with-projects/configure-webhooks.md)
    - [Configure a Per-Project CVE Allowlist](working-with-projects/configure-project-allowlist.md)
    - [Implementing Content Trust](working-with-projects/implementing-content-trust.md)
- [Working with Images, Tags, and Helm Charts](working-with-projects/working-with-images.md)
    - [Pulling and Pushing Images](working-with-projects/pulling-pushing-images.md)
    - [Create Labels](working-with-projects/create-labels.md)
    - [Retag Images](working-with-projects/retagging-images.md)
    - [Create Tag Retention Rules](working-with-projects/create-tag-retention-rules.md)
    - [Create Tag Immutability  Rules](working-with-projects/create-tag-immutability-rules.md)
    - [Manage Kubernetes Packages with Helm Charts](working-with-projects/managing-helm-charts.md)
- [Using API Explorer](working-with-projects/using-api-explorer/_index.md)

## Build, Customize, and Contribute to Harbor

This section describes how developers can build from Harbor source code, customize their deployments, and contribute to the open-source Harbor project.

- [Build Harbor from Source Code](build-customize-contribute/compile-guide.md)
- [Developing the Harbor Frontend](build-customize-contribute/ui-contribution-get-started.md)
- [Customize the Harbor Look & Feel ](build-customize-contribute/customize-look-feel.md)
- [Developing for Internationalization](build-customize-contribute/developer-guide-i18n.md)
- [Using Make](build-customize-contribute/use-make.md)
- [View and test Harbor REST API via Swagger](build-customize-contribute/configure-swagger.md)
- [Registry Landscape](build-customize-contribute/registry-landscape.md)
- [E2E Test Scripting Guide](build-customize-contribute/e2e_api_python_based_scripting_guide.md)

See also the list of [Articles from the Harbor Community](https://github.com/goharbor/harbor/blob/master/docs/README.md#articles-from-the-community).
