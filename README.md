<img src="assets/logo.svg" alt="topaz logo">

# Topaz - cloud-native authorization for modern applications and APIs

[![Go Report Card](https://goreportcard.com/badge/github.com/aserto-dev/topaz)](https://goreportcard.com/report/github.com/aserto-dev/topaz)
[![ci](https://github.com/aserto-dev/topaz/actions/workflows/ci.yaml/badge.svg)](https://github.com/aserto-dev/topaz/actions/workflows/ci.yaml)
![Apache 2.0](https://img.shields.io/github/license/aserto-dev/topaz)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/aserto-dev/topaz)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6652/badge)](https://bestpractices.coreinfrastructure.org/projects/6652)
[<img src="https://img.shields.io/badge/slack-@asertocommunity-yellow.svg?logo=slack">](https://www.aserto.com/slack)
[<img src="https://img.shields.io/badge/docs-%F0%9F%95%B6-blue">](https://www.topaz.sh/docs/intro)
<a href="https://twitter.com/intent/follow?screen_name=aserto_com"><img src="https://img.shields.io/badge/Follow-aserto__com-blue?style=flat&logo=twitter"></a>


Topaz is an open-source authorization service providing fine-grained, real-time, policy-based access control for applications and APIs.

It uses the [Open Policy Agent](https://www.openpolicyagent.org/) (OPA) as its decision engine, and provides a built-in directory that is inspired by the Google [Zanzibar](https://research.google/pubs/pub48190/) data model.

Authorization policies can leverage user attributes, group membership, application resources, and relationships between them. All data used for authorization is modeled and stored locally in an embedded database, so authorization decisions can be evaluated quickly and efficiently.

## Documentation and support

Read more at [topaz.sh](https://www.topaz.sh) and the [docs](https://www.topaz.sh/docs/intro).

Join the community [Slack channel](https://www.aserto.com/slack) for questions and help!

## Benefits

* **Authorization in one place**: a single authorization service, instead of spreading authorization logic everywhere.
* **Fine-grained**: following the Principle of Least Privilege, assign the smallest set of fine-grained permissions to each user or group.
* **Policy-based**: convert authorization "spaghetti code" into a policy expressed in its own domain-specific language, managed as code, and built into an immutable, signed artifact.
* **Real-time**: gate each protected resource with an authorization call that ensures the user has the right permission.
* **Blazing fast**: deploy the authorizer as a sidecar or microservice, right next to your app, for low latency and high availability.
* **Comprehensive decision logging**: log every decision to facilitate audit trails, compliance, and forensics.
* **Flexible authorization model**: Start simple, and grow from multi-tenant RBAC to ABAC or ReBAC, or a combination.
* **Capture your domain model**: Create object types and relationships that reflect your domain model.
* **Separation of concerns**: application developers can own the app logic, and security engineers can own the authorization policy.

## Table of Contents
- [Getting Topaz](#getting-topaz)
    - [Installation](#installation)
    - [Building from source](#building-from-source)
    - [Running with Docker](#running-with-docker)
- [Quickstart](#quickstart)
    - [Install container image](#install-topaz-authorizer-container-image)
    - [Create config for Todo policy](#create-a-configuration)
    - [Start in interactive mode](#start-topaz-in-interative-mode)
    - [Import sample data](#import-sample-data)
    - [Issue an API call](#issue-an-api-call)
    - [Issue a query](#issue-a-query)
    - [Run the sample application](#run-the-sample-application)
- [Command Line](#command-line-options)
- [gRPC Endpoints](#grpc-endpoints)
- [Demo video](#demo)
- [Credits](#credits)
- [Contribution Guidelines](#contribution-guidelines)

## Getting Topaz

### Installation

`topaz` is available on Linux, macOS and Windows platforms.

* Binaries for Linux, Windows and Mac are available as tarballs in the [release](https://github.com/aserto-dev/topaz/releases) page.

* Via Homebrew for macOS or LinuxBrew for Linux

   ```shell
  brew tap aserto-dev/tap && brew install aserto-dev/tap/topaz
   ```

* Via a GO install

  ```shell
  go install github.com/topaz/cmd/topaz@latest
  ```

### Building from source

 `topaz` is currently using go v1.17 or above. In order to build `topaz` from source you must:

 1. Install [mage](https://magefile.org/)
 2. Clone the repo
 3. Build and run the executable

      ```shell
      mage build && ./dist/build_linux_amd64/topaz
      ```

### Running with Docker

  You can run as a Docker container:

  ```shell
  docker run -it --rm ghcr.io/aserto-dev/topaz:latest --help
  ```

## Quickstart

These instructions help you get Topaz up and running as the authorizer for a sample Todo app.

### Install Topaz authorizer container image

The Topaz authorizer is packaged as a Docker container. You can get the latest image using the following command:

```shell
topaz install
```

### Create a configuration

This command creates a configuration file for the sample Todo **policy image**. A policy image is an OCI image that contains an OPA policy. The source code for the `ghcr.io/aserto-policies/policy-todo-rebac:latest` policy image can be found [here](https://github.com/aserto-templates/template-policy-todo-rebac/tree/main/content/src/policies).

```shell
topaz configure -d -s -r ghcr.io/aserto-policies/policy-todo-rebac:latest -n todo
```

The configuration file is generated in `$(HOME)/.config/topaz/cfg`.
* the config instructs Topaz to create a local directory instance (`-d`)
* when started, Topaz will seed the directory with default object types (`-s`)
* the config references an authorization policy for a sample "Todo" app, retrieved from the Open Policy Registry as a container image
* the config is named "todo"

#### Creating a configuration that uses a local policy CLI image

If you have a policy image in the local OCI store of your policy CLI that you want to use with topaz you can create a configuration to use that image from the local store. 

```
topaz configure -d -s -l ghcr.io/default:latest
```
The configuration file is generated in `$(HOME)/.config/topaz/cfg`.
* the config instructs Topaz to create a local directory instance (`-d`)
* when started, Topaz will seed the directory with default object types (`-s`)
* the config uses the opa local_bundles configuration to retrieve the policy image from the local policy CLI OCI store

### Start Topaz in interative mode

```shell
topaz run
```

### Import sample data

Retrieve the "Citadel" json files, placing them in the current directory:

```shell
curl https://raw.githubusercontent.com/aserto-dev/topaz/main/assets/citadel-objects.json >./citadel-objects.json
curl https://raw.githubusercontent.com/aserto-dev/topaz/main/assets/citadel-relations.json >./citadel-relations.json
```

Import the contents of the file into Topaz directory. This creates the sample users (Rick, Morty, and friends); groups; and relations.

```shell
topaz import -i -d .
```

### Issue an API call

To verify that Topaz is running with the right policy image, you can issue a `curl` call to interact with the REST API.

This API call retrieves the set of policies that Topaz has loaded:

```shell
curl -k https://localhost:8383/api/v2/policies
```

### Issue a query

Issue a query using the `is` REST API to verify that the user Rick is allowed to GET the list of todos:

```shell
curl -k -X POST 'https://localhost:8383/api/v2/authz/is' \
-H 'Content-Type: application/json' \
-d '{
     "identity_context": {
          "type": "IDENTITY_TYPE_SUB",
          "identity": "rick@the-citadel.com"
     },
     "policy_context": {
          "path": "todoApp.GET.todos",
          "decisions": ["allowed"]
     }
}'
```

### Run the sample application

To run the sample Todo app in the language of your choice, and see how Topaz is used to authorize requests, refer to the [docs](https://www.topaz.sh/docs/getting-started/samples).

To start an interactive session with the Topaz endpoints, see the [gRPC endpoints](#grpc-endpoints) section.

## Command line options

```shell
$ topaz --help
Usage: topaz <command>

Topaz CLI

Commands:
  backup       backup directory data
  configure    configure topaz service
  export       export directory objects
  install      install topaz
  import       import directory objects
  load         load a manifest file
  restore      restore directory data
  run
  save         save a manifest file
  start        start topaz instance
  status       display topaz instance status
  stop         stop topaz instance
  version      version information
  uninstall    uninstall topaz, removes all locally installed artifacts

Flags:
  -h, --help    Show context-sensitive help.

Run "topaz <command> --help" for more information on a command.
```

## gRPC Endpoints

To interact with the authorizer endpoint, install `grpcui` or `grpcurl` and point them to `localhost:8282`:

```shell
grpcui --insecure localhost:8282
```

To interact with the directory endpoint, use `localhost:9292`:

```shell
grpcui --insecure localhost:9292
```

For more information on APIs, see the [docs](https://www.topaz.sh/docs/intro).

## Demo
![demo](./assets/topaz.gif)

## Credits

Topaz uses a lot of great and amazing open source projects and libraries.

A big thank you to all of them!

## Contribution Guidelines

Topaz is a work in progress - if something is broken or there's a feature that you want, please file an issue and if so inclined submit a PR!

We welcome contributions from the community! Here are some general guidelines:

* File an issue first prior to submitting a PR!
* Ensure all exported items are properly commented
* If applicable, submit a test suite against your PR
