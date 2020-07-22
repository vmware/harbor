---
title: Tagging Artifacts
weight: 75
---

Harbor v2.0 now supports OCI images and OCI image indexes (https://github.com/opencontainers/image-spec/blob/master/image-index.md). An OCI image index (or OCI index) is a higher level manifest which points to a list of image manifests, ideal for one or more platforms.  Both the index itself and the images referenced within are referred to as artifacts in Harbor parlance. An OCI index could hold another OCI index and so on and so forth.  For any artifact referenced by an OCI index, the referenced artifact is known as the child artifact and the OCI index referencing the artifact is known as the parent artifact.  We can also say that the child artifact belongs to the parent artifact or is a part of the parent artifact.  

Users can add as many tags to any artifact as they wish without impacting the artifact digest or the associated storage. For an OCI index, users can add tags to the parent as well as add tags to the individual referenced artifacts within. Tags added to the parent artifact are not automatically inherited by the children artifacts. You can tag artifacts on the Harbor web console as follows:

In the Harbor interface, click on an artifact to see its current set of tags, then click 'ADD TAG', specify the name and click 'OK'

![add artifact](../../../img/addtag1.png)

