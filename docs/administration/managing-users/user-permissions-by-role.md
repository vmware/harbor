---
title: User Permissions By Role
weight: 20
---

Users have different abilities depending on the role they in a project.

On public projects all users will be able to see the list of repositories, images, image vulnerabilities, helm charts and helm chart versions, pull images, retag images (need push permission for destination image), download helm charts, download helm chart versions.

System admin have all permissions for the project.

## Project members permissions

The following table depicts the various user permission levels in a project.

| Action                                  | Limited Guest | Guest | Developer | Master | Project Admin |
| --------------------------------------- | ------------- | ----- | --------- | ------ | ------------- |
| See the project configurations          | ✓             | ✓     | ✓         | ✓      | ✓             |
| Edit the project configurations         |               |       |           |        | ✓             |
| See a list of project members           |               | ✓     | ✓         | ✓      | ✓             |
| Create/edit/delete project members      |               |       |           |        | ✓             |
| See a list of project logs              |               | ✓     | ✓         | ✓      | ✓             |
| See a list of project replications      |               |       |           | ✓      | ✓             |
| See a list of project replication jobs  |               |       |           |        | ✓             |
| See a list of project labels            |               |       |           | ✓      | ✓             |
| Create/edit/delete project labels       |               |       |           | ✓      | ✓             |
| See a list of repositories              | ✓             | ✓     | ✓         | ✓      | ✓             |
| Create repositories                     |               |       | ✓         | ✓      | ✓             |
| Edit/delete repositories                |               |       |           | ✓      | ✓             |
| See a list of images                    | ✓             | ✓     | ✓         | ✓      | ✓             |
| Retag image                             |               | ✓     | ✓         | ✓      | ✓             |
| Pull image                              | ✓             | ✓     | ✓         | ✓      | ✓             |
| Push image                              |               |       | ✓         | ✓      | ✓             |
| Scan/delete image                       |               |       |           | ✓      | ✓             |
| Add scanners to Harbor                  |               |       |           |        |               |
| Edit scanners in projects               |               |       |           |        | ✓             |
| See a list of image vulnerabilities     | ✓             | ✓     | ✓         | ✓      | ✓             |
| See image build history                 | ✓             | ✓     | ✓         | ✓      | ✓             |
| Add/Remove labels of image              |               |       | ✓         | ✓      | ✓             |
| See a list of helm charts               | ✓             | ✓     | ✓         | ✓      | ✓             |
| Download helm charts                    | ✓             | ✓     | ✓         | ✓      | ✓             |
| Upload helm charts                      |               |       | ✓         | ✓      | ✓             |
| Delete helm charts                      |               |       |           | ✓      | ✓             |
| See a list of helm chart versions       | ✓             | ✓     | ✓         | ✓      | ✓             |
| Download helm chart versions            | ✓             | ✓     | ✓         | ✓      | ✓             |
| Upload helm chart versions              |               |       | ✓         | ✓      | ✓             |
| Delete helm chart versions              |               |       |           | ✓      | ✓             |
| Add/Remove labels of helm chart version |               |       | ✓         | ✓      | ✓             |
| See a list of project robots            |               |       |           | ✓      | ✓             |
| Create/edit/delete project robots       |               |       |           |        | ✓             |
| See configured CVE allowlist            | ✓             | ✓     | ✓         | ✓      | ✓             |
| Create/edit/remove CVE allowlist        |               |       |           |        | ✓             |
| View webhook events                     |               |       |           | ✓      | ✓             |
| Add new webhook events                  |               |       |           |        | ✓             |
| Enable/disable webhooks                 |               |       |           |        | ✓             |
| Create/delete tag retention rules       |               |       | ✓         | ✓      | ✓             |
| Enable/disable tag retention rules      |               |       | ✓         | ✓      | ✓             |
| Create/delete tag immutability rules    |               |       |           |        | ✓             |
| Enable/disable tag immutability rules   |               |       |           |        | ✓             |
| See project quotas                      | ✓             | ✓     | ✓         | ✓      | ✓             |
| Edit project quotas  *                  |               |       |           |        |               |

&ast; Only the Harbor system administrator can edit project quotas and add new scanners.
