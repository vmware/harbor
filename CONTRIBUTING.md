# Contributing to Harbor

## Welcome

Harbor is developed in the open, and is constantly being improved by our **users, contributors, and maintainers**.  
It is because of you that we can bring great software to the community.

This guide provides information on filing issues and guidelines for open source contributors. **Please leave comments / suggestions if you find something is missing or incorrect.**

Contributors are encouraged to collaborate using the following resources in addition to the GitHub [issue tracker](https://github.com/goharbor/harbor/issues):

* [Bi-weekly public community meetings][community-meetings]
  * Catch up with [past meetings on YouTube][past-meetings]
* Chat with us on the CNCF Slack ([get an invite here][cncf-slack] )
  * [#harbor][users-slack] for end-user discussions
  * [#harbor-dev][dev-slack] for development of Harbor
* Want long-form communication instead of Slack? We have two distributions lists:
  * [harbor-users][users-dl] for end-user discussions
  * [harbor-dev][dev-dl] for development of Harbor

Follow us on Twitter at [@project_harbor][twitter]

## Getting Started

### Fork Repository

Fork the Harbor repository on GitHub to your personal account.
```
#Set golang environment
export GOPATH=$HOME/go
mkdir -p $GOPATH/src/github.com/goharbor

#Get code
go get github.com/goharbor/harbor
cd $GOPATH/src/github.com/goharbor/harbor

#Track repository under your personal account
git config push.default nothing # Anything to avoid pushing to goharbor/harbor by default
git remote rename origin goharbor
git remote add $USER git@github.com:$USER/harbor.git
git fetch $USER

```
**NOTES:** Note that GOPATH can be any directory, the example above uses $HOME/go. Change $USER above to your own GitHub username.

To build the project, please refer the [build](docs/build-customize-contribute/compile-guide.md) guideline.

### Repository Structure

Here is the basic structure of the harbor code base. Some of the key folders / files are commented for your references.
```
.
...
├── contrib       # Contain documents, scripts, and other helpful things which are contributed by the community
├── docs          # Keep documents here
├── make          # Resource for building and setting up Harbor environment
...
├── src           # Source code folder
├── tests         # Test cases for API / e2e testings
└── tools         # Keep supporting tools
...
```

The folder graph below shows the structure of the source code folder `harbor/src`, which will be your primary working directory. The key folders are also commented.
```
.
├── chartserver          # Source code contains the main logic to handle chart. 
├── cmd                  # Source code contains migrate script to handle DB upgrade.
├── common              # Source code for some general components like dao etc.
│   ├── api
│   ├── config
│   ├── dao
│   ├── http
│   ├── job
│   ├── models
│   ├── rbac
│   ├── registryctl
│   ├── secret
│   ├── security
│   └── utils
├── core          # Source code for the main busines logic. Contains rest apis and all service infomation. 
│   ├── api
│   ├── auth
│   ├── config
│   ├── controllers
│   ├── filter
│   ├── label
│   ├── notifier
│   ├── promgr
│   ├── proxy
│   ├── service
│   ├── systeminfo
│   ├── utils
│   └── views
├── jobservice          # Source code for the job service component
│   ├── api
│   ├── config
│   ├── core
│   ├── env
│   ├── errs
│   ├── job
│   ├── logger
│   ├── models
│   ├── opm
│   ├── period
│   ├── pool
│   ├── runtime
│   ├── tests
│   └── utils
├── portal               # The code of harbor web UI
│   ├── e2e
│   ├── lib             # Source code of @harbor/ui npm library which includes the main UI components of web UI
│   └── src             # General web page UI code of Harbor
├── registryctl          # Source code contains the main logic to handle registry. 
├── replication          # Source code contains the main logic of replication. 
├── testing              # Some utilities to handle testing. 
└── vendor              # Go code dependencies
    ├── github.com
    ├── golang.org
    ├── google.golang.org
    └── gopkg.in
```

### Setup Development Environment

#### Go
Harbor backend is written in [Go](http://golang.org/). If you don't have a Harbor backend service development environment, please [set one up](https://golang.org/doc/install).

|  Harbor  |  Requires Go  |
|----------|---------------|
|   1.1    |    1.7.3      |
|   1.2    |    1.7.3      |
|   1.3    |    1.9.2      |
|   1.4    |    1.9.2      |
|   1.5    |    1.9.2      |
|   1.6    |    1.9.2      |
|   1.7    |    1.9.2      |
|   1.8    |    1.11.2     |
|   1.9    |    1.12.12    |
|   1.10   |    1.12.12    |
|   1.11   |    1.14.5     |

Ensure your GOPATH and PATH have been configured in accordance with the Go environment instructions.

**Dependency Management:** Harbor uses [dep](https://github.com/golang/dep) for dependency management of go code.  The official maintainers will take the responsibility for managing the code in `vendor` directory.  Please don't try to submit a PR to update the dependency code, open an issue instead.  If your PR requires a change in the vendor code please make sure you discuss it with the maintainers in advance.

#### Web

Harbor web UI is built based on [Clarity](https://vmware.github.io/clarity/) and [Angular](https://angular.io/) web framework. To setup web UI development environment, please make sure the [npm](https://www.npmjs.com/get-npm) tool is installed first.

|  Harbor  |  Requires Angular  |  Requires Clarity  |
|----------|--------------------|--------------------|
|   1.1    |      2.4.1         |       0.8.7        |
|   1.2    |      4.1.3         |       0.9.8        |
|   1.3    |      4.3.0         |       0.10.17      |
|   1.4    |      4.3.0         |       0.10.17      |
|   1.5    |      4.3.0         |       0.10.27      |
|   1.6    |      4.3.0         |       0.10.27      |
|   1.7    |      6.0.3         |       0.12.10      |
|   1.8    |      7.1.3         |       1.0.0        |

**npm Package Dependency:** Run the following commands to restore the package dependencies.
```
#For the web UI
cd $REPO_DIR/src/portal
npm install

#For the UI library
cd $REPO_DIR/src/portal/lib
npm install
```


To run the code, please refer to the [build](docs/compile_guide.md) guideline.

## Contribute Workflow

PR are always welcome, even if they only contain small fixes like typos or a few lines of code. If there will be a significant effort, please document it as an issue and get a discussion going before starting to work on it.

Please submit a PR broken down into small changes bit by bit. A PR consisting of a lot features and code changes may be hard to review. It is recommended to submit PRs in an incremental fashion.

Note: If you split your pull request to small changes, please make sure any of the changes goes to master will not break anything. Otherwise, it can not be merged until this feature complete. 

The graphic shown below describes the overall workflow about how to contribute code to Harbor repository.
![contribute workflow](docs/img/workflow.png)

### Fork and clone

Fork the Harbor repository and clone the code to your local workspace. Per Go's [workspace instructions](https://golang.org/doc/code.html#Workspaces), place Harbor's code on your `GOPATH`. Refer to section [Fork Repository](#fork-repository) for details.

Define a local working directory:
```
working_dir=$GOPATH/src/github.com/goharbor
```

Set user to match your github profile name:
```
user={your github profile name}
```

Both `$working_dir` and `$user` are mentioned in the figure above.

### Branch
Changes should be made on your own fork in a new branch. The branch should be named  `XXX-description` where XXX is the number of the issue. PR should be rebased on top of master without multiple branches mixed into the PR. If your PR do not merge cleanly, use commands listed below to get it up to date.

```
#goharbor is the origin upstream

cd $working_dir/kubernetes
git fetch goharbor
git checkout master
git rebase goharbor/master
```

Branch from the updated `master` branch:

```
git checkout -b my_feature master
```

### Develop, Build and Test

Write code on the new branch in your fork. The coding style used in Harbor is suggested by the Golang community. See the [style doc](https://github.com/golang/go/wiki/CodeReviewComments) for details.

Try to limit column width to 120 characters for both code and markdown documents such as this one.

As we are enforcing standards set by [golint](https://github.com/golang/lint), please always run golint on source code before committing your changes. If it reports an issue, in general, the preferred action is to fix the code to comply with the linter's recommendation 
because golint gives suggestions according to the stylistic conventions listed in [Effective Go](https://golang.org/doc/effective_go.html) and the [CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments).
```
#Install fgt and golint

go get -u golang.org/x/lint/golint
go get github.com/GeertJohan/fgt

#In the #working_dir/harbor, run

go list ./... | grep -v -E 'vendor|tests' | xargs -L1 fgt golint

```

Unit test cases should be added to cover the new code. Unit test framework for backend services is using [go testing](https://golang.org/doc/code.html#Testing). The UI library test framework is built based on [Jasmine](http://jasmine.github.io/2.4/introduction.html) and [Karma](https://karma-runner.github.io/1.0/index.html), please refer to [Angular Testing](https://angular.io/guide/testing) for more details.

Run go test cases:
```
#cd #working_dir/src/[package]
go test -v ./...
```

Run UI library test cases:
```
#cd #working_dir/src/portal/lib
npm run test
```

To build the code, please refer to [build](docs/build-customize-contribute/compile-guide.md) guideline.

###  Keep sync with upstream


Once your branch gets out of sync with the goharbor/master branch, use the following commands to update:
```bash
git checkout my_feature
git fetch -a
git rebase goharbor/master

```

Please use `fetch / rebase` (as shown above) instead of `git pull`. `git pull` does a merge, which leaves merge commits. These make the commit history messy and violate the principle that commits ought to be individually understandable and useful (see below). You can also consider changing your `.git/config` file via git config `branch.autoSetupRebase` always to change the behavior of `git pull`.

### Commit

As Harbor has integrated the [DCO (Developer Certificate of Origin)](https://probot.github.io/apps/dco/) check tool, contributors are required to sign-off that they adhere to those requirements by adding a `Signed-off-by` line to the commit messages. Git has even provided a `-s` command line option to append that automatically to your commit messages, please use it when you commit your changes.

```bash
$ git commit -s -m 'This is my commit message'
```

Commit your changes if they're ready:
```bash
git add -A
git commit -s #-a
git push --force-with-lease $user my_feature
```

The commit message should follow the convention on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/). Be sure to include any related GitHub issue references in the commit message. See [GFM syntax](https://guides.github.com/features/mastering-markdown/#GitHub-flavored-markdown) for referencing issues and commits.

To help write conformant commit messages, it is recommended to set up the [git-good-commit](https://github.com/tommarshall/git-good-commit) commit hook. Run this command in the Harbor repo's root directory:

```
curl https://cdn.rawgit.com/tommarshall/git-good-commit/v0.6.1/hook.sh > .git/hooks/commit-msg && chmod +x .git/hooks/commit-msg
```

### Automated Testing
Once your pull request has been opened, harbor will run two CI pipelines against it.
1. In the travis CI, your source code will be checked via `golint`, `go vet` and `go race` that makes sure the code is readable, safe and correct. Also all of unit tests will be triggered via `go test` against the pull request. What you need to pay attention to is the travis result and the coverage report.
* If any failure in travis, you need to figure out whether it is introduced by your commits.
* If the coverage dramatic decline, you need to commit unit test to coverage your code.
2. In the drone CI, the E2E test will be triggered against the pull request. Also, the source code will be checked via `gosec`, and the result is stored in google storage for later analysis. The pipeline is about to build and install harbor from source code, then to run four very basic E2E tests to validate the basic functionalities of harbor, like:
* Registry Basic Verification, to validate the image can be pulled and pushed successful.
* Clair Basic Verification, to validate the image can be scanned successful.
* Notary Basic Verification, to validate the image can be signed successful.
* Ldap Basic Verification, to validate harbor can work in LDAP environment.

### Push and Create PR
When ready for review, push your branch to your fork repository on `github.com`:
```
git push --force-with-lease $user my_feature

```

Then visit your fork at https://github.com/$user/harbor and click the `Compare & Pull Request` button next to your `my_feature` branch to create a new pull request (PR). Description of a pull request should refer to all the issues that it addresses. Remember to put a reference to issues (such as `Closes #XXX` and `Fixes #XXX`) in commits so that the issues can be closed when the PR is merged.

Once your pull request has been opened it will be assigned to one or more reviewers. Those reviewers will do a thorough code review, looking for correctness, bugs, opportunities for improvement, documentation and comments, and style.

Commit changes made in response to review comments to the same branch on your fork.

## Reporting issues

It is a great way to contribute to Harbor by reporting an issue. Well-written and complete bug reports are always welcome! Please open an issue on Github and follow the template to fill in required information.

Before opening any issue, please look up the existing [issues](https://github.com/goharbor/harbor/issues) to avoid submitting a duplication.
If you find a match, you can "subscribe" to it to get notified on updates. If you have additional helpful information about the issue, please leave a comment.

When reporting issues, always include:

* Version of docker engine and docker-compose
* Configuration files of Harbor
* Log files in /var/log/harbor/

Because the issues are open to the public, when submitting the log and configuration files, be sure to remove any sensitive information, e.g. user name, password, IP address, and company name. You can
replace those parts with "REDACTED" or other strings like "****".

Be sure to include the steps to reproduce the problem if applicable. It can help us understand and fix your issue faster.

## Documenting

Update the documentation if you are creating or changing features. Good documentation is as important as the code itself.

The main location for the document is the `docs/` folder. The images referred in documents can be placed in `docs/img`.

Documents are written with Markdown text. See [Writing on GitHub](https://help.github.com/categories/writing-on-github/) for more details.

## Design new features

You can propose new designs for existing Harbor features. You can also design entirely new features, Please submit a proposal in GitHub.(https://github.com/goharbor/community/tree/master/proposals). Harbor maintainers will review this proposal as soon as possible. This is necessary to ensure the overall architecture is consistent and to avoid duplicated work in the roadmap.

[community-meetings]: https://github.com/goharbor/community/blob/master/MEETING_SCHEDULE.md
[past-meetings]: https://www.youtube.com/playlist?list=PLgInP-D86bCwTC0DYAa1pgupsQIAWPomv
[users-slack]: https://cloud-native.slack.com/archives/CC1E09J6S
[dev-slack]: https://cloud-native.slack.com/archives/CC1E0J0MC
[cncf-slack]: https://slack.cncf.io
[users-dl]: https://lists.cncf.io/g/harbor-users
[dev-dl]: https://lists.cncf.io/g/harbor-dev
[twitter]: http://twitter.com/project_harbor
