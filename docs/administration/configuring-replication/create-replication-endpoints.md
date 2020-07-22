---
title: Creating Replication Endpoints
weight: 20
---

To replicate image repositories from one instance of Harbor to another Harbor or non-Harbor registry, you first create replication endpoints.

1. Go to **Registries** and click the **+ New Endpoint** button.

   ![New replication endpoint](../../../img/replication-endpoint1.png)

1. For **Provider**, use the drop-down menu to select the type of registry to set up as a replication endpoint.

   The endpoint can be another Harbor instance, or a non-Harbor registry. Currently, the following non-Harbor registries are supported:

   - Docker Hub
   - Docker registry
   - AWS Elastic Container Registry
   - Azure Container Registry
   - Ali Cloud Container Registry
   - Google Container Registry
   - Huawei SWR
   - Helm Hub
   - Gitlab
   - Quay.io
   - Jfrog Artifactory

   ![Replication providers](../../../img/replication-endpoint2.png)

1. Enter a suitable name and description for the new replication endpoint.
1. Enter the full URL of the registry to set up as a replication endpoint.

   For example, to replicate to another Harbor instance, enter https://harbor_instance_address:443. The registry must exist and be running before you create the endpoint.

1. Enter the Access ID and Access Secret for the endpoint registry instance.

   Use an account that has the appropriate privileges on that registry, or an account that has write permission on the corresponding project in a Harbor registry.

   {{< note >}}
   - AWS ECR adapters should use access keys, not a username and password. The access key should have sufficient permissions, such as storage permission.
   - Google GCR adapters should use the entire JSON key generated in the service account. The namespace should start with the project ID.
   {{< /note >}}

1. Optionally, select the **Verify Remote Cert** check box.

   Deselect the check box if the remote registry uses a self-signed or untrusted certificate.

1. Click **Test Connection**.
1. When you have successfully tested the connection, click **OK**.

## Managing Registries  

You can list, add, edit and delete registries under **Administration** -> **Registries**. Only registries which are not referenced by any rules can be deleted.  

![browse project](../../../img/manage-registry.png)

## What to Do Next

After you configure replication endpoints, see [Creating a Replication Rule](create-replication-rules.md).
