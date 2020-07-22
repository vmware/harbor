---
title: Retagging Artifacts
weight: 75
---

User with sufficient privileges can copy artifacts in Harbor to different repositories and projects. For example, you can copy images as follows:

- `release/app:stg`  -->  `release/app:prd`
- `develop/app:v1.0` --> `release/app:v1.0`

To copy an artifact, you must have read permission (guest role or above) in the source project and write permission (developer role or above) in the target project.

In the Harbor interface, select the artifact to copy, and click `Copy`.

![retag artifact](../../../img/retag1.png)

In the Retag window, enter the project name, repository name, the new tag name, and click **Confirm**. 

![retag artifact](../../../img/retag2.png)