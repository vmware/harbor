---
title: Tag Immutability Rules
weight: 85
---

By default, users can repeatedly push an artifact with the same tag to a repository in Harbor. This causes the tag to migrate across the artifacts and every artifact that has its tag taken away becomes tagless. This is due to Docker distribution upstream which does not enforce the mapping between an image tag and the image digest. This can be undesirable in certain cases, because the tag can no longer be trusted to identify the image version. The sha256 digest remains reliable and always points to the same build, but it is not rendered in a human-readable format.

To prevent this, Harbor allows you to configure tag immutability at the project level, so that artifacts with certain tags cannot be pushed into Harbor if their tags match existing tags. This prevents existing artifacts from being overwritten. Tag immutability guarantees that an immutable tagged artifact cannot be deleted, and also cannot be altered in any way such as through re-pushing, re-tagging, or replication from another target registry.

Immutability rules use `OR` logic, so if you set multiple rules and a tag is matched by any of those rules, it is marked as immutable. 

## How Immutable Tags Prevent Tag Deletion

Since v2.0, you can delete any tag of an artifact without deleting the artifact itself. Therefore, you can lock down a particular tag by configuring an immutability rule matching this tag which means the artifact holding the tag also cannot be overwritten or deleted. However you can still delete other tags associated with this immutable artifact. Consider the follow example:

1. In the Docker client, push `hello-world:v1` into a project.
1. In the project, set an immutable tag rule in this project that matches the image and tag `hello-world:v1`.
1. Push `hello-world:v1` to the project.
1. In your local env, retag `hello-world:v1` to `hello-world:v2`.
1. Push `hello-world:v2` to the project.
1. In the Harbor interface, attempt to delete tag `v1` and `v2` of `hello-world` sequentially.

In this case, you cannot delete tag `v1` as it's an immutable tag and you cannot delete the artifact `hello-world` holding this tag. But you can delete tag `v2` even it shares the sha256 digest with `v1`. 

## Create a Tag Immutability Rule

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects**, select a project, select policy, and select **Tag Immutability**.

    ![Add an immutability rule](../../../img/tag-immutability.png)

1. Click **Add Rule**.

    - In the **Respositories** row, enter a comma-separated list of repositories to which to either apply or exclude from the rule by selecting either **matching** or **excluding** from the drop-down menu.
    - In the **Tags** row, enter a comma-separated list of tags to which to either apply or exclude from the rule by selecting either **matching** or **excluding** from the drop-down menu.
 
      ![Add an immutability rule](../../../img/add-immutability-rule.png)
1. Click **Add** to save the rule.

    You can add a maximum of 15 immutability rules per project. 

    After you add a rule, any tags that are identified by the rule are marked **Immutable** in the Repositories tab.
1. To modify an existing rule, use the **Action** drop-down menu next to a rule to disable, edit, or delete that rule. 

    ![Immutability rules](../../../img/edit-tag-immutability.png)

## Example

To make all tags for all repositories in the project immutable, set the following options:

- Set **For the respositories** to **matching** and enter `**`.
- Set **Tags** to **matching** and enter `**`.

To allow the tags `rc`, `test`, and `nightly` to be overwritten but make all other tags immutable, set the following options:

- Set **For the respositories** to **matching** and enter `**`.
- Set **Tags** to **excluding** and enter `rc,test,nightly`.
