# Interactivity

This page shows the payloads for browser request/responses, and App server request/responses.

## Get Bindings

<details><summary>Request from Browser</summary>

`GET` http://localhost:8065/plugins/com.mattermost.apps/api/v1/bindings?user_id=f88nesmr7ifhicr8wwf94oxiwa&channel_id=b77x9mu8xindzxkgem8guadsra&scope=webapp
</details>

<details><summary>Response to Browser</summary>

```json
[
    {
        "location_id": "/channel_header",
        "bindings": [
            {
                "app_id": "hello",
                "location_id": "send",
                "icon": "https://raw.githubusercontent.com/mattermost/mattermost-plugin-jira/master/assets/icon.svg",
                "label": "Survey a user",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send"
                }
            }
        ]
    },
    {
        "location_id": "/post_menu",
        "bindings": [
            {
                "app_id": "hello",
                "location_id": "send-me",
                "label": "Survey myself",
                "hint": "Send survey to myself",
                "description": "Send a customized emotional response survey to myself",
                "call": {
                    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send"
                }
            },
            {
                "app_id": "hello",
                "location_id": "send",
                "label": "Survey a user",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send",
                    "type": "form",
                    "expand": {
                        "app": "",
                        "acting_user": "",
                        "post": "All"
                    }
                }
            }
        ]
    },
    {
        "location_id": "/command",
        "bindings": [
            {
                "app_id": "hello",
                "label": "message",
                "hint": "[--user] message",
                "description": "send a message to a user",
                "call": {
                    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send"
                }
            },
            {
                "app_id": "hello",
                "location_id": "manage",
                "hint": "subscribe | unsubscribe ",
                "description": "manage channel subscriptions to greet new users",
                "bindings": [
                    {
                        "app_id": "hello",
                        "label": "subscribe",
                        "hint": "[--channel]",
                        "description": "subscribes a channel to greet new users",
                        "call": {
                            "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/subscribe",
                            "values": {
                                "mode": "on"
                            }
                        }
                    },
                    {
                        "app_id": "hello",
                        "label": "unsubscribe",
                        "hint": "[--channel]",
                        "description": "unsubscribes a channel from greeting new users",
                        "call": {
                            "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/subscribe",
                            "values": {
                                "mode": "off"
                            }
                        }
                    }
                ]
            }
        ]
    }
]
```

</details>

<details><summary>Request to App server</summary>

`GET` http://localhost:8065/plugins/com.mattermost.apps/hello/bindings?acting_user_id=f88nesmr7ifhicr8wwf94oxiwa&channel_id=b77x9mu8xindzxkgem8guadsra

</details>

<details><summary>Response from App server</summary>

```json
[
    {
        "location_id": "/channel_header",
        "bindings": [
            {
                "location_id": "send",
                "icon": "https://raw.githubusercontent.com/mattermost/mattermost-plugin-jira/master/assets/icon.svg",
                "label": "Survey a user",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send"
                }
            }
        ]
    },
    {
        "location_id": "/post_menu",
        "bindings": [
            {
                "location_id": "send-me",
                "label": "Survey myself",
                "hint": "Send survey to myself",
                "description": "Send a customized emotional response survey to myself",
                "call": {
                    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send"
                }
            },
            {
                "location_id": "send",
                "label": "Survey a user",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send",
                    "type": "form"
                }
            }
        ]
    },
    {
        "location_id": "/command",
        "bindings": [
            {
                "label": "message",
                "hint": "[--user] message",
                "description": "send a message to a user",
                "call": {
                    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send"
                }
            },
            {
                "location_id": "manage",
                "hint": "subscribe | unsubscribe ",
                "description": "manage channel subscriptions to greet new users",
                "bindings": [
                    {
                        "label": "subscribe",
                        "hint": "[--channel]",
                        "description": "subscribes a channel to greet new users",
                        "call": {
                            "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/subscribe",
                            "values": {
                                "mode": "on"
                            }
                        }
                    },
                    {
                        "label": "unsubscribe",
                        "hint": "[--channel]",
                        "description": "unsubscribes a channel from greeting new users",
                        "call": {
                            "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/subscribe",
                            "values": {
                                "mode": "off"
                            }
                        }
                    }
                ]
            }
        ]
    }
]
```
</details>

## Clicked Post Dropdown

<details><summary>Request from Browser</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/api/v1/call

```json
{
    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send",
    "type": "form",
    "expand": {
        "app": "",
        "acting_user": "",
        "post": "All"
    },
    "context": {
        "app_id": "hello",
        "location_id": "send",
        "team_id": "gzt3wbyu4tnqfmiqchyatzkx1r",
        "channel_id": "b77x9mu8xindzxkgem8guadsra",
        "post_id": "q389yjyqfpygtedmknkannck8c",
        "root_id": ""
    }
}
```

