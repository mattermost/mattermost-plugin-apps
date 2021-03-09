## Overview
### What are Apps?
- Apps are lighweight interactive add-ons to mattermost. 
- Apps can display interactive, dynamic Modal forms.
- Apps can attach themselves to locations in the Mattermost UI (e.g. channel bar buttons, post menu, channel menu, commands), and can add their custom /commands with full Autocomplete.
- Apps can receive webhooks from Mattermost, and from 3rd parties, and use the Mattermost REST APIs to post messages, etc. 
- Apps can be hosted externally (HTTP), on Mattermost Cloud (AWS Lambda), and soon on-prem and in customers' own AWS environments.
- Apps can be developed in any language*

### Hello App

## Anatomy of an App
### Manifest
### Bindings and Locations
### Functions
### Icons 
### Path routing
### OAuth2 support

## Development environment

## Forms and Functions
### Authentication
### Context Expansion
### Call vs Notification

## Interactive Features
### Locations and Bindings
### /commands and autocomplete
### Modals
### In-Post interactivity

## Using Mattermost APIs
### Authentication and Token Store
### Scopes and Permissions
### Apps Subscriptions API
### Apps KV API
### Mattermost REST API

## Using 3rd party APIs
### Authentication and Token Store
### 3rd party webhooks

## Lifecycle
### Development
### Submit to Marketplace
### Provision
### Publish
### Install
### Uninstall
### Upgrade/downgrade consideration