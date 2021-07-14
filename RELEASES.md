# Versioning and Release
This document describes the versioning and release process of Harbor. This document is a living document, contents will be updated according to each releases.

## Releases
Harbor releases will be versioned using dotted triples, similar to [Semantic Version](http://semver.org/). For this specific document, we will refer to the respective components of this triple as `<major>.<minor>.<patch>`. The version number may have additional information, such as "-rc1,-rc2,-rc3" to mark release candidate builds for earlier access. Such releases will be considered as "pre-releases".

### Major and Minor Releases
Major and minor releases of Harbor will be branched from master when the release reaches to `RC(release candidate)` state. The branch format should follow `release-<major>.<minor>.0`. For example, once the release `v1.0.0` reaches to RC, a branch will be created with the format `release-1.0.0`. When the release reaches to `GA(General Available)` state, The tag with format `v<major>.<minor>.<patch>` and should be made with command `git tag -s v<major>.<minor>.<patch>`. The release cadence is around 3 months, might be adjusted based on open source event, but will communicate it clearly.

### Patch releases
Patch releases are based on the major/minor release branch, the release cadence for patch release of recent minor release is one month to solve critical community and security issues. The cadence for patch release of recent minus two minor releases are on-demand driven based on the severity of the issue to be fixed.

### Pre-releases
`Pre-releases:mainly the different RC builds` will be compiled from their corresponding branches. Please note they are done to assist in the stabilization process, no guarantees are provided.

### Minor Release Support Matrix
| Version | Supported          |
| ------- | ------------------ |
| Harbor v2.3.x   | :white_check_mark: |
| Harbor v2.2.x   | :white_check_mark: |
| Harbor v2.1.x   | :white_check_mark: |

### Upgrade path and support policy
The upgrade path for Harbor is (1) 2.2.x patch releases are always compatible with its major and minor version. For example, previous released 2.2.x can be upgraded to most recent 2.2.3 release. (2) Harbor only supports two previous minor releases to upgrade to current minor release. For example, 2.3.0 will only support 2.1.0 and 2.2.0 to upgrade from, 2.0.0 to 2.3.0 is not supported. One should upgrade to 2.2.0 first, then to 2.3.0.
The Harbor project maintains release branches for the three most recent minor releases, each minor release will be maintained for approximately 9 months.

### Next Release
The activity for next release will be tracked in the [up-to-date project board](https://github.com/orgs/goharbor/projects/1). If your issue is not present in the corresponding release, please reach out to the maintainers to add the issue to the project board.

### Publishing a New Release

The following steps outline what to do when its time to plan for and publish a release. Depending on the release (major/minor/patch), not all the following items are needed.

1. Prepare information about what's new in the release.
  * For every release, update documentation for changes that have happened in the release. See the [goharbor/website](https://github.com/goharbor/website) repo for more details on how to create documentation for a release. All documentation for a release should be published by the time the release is out.
  * For every release, write release notes. See [previous releases](https://github.com/goharbor/harbor/releases) for examples of what to included in release notes.
  * For a major/minor release, write a blog post that highlights new features in the release. Plan to publish this the same day as the release. Highlight the themes, or areas of focus, for the release. Some examples of themes are security, bug fixes, feature improvements. If there are any new features or workflows introduced in a release, consider writing additional blog posts to help users learn about the new features. Plan to publish these after the release date (all blogs don’t have to be publish all at once).
1. Release a new version. Make the new version, docs updates, and blog posts available.
1. Announce the release and thank contributors. We should be doing the following for all releases.
  * In all messages to the community include a brief list of highlights and links to the new release blog, release notes, or download location. Also include shoutouts to community member contribution included in the release.
  * Send an email to the community via the [mailing list](https://lists.cncf.io/g/harbor-users)
  * Post a message in the Harbor [slack channel](https://cloud-native.slack.com/archives/CC1E09J6S)
  * Post to social media. Maintainers are encouraged to also post or repost from the Harbor account to help spread the word.
