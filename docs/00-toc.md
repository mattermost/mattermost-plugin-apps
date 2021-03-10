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
- [Hello World - Simple App anatomy](01-anatomy-hello.md)
- Development environment
  See https://docs.google.com/document/d/1-o9A8l65__rYbx6O-ZdIgJ7LJgZ1f3XRXphAyD7YfF4/edit#
- [Functions](02-functions.md)
- [Forms](03-forms.md)
- [In-Post Interactivity](04-in-post.md)
- [Using Mattermost APIs](05-mattermost-API.md)
- [Using 3rd party APIs](06-3rd-party-API.md)
- [App Lifecycle](07-app=lifecycle)

# Forms
## Binding Forms to Locations
### Channel Header
### Post Menu
### /Command Autocomplete
## Autocomplete
## Modals


# In-Post Interactivity

# Using Mattermost APIs
## Authentication and Token Store
## Scopes and Permissions
## Apps Subscriptions API
## Apps KV API
## Mattermost REST API

# Using 3rd party APIs
## Authentication and Token Store
## 3rd party webhooks

# App Lifecycle
## Development
## Submit to Marketplace
## Provision
## Publish
## Install
## Uninstall
## Upgrade/downgrade consideration