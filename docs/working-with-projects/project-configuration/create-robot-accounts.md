---
title: Create Robot Accounts
weight: 40
---

You can create robot accounts to run automated operations. Robot accounts have the following limitations:

1. Robot Accounts cannot log in to the Harbor interface.
1. Robot Accounts can only perform operations by using the Docker and Helm CLIs.

### Add a Robot Account

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects**, select a project, and select **Robot Accounts**.

    ![Robot accounts](../../../img/add-robot-account.png)

1. Click **New Robot Account**.
1. Enter a name and an optional description for this robot account.
1. Set expiration time for this robot account, you can also select checkbox **Never Expired** if you want to create a never expiring robot account. If not set, the expiration time of system configuration will be used for this robot account. 
1. Grant permission to the robot account to push images and to push and pull Helm charts.

    Robot accounts can always pull images, so you cannot deselect this option.
   
    ![Add a robot account](../../../img/add-robot-account-2.png)

1. Click **Save**.
1. In the confirmation window, click **Export to File** to download the access token as a JSON file, or click the clipboard icon to copy its contents to the clipboard.

   ![copy_robot_account_token](../../../img/copy-robot-account-token.png)

   {{< important >}}
   Harbor does not store robot account tokens, so you must either download the token JSON or copy and paste its contents into a text file. There is no way to get the token from Harbor after you have created the robot account.
   {{< /important >}}

   The new robot account appears as `robot$account_name` in the list of robot accounts. The `robot$` prefix makes it easily distinguishable from a normal Harbor user account.

   ![New robot account](../../../img/new-robot-account.png)

1. To delete or disable a robot account, select the account in the list, and select **Disable account** or **Delete** from the Action drop-down menu.

    ![Disable or delete a robot account](../../../img/disable-delete-robot-account.png)

### Configure the Expiry Period of Robot Accounts

By default, robot accounts expire after 30 days. You can set a longer or shorter lifespan for robot accounts by modifying the expiry period for robot account tokens. The expiry period applies to all robot accounts in all projects.

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Go to **Configuration** and select **System Settings**.
1. In the **Robot Token Expiration (Days)** row, modify the number of days after which robot account tokens expire. 

    ![Set robot account token expiry](../../../img/set-robot-account-token-duration.png)

### Authenticate with a Robot Account

To use a robot account in an automated process, for example a script, use `docker login` and provide the credentials of the robot account.

<pre>
docker login <i>harbor_address</i>
Username: robot$<i>account_name</i>
Password: <i>robot_account_token</i>
</pre>
