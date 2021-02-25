## Provisioning
- Packaging (http server -> Lambda function)
- Static assets: icons only
  - Serving them out of S3
  - Validate the contents to be harmless icons??

## Install
- Manifest -> Installed App
  - Consent to permissions, locations, OAuth app type
  - Create Bot+Access Token, OAuth App
  - HTTP: collect app’s JWT secret
- Invoke “OnInstall” callback on the App
  - Admin access token
- Also Uninstall/Enable/Disable per App

## Versioning
- Upgrade/Downgrade/callOnce
- adminToken??

## User connect

### Mattermost auth
- Now: session token
- Next: OAuth2-as-a-service, intercepting the Calls, Expand
- Future: OAuth2 with detailed scopes

### 3rd party auth
- Now: ???
- Next: Landing page, token storage “as a service”, Expand
- Future: ???

## Locations/Bindings

An App has the ability to register UI elements in different locations:
- Slash command
- Post dropdown menu button
- Channel header button
- Embedded in a post

When an App includes a UI element in a location, it’s called creating a binding to that location. Bindings are scoped to the user/channel. When a user visits a channel, the bindings are fetched for that user, in the context of the current channel, from all the registered Apps. This allows each App’s server to dynamically add things to the UI on a per-channel basis.

## Calls

A Call is a request to an App server on behalf of a user (as well as events such as MessageHasBeenPosted). When the user performs some action (e.g. visiting a channel or submitting a form), a Call is sent to give the App server context of what channel/post/binding a user is interacting with. The Call also contains certain tokens in its payload (bot, user, and/or admin depending on request).

There are currently 3 types of calls associated with user actions:
- Submit - submit a form/command or click on a UI binding
- Form - Fetch a form’s definition like a command or modal
- Lookup - Fetch autocomplete results for am autocomplete form field

A Call request is sent to the App’s server when:
- The user visits a channel (call to fetch bindings)
- The user clicks on a post menu or channel binding (may open a modal)
- Commands:
  - The user is filling out a command argument that fetches dynamic results (lookup call is performed)
  - The user submits a command (may open a modal)
- Modals:
  - The user types a search in a modal’s autocomplete select field (lookup call is performed)
  - The user selects a value from a “refresh” select element in a modal (the modal’s form will be re-fetched based on all filled out values)
  - The user submits the modal (a new form may be returned from the App)
- (TBD) A subscribed event like MessageHasBeenPosted occurs
- (TBD) A third-party webhook request comes in


## Context Expansion

By default, only the IDs of certain entities are provided in the Call. Since we want these App functions to be stateless and quick, the App can specify certain objects to be "expanded" within the request to the App.

The framework verifies that the user has access to the entities before expanding and including them in the request. Choosing which level of access token to use for expand depends on what the App initially requests when installed:

- Bot token (possibly less privileged than admin)
- Acting user token
- Admin token


Entities that can be expanded:

- App - details about the App itself
  - see what is provided in summary vs full

- Acting User - model.User struct

- ActingUserAccessToken - OAuth2 token for acting user
  - currently is just acting MM session token

- AdminAccessToken - Access token to do admin-level operations
  - currently is just acting MM session token
  - Eventually make a short-lived bot admin token, scoped to certain operations

- Channel - The current channel the acting user is interacting with

- Mentioned - Users/channels mentioned in the related post or command being run
  - not implemented yet
  - should be access controlled as far as what is able to be expanded per user/context

- ParentPost - Parent post of the selected post
- Post - Selected post, if the call is specific to a given post
- RootPost - Root post of the selected post

- Team - The current ream the acting user is interacting with


### Conext Expansion Example

In this example, the bindings specify to expand the post the user clicks on:

![binding-form-diagram.png](https://user-images.githubusercontent.com/6913320/109165112-2e6ac800-7749-11eb-8d83-d495258f3f1e.png)

<details><summary>Client Bindings Request</summary>

GET /plugins/com.mattermost.apps/api/v1/bindings?channel_id=ei748ohj3ig4ijofs5tr47wozh&scope=webapp

</details>

<details><summary>MM Bindings Request</summary>
POST /plugins/com.mattermost.apps/example/hello/bindings

```json
{
    "url": "/bindings",
    "context": {
        "app_id": "http-hello",
        "acting_user_id": "d7mezwndk3yf3renn4fzeranpw",
        "channel_id": "edet7g6h8ib8dftytcjcqne8ie",
        "mattermost_site_url": "https://mickmister.ngrok.io",
        "bot_access_token": "ouddbrqwupfypqsu1qxbdu3uqo",
        "acting_user_access_token": "yemrienc7pfypqsu1qxbdu3uqo",
        "admin_access_token": "ue8xi2sh7ebcciw8duww84ucme"
    }
}
```

</details>

<details><summary>App Bindings Response</summary>

```json
[
    {
        "location": "/post_menu",
        "bindings": [
            {
                "app_id": "http-hello",
                "location": "send",
                "label": "Survey a user",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "/send-modal",
                    "expand": {
                        "post": "all",
                        "bot_access_token": "all",
                        "acting_user_access_token": "all",
                        "admin_access_token": "all"
                    }
                }
            }
        ]
    }
]
```
</details>

<details><summary>Client Submit Request</summary>

POST /plugins/com.mattermost.apps/api/v1/call
```json
{
    "url": "/send-modal",
    "expand": {
        "post": "all"
    },
    "type": "submit",
    "context": {
        "app_id": "http-hello",
        "location": "send",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "ei748ohj3ig4ijofs5tr47wozh",
        "post_id": "b7pkajox3bgmjjexo4yisu4iih",
        "root_id": ""
    }
}
```

</details>

<details><summary>MM Submit Request</summary>

POST /plugins/com.mattermost.apps/example/hello/send
```json
{
    "url": "/send-modal",
    "context": {
        "app_id": "http-hello",
        "location": "send",
        "bot_user_id": "uzofd8otciyktj7mqbawi4hexc",
        "acting_user_id": "d7mezwndk3yf3renn4fzeranpw",
        "team_id": "cj3ioc8zrinixx5erp94taidsc",
        "channel_id": "edet7g6h8ib8dftytcjcqne8ie",
        "post_id": "y4a6wgpr63gsdpq7cgoz8auimc",
        "mattermost_site_url": "https://mickmister.ngrok.io",
        "bot_access_token": "ouddbrqwupfypqsu1qxbdu3uqo",
        "acting_user_access_token": "yemrienc7pfypqsu1qxbdu3uqo",
        "admin_access_token": "ue8xi2sh7ebcciw8duww84ucme",
        "post": {
            "id": "y4a6wgpr63gsdpq7cgoz8auimc",
            "create_at": 1614222460141,
            "update_at": 1614222460141,
            "edit_at": 0,
            "delete_at": 0,
            "is_pinned": false,
            "user_id": "uzofd8otciyktj7mqbawi4hexc",
            "channel_id": "edet7g6h8ib8dftytcjcqne8ie",
            "root_id": "",
            "parent_id": "",
            "original_id": "",
            "message": "Created OAuth2 App (`sqqug7kmcb83ffdo8ryzhbi8ko`).",
            "type": "",
            "props": {
                "from_bot": "true"
            },
            "hashtags": "",
            "pending_post_id": "",
            "reply_count": 0
        }
    }
}
```

</details>


<details><summary>App Form Response</summary>


```json
{
    "type": "form",
    "form": {
        "title": "Send a survey to user",
        "header": "Message modal form header",
        "footer": "Message modal form footer",
        "call": {
            "url": "/send"
        },
        "fields": [
            {
                "name": "user_id",
                "type": "user",
                "value": "",
                "description": "User to send the survey to",
                "label": "user",
                "hint": "enter user ID or @user",
                "position": 1,
                "modal_label": "User",
                "refresh": true
            },
            {
                "name": "some_autocomplete_field",
                "type": "dynamic_select",
                "description": "Some Autocomplete Field",
                "label": "autocomplete",
                "hint": "Pick one",
                "modal_label": "Autocomplete"
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "Provisioned bot account @builtin (`ux3jks3kn7fz9ghnycx7iy5e7w`).",
                "description": "Text to ask the user about",
                "label": "message",
                "hint": "Anything you want to say",
                "modal_label": "Text",
                "subtype": "textarea",
                "min_length": 2,
                "max_length": 1024
            }
        ]
    }
}
```

</details>


## Flow, types, responses (WIP)

- Authentication
	- To App:
- Expanded: config (Mattermost Site URL)
- Lambda: none (invoked with AWS Invoke API, authenticated with IAM)
- HTTP: optional secret-based JWT

- Webhooks
  - Any third-party webhook calls will be proxied through the framework
  - Exact implementation here is TBD
  - Webhook handlers will only have access to Bot Access Token (no User/Sysadmin OAuth2)

- Notifications
  - Subscribe schema (per channel?)
  - All the p.API hooks available
  - Use the Call format (Expand business)
