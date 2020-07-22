---
title: Configure a Per-Project CVE Allowlist
weight: 50
---

When you run vulnerability scans, images that are subject to Common Vulnerabilities and Exposures (CVE) are identified. According to the severity of the CVE and your security settings, these images might not be permitted to run. You can create allowlists of CVEs to ignore during vulnerability scanning. 

Harbor administrators can set a system-wide CVE allowlist. For information about site-wide CVE allowlists, see [Configure System-Wide CVE Allowlists](../../administration/vulnerability-scanning/configure-system-allowlist.md). By default, the system allowlist is applied to all projects. You can configure different CVE allowlists for individual projects, that override the system allowlist. 

1. Go to **Projects**, select a project, and select **Configuration**.
1. Under **CVE allowlist**, select **Project allowlist**.

    ![Project CVE allowlist](../../../img/cve-allowlist5.png)

1. Optionally click **Copy From System** to add all of the CVE IDs from the system CVE allowlist to this project allowlist.
1. Click **Add** and enter a list of additional CVE IDs to ignore during vulnerability scanning of this project.

    ![Add project CVEs](../../../img/cve-allowlist6.png)

    Either use a comma-separated list or newlines to add multiple CVE IDs to the list.

1. Click **Add** at the bottom of the window to add the CVEs to the project allowlist.
1. Optionally uncheck the **Never expires** checkbox and use the calendar selector to set an expiry date for the allowlist.
1. Click **Save** at the bottom of the page to save your settings.

After you have created a project allowlist, you can remove CVE IDs from the list by clicking the delete button next to it in the list. You can click **Add** at any time to add more CVE IDs to this project allowlist. 

If CVEs are added to the system allowlist after you have created a project allowlist, click **Copy From System** to add the new entries from the system allowlist to the project allowlist. 

{{< note >}}
If CVEs are deleted from the system allowlist after you have created a project allowlist, and if you added the system allowlist to the project allowlist, you must manually remove the deleted CVEs from the project allowlist. If you click **Copy From System** after CVEs have been deleted from the system allowlist, the deleted CVEs are not automatically removed from the project allowlist.
{{< /note >}}
