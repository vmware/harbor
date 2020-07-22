---
title: Configure LDAP/Active Directory Authentication
weight: 20
---

If you select LDAP/AD authentication, users whose credentials are stored in an external LDAP or AD server can log in to Harbor directly. In this case, you do not create user accounts in Harbor.

{{< important >}}
You can change the authentication mode from database to LDAP only if no local users have been added to the database. If there is at least one user other than `admin` in the Harbor database, you cannot change the authentication mode.
{{< /important >}}

Because the users are managed by LDAP or AD, self-registration, creating users, deleting users, changing passwords, and resetting passwords are not supported in LDAP/AD authentication mode.  

If you want to manage user authentication by using LDAP groups, you must enable the `memberof` feature on the LDAP/AD server. With the `memberof` feature, the LDAP/AD user entity's `memberof` attribute is updated when the group entity's `member` attribute is updated, for example by adding or removing an LDAP/AD user from the LDAP/AD group. This feature is enabled by default in Active Directory. For information about how to enable and verify `memberof` overlay in OpenLDAP, see [this technical note](https://technicalnotes.wordpress.com/2014/04/19/openldap-setup-with-memberof-overlay).

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Under **Administration**, go to **Configuration** and select the **Authentication** tab.
1. Use the **Auth Mode** drop-down menu to select **LDAP**.

   ![LDAP authentication](../../../img/select-ldap-auth.png)
1. Enter the address of your LDAP server, for example `ldaps://10.162.16.194`.
1. Enter information about your LDAP server.

   - **LDAP Search DN** and **LDAP Search Password**: When a user logs in to Harbor with their LDAP username and password, Harbor uses these values to bind to the LDAP/AD server. For example, `cn=admin,dc=example.com`.
   - **LDAP Base DN**: Harbor looks up the user under the LDAP Base DN entry, including the subtree. For example, `dc=example.com`.
   - **LDAP Filter**: The filter to search for LDAP/AD users. For example, `objectclass=user`. 
   - **LDAP UID**: An attribute, for example `uid`, or `cn`, that is used to match a user with the username. If a match is found, the user's password is verified by a bind request to the LDAP/AD server. 
   - **LDAP Scope**: The scope to search for LDAP/AD users. Select from **Subtree**, **Base**, and **OneLevel**.
   
     ![Basic LDAP configuration](../../../img/ldap-auth.png)  
1. If you want to manage user authentication with LDAP groups, configure the group settings.
   - **LDAP Group Base DN**: The base DN from which to lookup a group in LDAP/AD. For example, `ou=groups,dc=example,dc=com`.
   - **LDAP Group Filter**: The filter to search for LDAP/AD groups. For example, `objectclass=groupOfNames`. 
   - **LDAP Group GID**: The attribute used to name an LDAP/AD group. For example, `cn`.  
   - **LDAP Group Admin DN**: All LDAP/AD users in this group DN have Harbor system administrator privileges.
   - **LDAP Group Membership**: The user attribute usd to identify a user as a member of a group. By default this is `memberof`.
   - **LDAP Scope**: The scope to search for LDAP/AD groups. Select from **Subtree**, **Base**, and **OneLevel**.
   
     ![LDAP group configuration](../../../img/ldap-groups.png)
1. Uncheck **LDAP Verify Cert** if the LDAP/AD server uses a self-signed or untrusted certificate.

   ![LDAP certificate verification](../../../img/ldap-cert-test.png)
1. Click **Test LDAP Server** to make sure that your configuration is correct.
1. Click **Save** to complete the configuration.
