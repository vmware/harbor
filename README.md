# Harbor

[![CI](https://github.com/goharbor/harbor/workflows/CI/badge.svg?branch=master&event=push)](https://github.com/goharbor/harbor/actions?query=event%3Apush+branch%3Amaster+workflow%3ACI+)
[![Coverage Status](https://codecov.io/gh/goharbor/harbor/branch/master/graph/badge.svg)](https://codecov.io/gh/goharbor/harbor)
[![Go Report Card](https://goreportcard.com/badge/github.com/goharbor/harbor)](https://goreportcard.com/report/github.com/goharbor/harbor)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2095/badge)](https://bestpractices.coreinfrastructure.org/projects/2095)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/c8d726c9cfd047ffaf681449d673f246)](https://www.codacy.com/app/goharbor/harbor?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=goharbor/harbor&amp;utm_campaign=Badge_Grade)
![Code scanning - action](https://github.com/goharbor/harbor/workflows/Code%20scanning%20-%20action/badge.svg)
[![Nightly Status](https://us-central1-eminent-nation-87317.cloudfunctions.net/harbor-nightly-result)](https://www.googleapis.com/storage/v1/b/harbor-nightly/o)
![CONFORMANCE_TEST](https://github.com/goharbor/harbor/workflows/CONFORMANCE_TEST/badge.svg)
</br>

|![notification](https://raw.githubusercontent.com/goharbor/website/master/docs/img/readme/bell-outline-badged.svg)Community Meeting|
|------------------|
|The Harbor Project holds bi-weekly community calls in two different timezones. To join the community calls or to watch previous meeting notes and recordings, please visit the [meeting schedule](https://github.com/goharbor/community/blob/master/MEETING_SCHEDULE.md).|

</br> </br>

**Note**: The `master` branch may be in an *unstable or even broken state* during development.
Please use [releases](https://github.com/vmware/harbor/releases) instead of the `master` branch in order to get a stable set of binaries.

<img alt="Harbor" src="https://raw.githubusercontent.com/goharbor/website/master/docs/img/readme/harbor_logo.png">

Harbor is an open source trusted cloud native registry project that stores, signs, and scans content. Harbor extends the open source Docker Distribution by adding the functionalities usually required by users such as security, identity and management. Having a registry closer to the build and run environment can improve the image transfer efficiency. Harbor supports replication of images between registries, and also offers advanced security features such as user management, access control and activity auditing.

Harbor is hosted by the [Cloud Native Computing Foundation](https://cncf.io) (CNCF). If you are an organization that wants to help shape the evolution of cloud native technologies, consider joining the CNCF. For details about who's involved and how Harbor plays a role, read the CNCF
[announcement](https://www.cncf.io/blog/2018/07/31/cncf-to-host-harbor-in-the-sandbox/).

## Features

* **Cloud native registry**: With support for both container images and [Helm](https://helm.sh) charts, Harbor serves as registry for cloud native environments like container runtimes and orchestration platforms.
* **Role based access control**: Users access different repositories through 'projects' and a user can have different permission for images or Helm charts under a project.
* **Policy based replication**: Images and charts can be replicated (synchronized) between multiple registry instances based on policies with using filters (repository, tag and label). Harbor automatically retries a replication if it encounters any errors. This can be used to assist loadbalancing, achieve high availability, and facilitate multi-datacenter deployments in hybrid and multi-cloud scenarios.
* **Vulnerability Scanning**: Harbor scans images regularly for vulnerabilities and has policy checks to prevent vulnerable images from being deployed.
* **LDAP/AD support**: Harbor integrates with existing enterprise LDAP/AD for user authentication and management, and supports importing LDAP groups into Harbor that can then be given permissions to specific projects.  
* **OIDC support**: Harbor leverages OpenID Connect (OIDC) to verify the identity of users authenticated by an external authorization server or identity provider. Single sign-on can be enabled to log into the Harbor portal.  
* **Image deletion & garbage collection**: System admin can run garbage collection jobs so that images(dangling manifests and unreferenced blobs) can be deleted and their space can be freed up periodically.
* **Notary**: Support signing container images using Docker Content Trust (leveraging Notary) for guaranteeing authenticity and provenance.  In addition, policies that prevent unsigned images from being deployed can also be activated.
* **Graphical user portal**: User can easily browse, search repositories and manage projects.
* **Auditing**: All the operations to the repositories are tracked through logs.
* **RESTful API**: RESTful APIs are provided to facilitate administrative operations, and are easy to use for integration with external systems. An embedded Swagger UI is available for exploring and testing the API.
* **Easy deployment**: Harbor can be deployed via Docker compose as well Helm Chart, and a Harbor Operator was added recently as well.

## Architecture

For learning the architecture design of Harbor, check the document [Architecture Overview of Harbor](https://github.com/goharbor/harbor/wiki/Architecture-Overview-of-Harbor).

## API

* Harbor RESTful API: The APIs for most administrative operations of Harbor and can be used to perform integrations with Harbor programmatically.
  * Part 1: [New or changed APIs](https://editor.swagger.io/?url=https://raw.githubusercontent.com/goharbor/harbor/master/api/v2.0/swagger.yaml)
  * Part 2: [Legacy APIs](https://editor.swagger.io/?url=https://raw.githubusercontent.com/goharbor/harbor/master/api/v2.0/legacy_swagger.yaml)

## Install & Run

**System requirements:**

**On a Linux host:** docker 17.06.0-ce+ and docker-compose 1.18.0+ .

Download binaries of **[Harbor release ](https://github.com/vmware/harbor/releases)** and follow **[Installation & Configuration Guide](https://goharbor.io/docs/2.0.0/install-config/)** to install Harbor.

If you want to deploy Harbor on Kubernetes, please use the **[Harbor chart](https://github.com/goharbor/harbor-helm)**.

Refer to the **[documentation](https://goharbor.io/docs/)** for more details on how to use Harbor.

## OCI Distribution Conformance Tests

Check the OCI distribution conformance tests [report](https://storage.googleapis.com/harbor-conformance-test/report.html) of Harbor.

## Compatibility

The [compatibility list](https://goharbor.io/docs/edge/install-config/harbor-compatibility-list/) document provides compatibility information for the Harbor components.

* [Replication adapters](https://goharbor.io/docs/edge/install-config/harbor-compatibility-list/#replication-adapters)
* [OIDC adapters](https://goharbor.io/docs/edge/install-config/harbor-compatibility-list/#oidc-adapters)
* [Scanner adapters](https://goharbor.io/docs/edge/install-config/harbor-compatibility-list/#scanner-adapters)

## Community

* **Twitter:** [@project_harbor](https://twitter.com/project_harbor)  
* **User Group:** Join Harbor user email group: [harbor-users@lists.cncf.io](https://lists.cncf.io/g/harbor-users) to get update of Harbor's news, features, releases, or to provide suggestion and feedback.  
* **Developer Group:** Join Harbor developer group: [harbor-dev@lists.cncf.io](https://lists.cncf.io/g/harbor-dev) for discussion on Harbor development and contribution.
* **Slack:** Join Harbor's community for discussion and ask questions: [Cloud Native Computing Foundation](https://slack.cncf.io/), channel: [#harbor](https://cloud-native.slack.com/messages/harbor/) and [#harbor-dev](https://cloud-native.slack.com/messages/harbor-dev/)

## Community contributed CLI

### Harbor CLI

A community contributed [Unofficial Harbor CLI](https://github.com/hinyinlam-pivotal/harbor-cli) can be downloaded to interact with Harbor in system level and repo level.
The `harbor-cli` is built on top of `harbor-client-java`, tutorial and usage examples also included in the author's repo.

For any feedback, improvement and issues in Java SDK and CLI, please contact the [author](https://github.com/hinyinlam-pivotal).

## Demos

* **[Live Demo](https://demo.goharbor.io)** - A demo environment with the latest Harbor stable build installed. For additional information please refer to [this page](https://goharbor.io/docs/2.0.0/install-config/demo-server/).
* **[Video Demos](https://github.com/goharbor/harbor/wiki/Video-demos-for-Harbor)** - Demos for Harbor features and continuously updated.

## Partners and Users

For a list of users, please refer to [ADOPTERS.md](ADOPTERS.md).

## Security

### Security Audit

A third party security audit was performed by Cure53 in October of 2019. You can see the full report [here](https://goharbor.io/docs/2.0.0/security/Harbor_Security_Audit_Oct2019.pdf).

### Reporting security vulnerabilities

If you've found a security related issue, a vulnerability, or a potential vulnerability in Harbor please let the [Harbor Security Team](mailto:cncf-harbor-security@lists.cncf.io) know with the details of the vulnerability. We'll send a confirmation
email to acknowledge your report, and we'll send an additional email when we've identified the issue
positively or negatively.

For further details please see our complete [security release process](SECURITY.md).

## License

Harbor is available under the [Apache 2 license](LICENSE).

This project uses open source components which have additional licensing terms.  The official docker images and licensing terms for these open source components can be found at the following locations:

* Photon OS 1.0: [docker image](https://hub.docker.com/_/photon/), [license](https://github.com/vmware/photon/blob/master/COPYING)
