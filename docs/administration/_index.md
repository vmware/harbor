---
title: Harbor Administration
weight: 10
---

This section describes how to configure and maintain Harbor after deployment. These operations are performed by the Harbor system administrator. The Harbor system administrator performs global configuration operations that apply to the whole Harbor instance.

The operations that are performed by the Harbor system administrator are the following.

- Select database, LDAP/Active Directory, or OIDC based authentication. For information, see [Configuring Authentication](configure-authentication).
- Add users in database authentication mode and assign the system administrator role to other users. For information, see [Managing Users](managing-users).
- Configure global settings, such as configuring an email server, setting the registry to read-only mode, and restriction who can create projects. For information, see [Configure Global Settings](general-settings).
- Apply resource quotas to projects. For information, see [Configure Project Quotas](configure-project-quotas).
- Set up replication of images between Harbor and another Harbor instance or a 3rd party replication target. For information, see [Configuring Replication](configuring-replication).
- Set up vulnerability scanners to check the images in the registry for CVE vulnerabilities. For information, see [Vulnerability Scanning](vulnerability-scanning).
- Perform garbage collection, to remove unnecessary data from Harbor. For information, see [Garbage Collection](garbage-collection).
- Upgrade Harbor when a new version becomes available. For information, see [Upgrading Harbor](upgrade/upgrade-migrate-data.md).
