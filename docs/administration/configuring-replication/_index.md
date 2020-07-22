---
title: Configuring Replication
weight: 30
---

Replication allows users to replicate resources, namely images and charts, between Harbor and non-Harbor registries, in both pull or push mode. 

When the Harbor system administrator has set a replication rule, all resources that match the defined filter patterns are replicated to the destination registry when the triggering condition is met. Each resource that is replicated starts a replication task. If the namespace does not exist in the destination registry, a new namespace is created automatically. If it already exists and the user account that is configured in the replication policy does not have write privileges in it, the process fails. Member information is not replicated.  

There might be some delay during replication based on the condition of the network. If a replication task fails, it is re-scheduled for a few minutes later and retried several times.  

{{< note >}}
Due to API changes, replication between different versions of Harbor is not supported.
{{< /note >}}