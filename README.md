> [!WARNING]
> This repository is deprecated and no longer maintained. For more information about the transition to the Plugin Framework, please see the [official announcement](https://forum.mattermost.com/t/transition-from-apps-framework-to-plugin-framework-in-mattermost-v10/19330). Refer to the official Mattermost documentation for current integration solutions.

# Mattermost Apps Framework Plugin

![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-apps/master.svg)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-apps/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-apps/branch/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattermost/mattermost-plugin-apps)](https://goreportcard.com/report/github.com/mattermost/mattermost-plugin-apps)

## Contents
- [Overview](#overview)
- [Deploying and Installing Apps](#deploying-and-installing-apps)
- [Develop an App](#develop-an-app)
- [Running the Tests](#running-the-tests)

## Overview

This plugin serves as the core of the Mattermost Apps Framework. It extends the Mattermost server's API to allow for the creation of feature-rich integrations, with functionality supported on the Mattermost web client and mobile client. Take a look at the [app developer documentation](https://developers.mattermost.com/integrate/apps) for more information.

Join the [Mattermost Apps channel](https://community.mattermost.com/core/channels/mattermost-apps) on our community server to discuss technical details and use cases for the app you're creating.

## Deploying and Installing Apps

See [legacy documentation](https://github.com/mattermost/mattermost-developer-documentation/blob/a47fef0b3848eff9be54ca58f0b875958cd4421b/site/content/integrate/apps/_index.md)

## Develop an App

Refer to the [Mattermost Apps Quick Start Guide](https://developers.mattermost.com/integrate/apps/quickstart/) for instructions on how to use the development environment and examples in the [mattermost/mattermost-app-examples](https://github.com/mattermost/mattermost-app-examples) repository.

## Running the Tests

`mattermost-plugin-apps` has two types of tests: unit tests and end to end tests.

### Unit tests

To run the unit tests, you just need to execute:

```sh
make test
```

### End to end tests

The Apps Framework e2e tests written in go require the same Docker containers used in the [Mattermost development environment](https://developers.mattermost.com/contribute/server/developer-setup/) to be running. However these tests don't need a Mattermost server to be running. The tests instead mimic the behavior of a running server using shared code of the `mattermost-server` repository. You can think of it as a "fake server" running, completely separate from the running containers, but communicating with the containers.

To run the end to end test suite, you need to have the Mattermost server project downloaded and configured in your system. Check the [Developer Setup](https://developers.mattermost.com/contribute/server/developer-setup/) guide on how to configure a local server instance. The tests will search for a `mattermost-server` folder in the same directory where the `mattermost-plugin-apps` is.

With the `mattermost-server` folder present, the only thing that needs to be done before running the tests themselves is to start the Mattermost development environment. The environment only needs to be started once, and then the tests can run as many times as needed. To start the Docker environment, change to the `mattermost-server` project directory and run:

```sh
make start-docker
```

Change your directory back to `mattermost-plugin-apps` and run the end to end test suite with:

```sh
make test-e2e
```

## How to Release

To trigger a release, follow these steps:

1. **For Patch Release:** Run the following command:
    ```
    make patch
    ```
   This will release a patch change.

2. **For Minor Release:** Run the following command:
    ```
    make minor
    ```
   This will release a minor change.

3. **For Major Release:** Run the following command:
    ```
    make major
    ```
   This will release a major change.

4. **For Patch Release Candidate (RC):** Run the following command:
    ```
    make patch-rc
    ```
   This will release a patch release candidate.

5. **For Minor Release Candidate (RC):** Run the following command:
    ```
    make minor-rc
    ```
   This will release a minor release candidate.

6. **For Major Release Candidate (RC):** Run the following command:
    ```
    make major-rc
    ```
   This will release a major release candidate.


