# mattermost-plugin-apps

![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-apps/master.svg)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-apps/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-apps/branch/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattermost/mattermost-plugin-apps)](https://goreportcard.com/report/github.com/mattermost/mattermost-plugin-apps)

# Mattermost Apps

This plugin serves as the core of the Mattermost App Framework. It extends the Mattermost server's API to allow for the creation of feature-rich integrations, with functionality supported on the Mattermost web client and mobile client. Take a look at the [App developer documentation](https://developers.mattermost.com/integrate/apps) for more information.

Join the [Mattermost Apps channel](https://community.mattermost.com/core/channels/mattermost-apps) on our community server to discuss technical details and use cases for the App you're creating.

## Getting Started

Use the App Framework [Docker development environment](dev) to get up and running quickly. This development environment is used to spin up several Docker containers, so that a Mattermost server can communicate with those containers. When you are making changes to Mattermost projects (e.g. server, webapp), you have to manually run the Mattermost server, which communicates to these Docker containers.

Running the command `make dev_server` spins up a test Mattermost instance with all of the settings configured to develop Apps.

Learn more about developing Apps by reading the [App developer documentation](https://developers.mattermost.com/integrate/apps).

## Running the tests

`mattermost-plugin-apps` has two types of tests: unit tests and end to end tests.

### Unit tests

To run the unit tests, you just need to execute:

```sh
make test
```

### End to end tests

The App Framework e2e tests in the App Framework project require the same Docker containers, used in the development environment step, to be running. However these tests don't need a Mattermost server to be running. The tests instead mimic the behavior of a running server using shared code of the `mattermost-server` repository. You can think of it as a "fake server" running, completely separate from the running containers, but communicating with the containers.

When you're developing your own App, you need an actual Mattermost server to be running. The Mattermost App Framework Docker development environment helps accomplish this by setting up a minimalistic environment with just two containers. One is for the database Mattermost communicates with, and the other container runs the actual Mattermost server. The other containers present in the Mattermost Docker development environment are unnecessary for the purposes of building Apps. So the advantage here is that there is just one dependency to start developing Apps.

Some differences between the environments:

* The App Framework Docker development environment has its own Mattermost server, and it's fully configured to start developing Apps. The config values are set correctly so no modifications need to be done there.

* The App Framework Docker development environment also has a starter App built-in as a third container, but this can be ignored if the developer wishes to run their App outside of the dev environment, while still using it by communicating with it from outside of the containers.

* The App Framework e2e tests can't be run with the App Framework development environment.

More specific information about the App Framework Docker development environment is explained in [dev/README.md](dev/README.md).

To run the end to end test suite, you need to have the Mattermost server project downloaded and configured in your system. Check the [Developer Setup](https://developers.mattermost.com/contribute/server/developer-setup/) guide on how to configure a local server instance. The tests will search for a `mattermost-server` folder in the same directory where the `mattermost-plugin-apps` is.

With the `mattermost-server` folder present, the only thing that needs to be done before running the tests themselves is to start the Mattermost Docker development environment. The environment only needs to be started once, and then the tests can run as many times as needed. To start the Docker environment, change to the `mattermost-server` project directory and run:

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

Dev: Lev Brouk (@lev.brouk)
PM: Aaron Rothschild (@aaron.rothschild)
