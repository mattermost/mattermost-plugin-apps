# Using Mattermost APIs

## Authentication and Permissions

**OAuth2 is not yet implemented, for now session tokens are passed in as ActingUserAccessToken**

An app can use the Mattermost server REST API, as well as new "App Services" APIs offered specifically to Mattermost Apps. An app authenticates its requests to Mattermost by providing access tokens, usually Bot Access token, or user's OAuth2 access token. Each call request sent to the app includes Mattermost site URL, and optionally one or more access tokens the app can use.

What tokens the app gets, and what access the app may have with them depends on the combination of App granted permissions, the tokens requested in call.Expand, and their respective access rights.

- godoc: [Permission](https://pkg.go.dev/github.com/mattermost/mattermost-plugin-apps/apps#Permission) -
  [local](http://localhost:6060/pkg/github.com/mattermost/mattermost-plugin-apps/apps#Permission) -
  describes the available permissions.
- tickets:
  - [MM-??]()

## Apps Subscriptions API

## Apps KV Store API

## Mattermost REST API

## Go client