</details>

<details><summary>Response to Browser</summary>

```json
{
    "type": "form",
    "form": {
        "title": "Send a survey to user",
        "header": "Message modal form header",
        "footer": "Message modal form footer",
        "fields": [
            {
                "name": "userID",
                "type": "user",
                "description": "User to send the survey to",
                "label": "User",
                "hint": "enter user ID or @user",
                "position": 1,
                "modal_label": "User"
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "Here is the value of your post",
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

<details><summary>Request to App server</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/hello/send

```json
{
    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send",
    "type": "form", // Notice type == "form". This means it is a request to fetch a form's definition.
    "context": {
        "app_id": "hello",
        "location_id": "send",
        "acting_user_id": "f88nesmr7ifhicr8wwf94oxiwa",
        "team_id": "gzt3wbyu4tnqfmiqchyatzkx1r",
        "channel_id": "b77x9mu8xindzxkgem8guadsra",
        "post_id": "q389yjyqfpygtedmknkannck8c",
        "post": {
            "id": "q389yjyqfpygtedmknkannck8c",
            "create_at": 1605082890776,
            "update_at": 1605082890776,
            "edit_at": 0,
            "delete_at": 0,
            "is_pinned": false,
            "user_id": "f88nesmr7ifhicr8wwf94oxiwa",
            "channel_id": "b77x9mu8xindzxkgem8guadsra",
            "root_id": "",
            "parent_id": "",
            "original_id": "",
            "message": "Here's the body of the post",
            "type": "",
            "props": {
                "disable_group_highlight": true
            },
            "hashtags": "",
            "pending_post_id": "",
            "reply_count": 0
        }
    },
    "expand": {
        "app": "",
        "acting_user": "",
        "post": "All"
    }
}
```

</details>

<details><summary>Response from App server</summary>

```json
{
    "type": "form",
    "form": {
        "title": "Send a survey to user",
        "header": "Message modal form header",
        "footer": "Message modal form footer",
        "fields": [
            {
                "name": "userID",
                "type": "user",
                "description": "User to send the survey to",
                "label": "User",
                "hint": "enter user ID or @user",
                "position": 1,
                "modal_label": "User"
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "Here's the body of the post",
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

## Submitted Modal

<details><summary>Request from Browser</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/api/v1/call
```json
{
    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send",
    "type": "",
    "expand": {
        "app": "",
        "acting_user": "",
        "post": "All"
    },
    "context": {
        "app_id": "hello",
        "location_id": "send",
        "team_id": "gzt3wbyu4tnqfmiqchyatzkx1r",
        "channel_id": "b77x9mu8xindzxkgem8guadsra",
        "post_id": "q389yjyqfpygtedmknkannck8c",
        "root_id": ""
    },
    "values": {
        "userID": "crwxjq4kk7fk3qztocjzenkppc",
        "message": "Here's the edited body of the post"
    }
}
```

</details>

<details><summary>Response to Browser</summary>

```json
{"markdown": "Successfully sent survey"}
```

</details>

<details><summary>Request to App server</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/hello/send

```json
{
    // Notice type == "", or missing. This means it is a form submission.
    "url": "http://localhost:8065/plugins/com.mattermost.apps/hello/send",
    "values": {
        "message": "Here's the body of the post",
        "userID": "crwxjq4kk7fk3qztocjzenkppc"
    },
    "context": {
        "app_id": "hello",
        "location_id": "send",
        "acting_user_id": "f88nesmr7ifhicr8wwf94oxiwa",
        "team_id": "gzt3wbyu4tnqfmiqchyatzkx1r",
        "channel_id": "b77x9mu8xindzxkgem8guadsra",
        "post_id": "q389yjyqfpygtedmknkannck8c",
        "post": {
            "id": "q389yjyqfpygtedmknkannck8c",
            "create_at": 1605082890776,
            "update_at": 1605082890776,
            "edit_at": 0,
            "delete_at": 0,
            "is_pinned": false,
            "user_id": "f88nesmr7ifhicr8wwf94oxiwa",
            "channel_id": "b77x9mu8xindzxkgem8guadsra",
            "root_id": "",
            "parent_id": "",
            "original_id": "",
            "message": "Here's the body of the post",
            "type": "",
            "props": {
                "disable_group_highlight": true
            },
            "hashtags": "",
            "pending_post_id": "",
            "reply_count": 0
        }
    },
    "expand": {
        "app": "",
        "acting_user": "",
        "post": "All"
    }
}
```

</details>

<details><summary>Response from App server</summary>

```json
{"markdown": "Successfully sent survey"}
```

</details>
