---
title: Running Replication Manually
weight: 30
---

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Expand **Administration**, and select **Replications**.
1. Select a replication rule and click **Replicate**. 

    ![Add a replication rule](../../../img/replication-rule6.png)

    The resources to which the rule is applied start to replicate from the source registry to the destination immediately.     
1. Click the rule to see its execution status.
1. Click the **ID** of the execution to see the details of the replication  and the task list. The count of `IN PROGRESS` status in the summary includes both `Pending` and `In Progress` tasks.  
1. Optionally click **STOP** to stop the replication. 
1. Click the log icon to see detailed information about the replication task. 

    ![View replication task](../../../img/list-tasks.png)

To edit or delete a replication rule, select the replication rule in the **Replications** view and click **Edit** or **Delete**. Only rules which have no executions in progress can be edited deleted.  

![Delete or edit rule](../../../img/replication-rule6.png)
