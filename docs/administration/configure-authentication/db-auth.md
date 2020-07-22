---
title: Configure Database Authentication
weight: 15
---

In database authentication mode, user accounts are stored in the local database. By default, only the Harbor system administrator can create user accounts to add users to Harbor. You can optionally configure Harbor to allow self-registration.  

{{< important >}}
If you create users in the database, Harbor is locked in database mode. You cannot change to a different authentication mode after you have created local users.
{{< /important >}}

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Under **Administration**, go to **Configuration** and select the **Authentication** tab.
1. Leave **Auth Mode** set to the default **Database** option.

   ![Database authentication](../../../img/db-auth.png)
   
1. Optionally select the **Allow Self-Registration** check box.

   ![Enable self-registration](../../../img/new-self-reg.png)
    
   If you enable the self registration option, users can register themselves in Harbor. Self-registration is disabled by default. If you enable self-registration, unregistered users can sign up for a Harbor account by clicking **Sign up for an account** in the Harbor log in page.
    
    ![Enable self-registration](../../../img/self-registration-login.png)
    
## What to Do Next

For information about how to create users in database authentication mode, see [Create User Accounts in Database Mode](../managing-users/create-users-db.md).
