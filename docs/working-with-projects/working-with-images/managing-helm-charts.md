---
title: Managing Helm Charts
weight: 95
---

[Helm](https://helm.sh) is a package manager for [Kubernetes](https://kubernetes.io). Helm uses a packaging format called [charts](https://docs.helm.sh/developing_charts). Since version 1.6.0 Harbor is now a composite cloud-native registry which supports both container image management and Helm charts management. Access to Helm charts in Harbor is controlled by [role-based access controls (RBAC)](https://en.wikipedia.org/wiki/Role-based_access_control) and is restricted by projects.

There are two places to manage helm charts. First one is in the ChartMuseum which is provided by Harbor from version 1.6.0. The second one is in the OCI-compatible registry which is provided by Harbor from version 2.0.0. This means you can now manage Helm charts alongside your container images through the same set of projects and repositories.

## Manage Helm Charts with the ChartMuseum in Harbor Interface

### List charts

Click your project to enter the project detail page after successful logging in. The existing helm charts will be listed under the tab `Helm Charts` which is beside the image `Repositories` tab with the following information:
* Name of helm chart
* The status of the chart: Active or Deprecated
* The count of chart versions
* The created time of the chart

![list charts](../../../img/list-charts.png)

You can click the icon buttons on the top right to switch views between card view and list view.

### Upload a New Chart

Click the `UPLOAD` button on the top left to open the chart uploading dialog. Choose the uploading chart from your filesystem. Click the `UPLOAD` button to upload it to the chart repository server.

![upload charts](../../../img/upload-charts.png)

If the chart is signed, you can choose the corresponding provenance file from your filesystem and Click the `UPLOAD` button to upload them together at once.

If the chart is successfully uploaded, it will be displayed in the chart list at once.

### List Chart Versions

Clicking the chart name from the chart list will show all the available versions of that chart with the following information:
* the chart version number
* the maintainers of the chart version
* the template engine used (default is gotpl)
* the created timestamp of the chart version

![list charts versions](../../../img/list-chart-versions.png)

Obviously, there will be at least 1 version for each of the charts in the top chart list. Same with chart list view, you can also click the icon buttons on the top right to switch views between card view and list view.

Check the checkbox at the 1st column to select the specified chart versions:
* Click the `DELETE` button to delete all the selected chart versions from the chart repository server. Batch operation is supported.
* Click the `DOWNLOAD` button to download the chart artifact file. Batch operation is not supported.
* Click the `UPLOAD` button to upload the new chart version for the current chart

### Adding Labels to and Removing Labels from Chart Versions
Users who have Harbor system administrator, project administrator or project developer role can click the `ADD LABELS` button to add labels to or remove labels from chart versions.

![add labels to chart versions](../../../img/add-labels-to-chart-versions.png)


### Filtering Chart Versions by Label
The chart versions can be filtered by labels:

![filter chart versions by labels](../../../img/filter-chart-versions-by-label.png)

### View Chart Version Details
Clicking the chart version number link will open the chart version details view. You can see more details about the specified chart version here. There are three content sections:
* **Summary:**
  * readme of the chart
  * overall metadata like home, created timestamp and application version
  * related helm commands for reference, such as `helm add repo` and `helm install` etc.
![chart details](../../../img/chart-details.png)
* **Dependencies:**
  * list all the dependant sun charts with 'name', 'version' and 'repository' fields
![chart dependencies](../../../img/chart-dependencies.png)
* **Values:**
  * display the content from `values.yaml` file with highlight code preview
  * clicking the icon buttons on the top right to switch the yaml file view to k-v value pair list view
![chart values](../../../img/chart-values.png)

Clicking the `DOWNLOAD` button on the top right will start the downloading process.

## Working with ChartMuseum via the Helm CLI

As a helm chart repository, Harbor can interoperate with Helm CLI. To install Helm CLI, please refer [install helm](https://helm.sh/docs/intro/install/). Run command `helm version` to make sure the version of Helm CLI is v2.9.1+.

```sh
helm version

#Client: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
#Server: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
```

### Add Harbor to the Repository List

Before working, Harbor should be added into the repository list with `helm repo add` command. Two different modes are supported.

* Add Harbor as a unified single index entry point

    With this mode Helm can be made aware of all the charts located in different projects and which are accessible by the currently authenticated user.

    ```sh
    helm repo add --ca-file ca.crt --username=admin --password=Passw0rd myrepo https://xx.xx.xx.xx/chartrepo
    ```

   {{< note >}}
   Providing both a CA file and cert files is necessary due to an issue in Helm.
   {{< /note >}}

* Add Harbor project as separate index entry point

    With this mode, Helm can only pull charts in the specified project.

    ```sh
    helm repo add --ca-file ca.crt --username=admin --password=Passw0rd myrepo https://xx.xx.xx.xx/chartrepo/myproject
    ```

### Push Charts to the Repository Server with the CLI

As an alternative, you can also upload charts via the CLI. It is not supported by the native helm CLI. A plugin from the community should be installed before pushing. Run `helm plugin install` to install the `push` plugin first.

```sh
helm plugin install https://github.com/chartmuseum/helm-push
```

After a successful installation, run the `push` command to upload your charts:

```sh
helm push --ca-file=ca.crt --username=admin --password=passw0rd chart_repo/hello-helm-0.1.0.tgz myrepo
```

{{< note >}}
The `push` command does not yet support pushing a prov file of a signed chart.
{{< /note >}}

### Install Charts

Before installing, make sure your helm is correctly initialized with command `helm init` and the chart index is synchronized with command `helm repo update`.

Search the chart with the keyword if you're not sure where it is:

```sh
helm search hello

#NAME                            CHART VERSION   APP VERSION     DESCRIPTION
#local/hello-helm                0.3.10          1.3             A Helm chart for Kubernetes
#myrepo/chart_repo/hello-helm    0.1.10          1.2             A Helm chart for Kubernetes
#myrepo/library/hello-helm       0.3.10          1.3             A Helm chart for Kubernetes
```

If everything is ready, install the chart in your Kubernetes cluster:

```sh
helm install --ca-file=ca.crt --username=admin --password=Passw0rd --version 0.1.10 repo248/chart_repo/hello-helm
```

For other more helm commands like how to sign a chart, please refer to the [helm doc](https://docs.helm.sh/helm/#helm).


## Manage Helm Charts with the OCI-compatible registry of Harbor

Helm 3 now supports registry operations for an OCI-compatible registry including pushing and pulling. To install the latest Helm CLI, please refer [install helm](https://helm.sh/docs/intro/install/). Please also run `helm version` command to make sure the version of Helm CLI is v3.0.0+.

```sh
helm version

#version.BuildInfo{Version:"v3.2.1", GitCommit:"fe51cd1e31e6a202cba7dead9552a6d418ded79a", GitTreeState:"clean", GoVersion:"go1.13.10"}
```

### Login to the OCI-compatible registry of Harbor

Before pull/push helm charts with the OCI-compatible registry of Harbor, Harbor should be logged with `helm registry login` command.

```sh
helm registry login xx.xx.xx.xx
```

{{< note >}}
The CA file used by the Harbor is necessary to be trusted in the system due to an [issue](https://github.com/helm/helm/issues/6324) in Helm.
{{< /note >}}

### Push Charts to the artifact Repository with the CLI

After logging in, run the `helm chart save` command to save a chart directory which will prepare the artifact for the pushing.

```sh
helm chart save dummy-chart xx.xx.xx.xx/library/dummy-chart
```


When the chart was saved run the `helm chart push` command to push your charts:

```sh
helm chart push xx.xx.xx.xx/library/dummy-chart:version
```

### Pull Charts from the artifact Repository with the CLI

To pull charts from the the OCI-compatible registry of Harbor, run the `helm chart pull` command just like pulling image via docker cli.

```sh
helm chart pull xx.xx.xx.xx/library/dummy-chart:version
```

### Manage Helm Charts artifacts in Harbor Interface

The charts pushed to the OCI-compatible registry of Harbor are treated like any other type of artifact. We can list, copy, delete, update labels, get details, add or remove tags for them just like we can for container images.

![chart artifact details](../../../img/chart-artifact-details.png)