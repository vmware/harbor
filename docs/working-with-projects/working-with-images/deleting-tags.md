---
title: Detagging Artifacts
weight: 75
---

Harbor v2.0 now supports OCI images and OCI image indexes (https://github.com/opencontainers/image-spec/blob/master/image-index.md). An OCI image index (or OCI index) is a higher level manifest which points to a list of image manifests, ideal for one or more platforms.  Both the index itself and the images referenced within are referred to as artifacts in Harbor parlance. An OCI index could hold another OCI index and so on and so forth.  For any artifact referenced by an OCI index, the referenced artifact is known as the child artifact and the OCI index referencing the artifact is known as the parent artifact.  We can also say that the child artifact belongs to the parent artifact or is a part of the parent artifact.  

Users can delete any existing tag from an artifact without deleting the artifact digest and all other existing tags. For an OCI index, users can delete tags from the parent as well as from the referenced artifacts within. Tags removed from the parent artifact are not automatically removed from children artifacts. For example, you can tag artifacts as follows:

In the Harbor interface, click on an artifact to see its current set of tags, then select the tag you wish to delete and click 'REMOVE TAG', and then click 'OK'

![delete tag](../../../img/deletetag1.png)

You can remove all tags from an artifact without deleting the artifact manifest itself.  The artifact is still visible on the web console with nothing listed under 'Tags '