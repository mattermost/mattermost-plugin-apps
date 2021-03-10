# Overview

Apps are lighweight interactive add-ons to mattermost. Apps can:
- display interactive, dynamic Modal forms and Message Actions.
- attach themselves to locations in the Mattermost UI (e.g. channel bar buttons,
  post menu, channel menu, commands), and can add their custom /commands with
  full Autocomplete.
- receive webhooks from Mattermost, and from 3rd parties, and use the Mattermost
  REST APIs to post messages, etc. 
- be hosted externally (HTTP), on Mattermost Cloud (AWS Lambda), and soon
  on-prem and in customers' own AWS environments.
- be developed in any language*

# Contents
#### Development environment
- [Google Doc](https://docs.google.com/document/d/1-o9A8l65__rYbx6O-ZdIgJ7LJgZ1f3XRXphAyD7YfF4/edit#) - DRAFT Dev environment doc.

#### Hello, World
- [Anatomy](01-anatomy-hello.md)

#### Functions, Calls
- [Post Menu Flow](02-example-post-menu.md) - message flow to define a Post Menu action, then have a user click it.
- godoc: [Call](https://pkg.go.dev/github.com/mattermost/mattermost-plugin-apps/apps#Call) - describes how to call a function.
- godoc: [CallRequest](https://pkg.go.dev/github.com/mattermost/mattermost-plugin-apps/apps#CallRequest) - structure of a request to a function.
- godoc: [Context](https://pkg.go.dev/github.com/mattermost/mattermost-plugin-apps/apps#Context) - extra data passed to call requests.
- godoc: [Expand](https://pkg.go.dev/github.com/mattermost/mattermost-plugin-apps/apps#Expand) - controls context expansion.

#### Forms
- [godoc Form](https://pkg.go.dev/github.com/mattermost/mattermost-plugin-apps/apps#Form)

#### Bindings and Locations

#### Autocomplete
#### Modals

#### [In-Post Interactivity](04-in-post.md)

#### [Using Mattermost APIs](05-mattermost-API.md)

#### [Using 3rd party APIs](06-3rd-party-API.md)

#### App Lifecycle
- godoc: [appsctl](https://pkg.go.dev/github.com/mattermost/mattermost-plugin-apps/cmd/appsctl) - CLI tool used to provision Mattermost Apps in development and production.
- [App Lifecycle](07-app=lifecycle)







