---
title: Creating a Replication Rule
weight: 25
---

A replication endpoint must exist before you create a replication rule. To create an endpoint, follow the instructions in [Creating Replication Endpoints](create-replication-endpoints.md).

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Expand **Administration**, and select **Replications**.

   ![Add a replication rule](../../../img/replication-rule1.png)
1. Click **New Replication Rule**.
1. Provide a name and description for the replication rule.
1. Select **Push-based** or **Pull-based** replication, depending on whether you want to replicate artifacts to or from the remote registry.

   ![Replication mode](../../../img/replication-rule2.png)
1. For **Source resource filter**, identify the artifacts to replicate.  

   ![Replication filters](../../../img/replication-rule3.png)

   * **Name**: Replicate resources with a given name by entering an artifact name or fragment.
   * **Tag**: Replicate resources with a given tag by entering a tag name or fragment.
   * **Label**: Replicate resources with a given label by using the drop-down menu to select from the available labels.
   * **Resource**: Replicate artifacts, charts, or both.

   The name filter and tag filters support the following patterns:

   * **\***: Matches any sequence of non-separator characters `/`.
   * **\*\***: Matches any sequence of characters, including path separators `/`. Note that the doublestar must appear as a path component by itself. A pattern such as /path\*\* is invalid and will be treated the same as /path*, but /path\*/\*\* should achieve the desired result.
   * **?**: Matches any single non-separator character `/`.
   * **{alt1,...}**: Matches a sequence of characters if one of the comma-separated alternatives matches.

   **NOTE:** You must add `library` if you want to replicate the official artifacts of Docker Hub. For example, `library/hello-world` matches the official hello-world artifacts.  

   Pattern | String(Match or not)
   ---------- | -------
   `library/*`      | `library/hello-world`(Y)<br> `library/my/hello-world`(N)
   `library/**`     | `library/hello-world`(Y)<br> `library/my/hello-world`(Y)
   `{library,goharbor}/**` | `library/hello-world`(Y)<br> `goharbor/harbor-core`(Y)<br> `google/hello-world`(N)
   `1.?`      | `1.0`(Y)<br> `1.01`(N)
1. Use the **Destination Registry** drop-down menu to select from the configured replication endpoints. 
1. Enter the name of the namespace in which to replicate resources in the **Destination namespace** text box.

   If you do not enter a namespace, resources are placed in the same namespace as in the source registry. 

   ![Destination and namespaces](../../../img/replication-rule4.png)
   
   **NOTE:** Because of major API changes in the v2.0 release to support [OCI](https://github.com/opencontainers/distribution-spec).
   You **can not** replicate from harbor 1.x to 2.0, and you **can not** replicate artifacts with **manifest list** from 2.0 to 1.x. 
   
1. Use the Trigger Mode drop-down menu to select how and when to run the rule.
   * **Manual**: Replicate the resources manually when needed. **Note**: Deletion operations are not replicated. 
   * **Scheduled**: Replicate the resources periodically by defining a cron job. **Note**: Deletion operations are not replicated. 
   * **Event Based**: When a new resource is pushed to the project, or an artifact is retagged, it is replicated to the remote registry immediately. If you select the **Delete remote resources when locally deleted**, if you delete an artifact, it is automatically deleted from the replication target.

   {{< note >}}
   You can filter artifacts for replication based on the labels that are applied to the artifacts. However, changing a label on an artifact does not trigger replication. Event-based replication is limited to pushing, retagging, and deleting artifacts.
   {{< /note >}}

   ![Trigger mode](../../../img/replication-rule5.png)
      
1. Optionally select the Override checkbox to force replicated resources to replace resources at the destination with the same name.
1. Click **Save** to create the replication rule.  

## What to Do Next

After you create a replication rule, see [Running Replication Manually](manage-replications.md).
