# Post Menu Example 

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
            "message": "Created Mattermost OAuth2 App (`sqqug7kmcb83ffdo8ryzhbi8ko`).",
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
                "value": "Using bot account @builtin (`ux3jks3kn7fz9ghnycx7iy5e7w`).",
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

<details><summary>Diagram Source</summary>

https://sequencediagram.org

```
title Bindings + Form Example

Client->MM: Visit ChannelA, fetch bindings (Client Bindings Request)
MM->App:Fetch bindings with call (MM Bindings Request)
App->3rd Party Integration:Check user status
3rd Party Integration->App:Return user status
App->MM:Return bindings (App Bindings Response)
MM->Client:Return bindings, render App's post menu item
Client->MM:Clicked post menu item. Perform submit call (Client Submit Request)
MM->App:Perform submit call (MM Submit Request)
App->3rd Party Integration:Do something useful
3rd Party Integration->App:Return something useful
App->MM:Return new modal form (App Form Response)
MM->Client:Return modal form, open modal
```

</details>