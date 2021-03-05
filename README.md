# mattermost-plugin-apps

![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-apps/master.svg)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-apps/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-apps/branch/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattermost/mattermost-plugin-apps)](https://goreportcard.com/report/github.com/mattermost/mattermost-plugin-apps)


# Proof Of Concept - Mattermost Apps

This plugin is being developed to test some concepts of creating Apps, which do not rely on a Go executable being installed on the Mattermost server/cluster to extend functionality.  The Apps will not be able to use Go RPC to communicate with the Mattermost Server, only through the "App Plugin" which acts as a sort of proxy to the server's activity.

Apps will generally be communicating with our REST API and authenticating via OAuth.

This is a precursor to our "Mattermost Apps" and "Mattermost Apps Marketplace" we are currently researching.

## Getting Started

Join the "Integrations and Apps" channel to provide thoughts and feedback.

## Running the tests

`mattermost-plugin-apps` has two types of tests: unit tests and end to end tests.

### Unit tests

To run the unit tests, you just need to execute:

```sh
make test
```

### End to end tests

To run the end to end test suite, you need to have the Mattermost server project downloaded and configured in your system. Check the [Developer Setup](https://developers.mattermost.com/contribute/server/developer-setup/) guide on how to configure a local server instance. The tests will search for a `mattermost-server` folder in the same directory where the `mattermost-plugin-apps` is.

With the `mattermost-server` folder present, the only thing that needs to be done before running the tests themselves is to start the Mattermost docker development environment. The environment only needs to be started once, and then the tests can run as many times as needed. To start the docker environment, change to the `mattermost-server` project directory and run:

```sh
make start-docker
```

Change your directory back to `mattermost-plugin-apps` and run the end to end test suite with:

```sh
make test-e2e
```

## Installing AWS Apps

To install an AWS App you need to set `APPS_INVOKE_AWS_ACCESS_KEY` and `APPS_INVOKE_AWS_SECRET_KEY`. These credentials should only allow the invocation of lambda function, not the creation.

## Provisioning

To provision an App to AWS you first need to store your AWS access key in an environment variable called `APPS_PROVISION_AWS_ACCESS_KEY` and the secret key in `APPS_PROVISION_AWS_SECRET_KEY`.

Only once you need to run `go run ./cmd/appsctl/ provision bucket` to create the s3 bucket.

To provision an app run `go run ./cmd/appsctl/ provision app /PATH/TO/YOUR/APP/BUNDLE`. Use `--update` to update the lambda functions if they already exist. This option should not be used in production.

Run `go run ./cmd/appsctl/ generate-terraform-data /PATH/TO/YOUR/APP/BUNDLE` to generate terraform variables. It will print the indented json directly in the console. Pipe as needed.

## Contacts

Dev: Lev Brouk (@lev.brouk)
PM: Aaron Rothschild (@aaron.rothschild)
