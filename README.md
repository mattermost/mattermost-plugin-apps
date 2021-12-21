# mattermost-plugin-apps

![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-apps/master.svg)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-apps/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-apps/branch/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattermost/mattermost-plugin-apps)](https://goreportcard.com/report/github.com/mattermost/mattermost-plugin-apps)

# Mattermost Apps

This plugin serves as the core of the Mattermost Apps Framework. It extends the Mattermost server's API to allow for the creation of feature-rich integrations, with functionality supported on the Mattermost web client and mobile client. Take a look at the [App developer documentation](https://developers.mattermost.com/integrate/apps) for more information.

Join the [Mattermost Apps channel](https://community.mattermost.com/core/channels/mattermost-apps) on our community server to discuss technical details and use cases for the App you're creating.

## Getting Started

Use the [Apps Framework development environment](dev/README.md) to get up and running quickly. Running the command `make dev_server` spins up a test Mattermost instance with all of the settings configured to develop Apps.

Learn more about developing Apps by reading the [App developer documentation](https://developers.mattermost.com/integrate/apps/).

## Running the tests

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

## Deploying and Installing Apps

See [documentation](https://developers.mattermost.com/integrate/apps/deploy/)

## Contacts

- Dev: Lev Brouk (@lev.brouk)
- PM: Aaron Rothschild (@aaron.rothschild)
