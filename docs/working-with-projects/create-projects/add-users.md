---
title: Assign Users to a Project
weight: 25
---

You can add individual users to an existing project and assign a role to them. You can add an LDAP/AD or OIDC user to the project members if you  use LDAP/AD or OIDC authentication, or a user that you have already created if you use database authentication. If you use LDAP/AD or OIDC authentication, you can add groups to projects and assign a role to the group.

For more information about users and roles in Harbor, see [User Permissions By Role](../../administration/managing-users/user-permissions-by-role.md).

## Add Individual Members to Projects 

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects** and select a project. 
1. Select the **Members** tab and click **+User**.

   ![browse project](../../../img/project-members.png)
1. Enter the name of an existing database, LDAP/AD, or OIDC user and select a role for this user.

   ![browse project](../../../img/new-add-member.png)
1. Optionally select one or more members, click **Action**, and select a different role for the user or users, or select **Remove** to remove them from the project.

   ![browse project](../../../img/new-remove-update-member.png)

## Add LDAP/AD Groups to Projects

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects** and select a project. 
1. Select the **Members** tab and click **+Group**.

   ![Add group](../../../img/add-group.png)
1. Select **Add an existing user group to project members** or **Add a group from LDAP to project member**.

   ![Screenshot of add group dialog](../../../img/ldap-group-addgroup-dialog.png)
   
   - If you selected **Add an existing user group to project members**, enter the name of a group that you have already used in Harbor and assign a role to that group.
   - If you selected **Add a group from LDAP to project member**, enter the LDAP Group DN and assign a role to that group.

Once an LDAP group has been assigned a role in a project, all LDAP/AD users in this group have the privileges of the role you assigned to the group. If a user has both user-level role and group-level role, these privileges are merged.

If a user in the LDAP group has admin privilege, the user has the same privileges as the Harbor system administrator.

## Add OIDC Groups to Projects

To be able to add OIDC groups to projects, your OIDC provider and Harbor instance must be configured correctly. For information about how to configure OIDC so that Harbor can use groups, see [OIDC Provider Authentication](../../administration/configure-authentication/oidc-auth.md).

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects** and select a project. 
1. Select the **Members** tab and click **+Group**.

   ![Add group](../../../img/add-group.png)
1. Enter the name of a group that already exists in your OIDC provider and assign a role to that group.

   ![Add group](../../../img/add-oidc-group.png)

{{< note >}}
Unlike with LDAP groups, Harbor cannot check whether OIDC groups exist when you add them to a project. If you mistype the group name, or if the group does not exist in your OIDC provider, Harbor still creates the group.
{{< /note >}}
