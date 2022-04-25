# Hello, world!

This is the quick start guide that explains how to write your first Mattermost app. This simple app:

- Provides a [manifest](https://github.com/mattermost/mattermost-plugin-apps/blob/master/examples/go/hello-world/hello.go#:~:text=var%20Manifest)
  ([docs](https://developers.mattermost.com/integrate/apps/api/manifest/)), in
  which it declares itself an HTTP app that acts as a bot, and binds to
  locations in the user interface.
- Contains a [send form](https://github.com/mattermost/mattermost-plugin-apps/blob/master/examples/go/hello-world/hello.go#:~:text=var%20SendForm)
  ([docs](https://developers.mattermost.com/integrate/apps/api/interactivity/))
  that allows to enter a custom message.
- [Binds](https://github.com/mattermost/mattermost-plugin-apps/blob/master/examples/go/hello-world/hello.go#:~:text=var%20Bindinings%20callback)
  ([docs](https://developers.mattermost.com/integrate/apps/api/bindings/)) the
  form to the channel header, and defines a `/helloworld send` command.
- Contains the [Send](https://github.com/mattermost/mattermost-plugin-apps/blob/master/examples/go/hello-world/hello.go#:~:text=func%20Send)
  call handler that sends a parameterized message back to the user.

## Prerequisites

Before you can start with your app, you first need to set up a local developer
environment following the [server](https://developers.mattermost.com/contribute/server/developer-setup/)
and [webapp](https://developers.mattermost.com/contribute/webapp/developer-setup/)
setup guides. You need Mattermost v6.6 or later.

In the System Console, ensure that the following are set to **true**:

- `Enable Bot Account Creation`
- `Enable OAuth 2.0 Service Provider`

You also need at least `go1.16` installed. Please follow the guide
[here](https://golang.org/doc/install) to install the latest version.

### Install the Apps plugin

The [apps plugin](https://github.com/mattermost/mattermost-plugin-apps) is a
communication bridge between your app and the Mattermost server. To install it
on your local server, start by cloning the code in a directory of your choice
run:

```bash
git clone https://github.com/mattermost/mattermost-plugin-apps.git
```

Then build the plugin using:

```bash
cd mattermost-plugin-apps
make dist
```

Then upload it to your local Mattermost server via the System Console.

## Running the app

```bash
cd ./examples/go/hello-world
go run .
```

The app runs on port `4000`.

## Installing the app

Run the following slash commands on your Mattermost server:

```
/apps install http http://localhost:4000/manifest.json
```

Confirm the installation in the modal that pops up.

## Using the app

Select the "Hello World" channel header button in Mattermost, which brings up a modal:

**TODO** need image

Type `testing` and select **Submit**, you should see:

**TODO need image**

You can also use the `/helloworld send` command by typing `/helloworld send
--message Hi!`. This posts the message to the Mattermost channel that you're
currently in.

**TODO** need image
