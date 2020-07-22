---
title: Repositories
weight: 75
---

A repository is a collection of artifacts.  Since version v2.0, in addition to container images, Harbor can manage different kinds of artifacts that are bundled in OCI-compatible format, such as helm chart (requires helm v3), CNAB, OPA bundle, etc.

### List Repositories

Click your project to enter the project detail page after successful logging in.  Click the "Repositories" tab to view the list of of repositories. 

![list_repositories](../../../img/list-repositories.png)

### Description of a repository

Click the repository, then click the "Info" tab.  You can view the description of the project.  Users with project admin, master or developer role can click the "Edit" button to edit the description.  You can style the description via Markdown syntax.

![edit_repository_description](../../../img/edit-repository-description.png)

### List artifacts in a repository

Click the "Artifacts" tab to view the list of artifacts in a repository.
Each artifact is identified by its sha256 digest in the list of artifacts, and different types of artifacts can be distinguished by the icon on the left of the digest.  Hover your mouse on the icon you can see the name of the type.  

By clicking the icon in the column **Pull Command**, the command to pull the artifact in the row of the icon will be copied to the clipboard.  
The column **Annotations** in the grid shows the manifest annotations of the artifact, which are a set of key-value pairs.  More details about the annotations please refer to (https://github.com/opencontainers/image-spec/blob/master/annotations.md).
The column **Push Time** in the grid shows the time each artifact is pushed to the registry.

![list_artifacts](../../../img/list-artifacts.png)

By clicking the search icon in the top right of the list of artifacts, you can user different types of filters to filter the items in the artifact list.  You can choose to filter by type, tags, labels.  Particularly, if you choose to filter by tags, you can choose to view only the tagged or untagged artifacts.

![filter_artifacts](../../../img/filter-artifacts.png)

Since Harbor v2.0.0, [Image index](https://raw.githubusercontent.com/opencontainers/image-spec/master/image-index.md) can also be managed as an artifact in a repository.  If an artifact is an index, there will be a folder icon on the right side of its digest.

![image_index](../../../img/index-icon.png)

Click the folder icon, you can see the list of artifacts that is referenced by the index.  The artifacts in this view is read only.  i.e. You can not remove an artifact from an index via Harbor's UI, and none of the actions like 'copy digest', 'add labels', 'copy' are available.

![index_detail](../../../img/index-detail.png)
