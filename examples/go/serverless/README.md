# Example Serverless App

This app is similar to [Hello, World!](../hello-world) app, but is buildable
for, and deployable to AWS Lambda and OpenFaaS, in addition to being runnable as
an HTTP server - for debugging, or otherwise.

It is also possible, but is not illustrated in this example, to package go apps
as Mattermost Plugins so that they can run on the Mattermost server itself,
without another hosting platform. See the work-in-progress [Google Calendar
App](https://github.com/mattermost/mattermost-app-gcal) as an example.



## What's in the App?

This app contains:

- [manifest](./manifest.json). In addition to what's in [Hello, World!](../hello-world),
  this app declares several `Deploy` modes: `http`, `aws_lambda`, and
  `open_faas`. Note that the manifest is packaged as a separate file, to access
  from [appsctl](https://developers.mattermost.com/integrate/apps/deploy/) for
  deployment.
- [Makefile](./Makefile) is used to build the bundles to be deployed with
  [appsctl](https://developers.mattermost.com/integrate/apps/deploy/).
- `./function` contains the "core" of the app, everything except the static
  files, and platform-specific `main` packages is there. Note that because
  OpenFaaS deployment presently builds the app from its own template, the
  `function` package has its own `go.mod`.
- `./aws`, `./http` - main "stubs" for the respective platforms.
- `./openfaas` - support files to build the OpenFaaS-deployable bundle.
- `./static` - the app's icon.

## Prerequisites

See [Hello, World!](../hello-world) for the steps to set up a minimal dev environment.

You also will need [appsctl](https://developers.mattermost.com/integrate/apps/deploy/)
installed, see the link for instructions.

## Run as HTTP

To run the app as a localhost HTTP server,
```bash
cd ./examples/go/serverless
make run
```

It will print out the exact `/apps install` command to install it onto Mattermost.

## Deploy and run on AWS

If you have an AWS Account, and would like to deploy and run the app as a lambda function,

```bash
cd ./examples/go/serverless
make dist-aws
```

This will create a `./dist/bundle-aws.zip` file containg the AWS-deployable app bundle. Follow the steps in [Admin's Guide](https://developers.mattermost.com/integrate/apps/deploy/deploy-aws) to setup your `appsctl` environment and initialize the AWS resources with `appsctl aws init` command. Then execute


```bash
appsctl aws deploy -v ./dist/bundle-aws.zip --install
```

If you need to re-deploy an update without changing the app's version number, you'll need to specify

```bash
appsctl aws deploy -v ./dist/bundle-aws.zip --update --install
```

This will create the app-specific Lambda and S3 resources, and update the
policies to access them from Mattermost.

### Deploy and run on OpenFaaS.

Build the bundle with

```bash
cd ./examples/go/serverless
make dist-openfaas
```

See [Admin's Guide](https://developers.mattermost.com/integrate/apps/deploy/deploy-openfaas)
for how to deploy to OpenFaaS with `faascli` and `appsctl`. Note that OpenFaaS
requires a (writeable) docker registry that is uses to store the functions, so
`appsctl` would need to push the function to.

A typical command would look like:

```bash
appsctl openfaas deploy -v ./dist/bundle-openfaas.zip --docker-registry=CHANGE-ME --install
```
