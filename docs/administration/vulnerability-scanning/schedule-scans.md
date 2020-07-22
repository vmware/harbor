---
title: Schedule Scans
weight: 40
---

You can set policies to control when vulnerability scanning should run.

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Expand **Administration**, and select **Interrogation Services**. 
1. Select the **Vulnerability** tab and click the **Edit** button next to **Schedule to scan all**.  
1. Use the drop down-menu to select how often to run scans.

   ![browse project](../../../img/scan-policy.png)
   
   * **None**: No scans are scheduled.
   * **Hourly**: Run a scan at the beginning of every hour.
   * **Daily**: Run a scan at midnight every day.
   * **Weekly**: Run a scan at midnight every Saturday.
   * **Custom**: Run a scan according to a `cron` job.
1. Click **Save**.
