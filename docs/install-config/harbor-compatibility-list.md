---
title: Harbor Compatibility List
weight: 15
---

This document provides compatibility information for all Harbor components.

## Replication Adapters

|     | Registries       | Pull Mode | Push Mode | Introduced in Release | Automated Pipeline Covered |
|-----|------------------|-----------|-----------|-----------------------|---------------------------|
| [Harbor](https://goharbor.io/)|  ![Harbor](../../img/replication-adapters/harbor-logo.png)|![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| V1.8 | Y |
| [distribution](https://github.com/docker/distribution) | ![distribution](../../img/replication-adapters/distribution.png)|![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| V1.8 | Y |
| [docker hub](https://hub.docker.com/) | ![docker hub](../../img/replication-adapters/docker-hub.png)|![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| V1.8 | Y |
| [Huawei SWR](https://www.huaweicloud.com/en-us/product/swr.html) | ![Huawei SWR](../../img/replication-adapters/hw.png)|![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| V1.8 | N |
| [GCR](https://cloud.google.com/container-registry/) | ![GCR](../../img/replication-adapters/gcr.png)|![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| V1.9 | Y |
| [ECR](https://aws.amazon.com/ecr/) | ![ECR](../../img/replication-adapters/ecr.png)|![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| V1.9 | Y |
| [ACR](https://azure.microsoft.com/en-us/services/container-registry/) | ![ACR](../../img/replication-adapters/acr.png)|![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| V1.9 | N |
| [AliCR](https://www.alibabacloud.com/product/container-registry) | ![AliCR](../../img/replication-adapters/ali-cr.png)|![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| V1.9 | N |
| [Helm Hub](https://hub.helm.sh/) | ![Helm Hub](../../img/replication-adapters/helm-hub.png)|![Y](../../img/replication-adapters/right.png)| N/A | V1.9 | N |
| [Artifactory](https://jfrog.com/artifactory/) | ![Artifactory](../../img/replication-adapters/artifactory.png)|![Y](../../img/replication-adapters/right.png)| ![Y](../../img/replication-adapters/right.png) | V1.10 | N |
| [Quay](https://github.com/quay/quay) | ![Quay](../../img/replication-adapters/quay.png)|![Y](../../img/replication-adapters/right.png)| ![Y](../../img/replication-adapters/right.png) | V1.10 | N |
| [GitLab Registry](https://docs.gitlab.com/ee/user/packages/container_registry/) | ![GitLab Registry](../../img/replication-adapters/gitlab.png)|![Y](../../img/replication-adapters/right.png)| ![Y](../../img/replication-adapters/right.png) | V1.10 | N |

{{< note >}}
* `Pull` mode replicates artifacts from the specified source registries into Harbor. 
* `Push` mode replicates artifacts from Harbor to the specified target registries.
{{< /note >}}

## OIDC Adapters

|   |  OIDC Providers | Officially Verified | End User Verified   | Verified in Release |
|---|-----------------|---------------------|---------------------|-----------------------|
| [Google Identity](https://developers.google.com/identity/protocols/OpenIDConnect) | ![google identity](../../img/OIDC/google-identity.png)| ![Y](../../img/replication-adapters/right.png) |  |V1.9|
| [Dex](https://github.com/dexidp/dex) | ![dex](../../img/OIDC/dex.png) | ![Y](../../img/replication-adapters/right.png)| | V1.9 |
| [Ping Identity](https://www.pingidentity.com) | ![ping identity](../../img/OIDC/ping.png) | | ![Y](../../img/replication-adapters/right.png)| V1.9 |
| [Keycloak](https://www.keycloak.org/) | ![Keycloak](../../img/OIDC/keycloak.png) | ![Y](../../img/replication-adapters/right.png) | | V1.10 |
| [Auth0](https://auth0.com/) | ![Auth0](../../img/OIDC/auth0.png) | ![Y](../../img/replication-adapters/right.png) | | V1.10 |

## Scanner Adapters

|   | Scanners | Providers | Evaluated | As Default | Onboard in Release |
|---|----------|-----------|-----------|------------|--------------------|
| [Clair](https://github.com/goharbor/harbor-scanner-clair)    |![Clair](../../img/scanners/clair.png)| CentOS    |![Y](../../img/replication-adapters/right.png)|![Y](../../img/replication-adapters/right.png)| v1.10 |
| [Anchore](https://github.com/anchore/harbor-scanner-adapter) |![Anchore](../../img/scanners/anchore.png)   | Anchore    |![Y](../../img/replication-adapters/right.png)| N | v1.10 |
| [Trivy](https://github.com/aquasecurity/harbor-scanner-trivy)|![Trivy](../../img/scanners/trivy.png)| Aqua    |![Y](../../img/replication-adapters/right.png)| ![Y](../../img/replication-adapters/right.png) | v1.10 |
| [CSP](https://github.com/aquasecurity/harbor-scanner-aqua)   |![Aqua](../../img/scanners/aqua.png)| Aqua    | ![Y](../../img/replication-adapters/right.png) | N | v1.10 |
| [DoSec](https://github.com/dosec-cn/harbor-scanner/blob/master/README_en.md)|![DoSec](../../img/scanners/dosec.png)    | DoSec    | ![Y](../../img/replication-adapters/right.png) | N | v1.10 |

{{< note >}}
* `Evaluated` means that the scanner implementation has been officially tested and verified.
* `As Default` means that the scanner is provided as a default option and can be deployed together with the main Harbor components by providing extra options during installation. You must install other scanners manually.
{{< /note >}}
