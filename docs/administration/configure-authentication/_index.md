---
title: Configuring Authentication
weight: 10
---

Harbor supports different modes for authenticating users and managing user accounts. You should select an authentication mode as soon as you deploy Harbor.

{{< important >}}
If you create user accounts in the Harbor database, Harbor is locked in database mode. You cannot change to a different authentication mode after you have created local users.
{{< /important >}}

- [Database Authentication](db-auth.md): You create and manage user accounts directly in Harbor. The user accounts are stored in the Harbor database.
- [LDAP/Active Directory Authentication](ldap-auth.md): You connect Harbor to an external LDAP/Active Directory server. The user accounts are created and managed by your LDAP/AD provider.
- [OIDC Provider Authentication](oidc-auth.md): You connect Harbor to an external OIDC provider. The user accounts are created and managed by your OIDC provider.

The Harbor interface offers an option to configure UAA authentication. This authentication mode is not recommended and is not documented in this guide.
