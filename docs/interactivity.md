# Interactivity

This page shows the payloads for browser request/responses, and App server request/responses.

## Get Bindings

<details><summary>Request from Browser</summary>

`GET` http://localhost:8065/plugins/com.mattermost.apps/api/v1/bindings?user_id=mum5qskypidf3x3enkindgajrh&channel_id=zanqhwfdtjfi8yqyapd5qh6udh&scope=webapp
</details>

<details><summary>Response to Browser</summary>

```json
[
    {
        "app_id": "hello",
        "location": "/channel_header",
        "bindings": [
            {
                "app_id": "hello",
                "location": "send",
                "presentation": "modal",
                "icon": "https://www.clipartmax.com/png/middle/243-2431175_hello-hello-icon-png.png",
                "label": "Survey a user",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
                    "type": "form",
                    "expand": {
                        "post": "All"
                    }
                }
            }
        ]
    },
    {
        "app_id": "hello",
        "location": "/post_menu",
        "bindings": [
            {
                "app_id": "hello",
                "location": "send-me",
                "label": "Survey myself",
                "hint": "Send survey to myself",
                "description": "Send a customized emotional response survey to myself",
                "call": {
                    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send"
                }
            },
            {
                "app_id": "hello",
                "location": "send",
                "presentation": "modal",
                "label": "Survey a user",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
                    "type": "form",
                    "expand": {
                        "post": "All"
                    }
                }
            }
        ]
    },
    {
        "app_id": "hello",
        "location": "/command",
        "bindings": [
            {
                "app_id": "hello",
                "location": "message",
                "label": "message",
                "hint": "[--user] message",
                "description": "send a message to a user",
                "call": {
                    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send"
                }
            },
            {
                "app_id": "hello",
                "location": "manage",
                "label": "manage",
                "hint": "subscribe | unsubscribe ",
                "description": "manage channel subscriptions to greet new users",
                "bindings": [
                    {
                        "app_id": "hello",
                        "location": "subscribe",
                        "label": "subscribe",
                        "hint": "[--channel]",
                        "description": "subscribes a channel to greet new users",
                        "call": {
                            "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/subscribe",
                            "values": {
                                "mode": "on"
                            }
                        }
                    },
                    {
                        "app_id": "hello",
                        "location": "unsubscribe",
                        "label": "unsubscribe",
                        "hint": "[--channel]",
                        "description": "unsubscribes a channel from greeting new users",
                        "call": {
                            "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/subscribe",
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

`GET` http://localhost:8065/plugins/com.mattermost.apps/hello/bindings?acting_user_id=mum5qskypidf3x3enkindgajrh&channel_id=zanqhwfdtjfi8yqyapd5qh6udh

</details>

<details><summary>Response from App server</summary>

```json
[
    {
        "location": "/channel_header",
        "bindings": [
            {
                "location": "send",
                "icon": "https://www.clipartmax.com/png/middle/243-2431175_hello-hello-icon-png.png",
                "presentation": "modal",
                "label": "Survey a user",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
                    "type": "form",
                    "expand": {
                        "post": "All"
                    }
                }
            }
        ]
    },
    {
        "location": "/post_menu",
        "bindings": [
            {
                "location": "send-me",
                "label": "Survey myself",
                "hint": "Send survey to myself",
                "description": "Send a customized emotional response survey to myself",
                "call": {
                    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send"
                }
            },
            {
                "location": "send",
                "label": "Survey a user",
                "presentation": "modal",
                "hint": "Send survey to a user",
                "description": "Send a customized emotional response survey to a user",
                "call": {
                    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
                    "type": "form",
                    "expand": {
                        "post": "All"
                    }
                }
            }
        ]
    },
    {
        "location": "/command",
        "bindings": [
            {
                "location": "message",
                "label": "message",
                "hint": "[--user] message",
                "description": "send a message to a user",
                "call": {
                    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send"
                }
            },
            {
                "location": "manage",
                "label": "manage",
                "hint": "subscribe | unsubscribe ",
                "description": "manage channel subscriptions to greet new users",
                "bindings": [
                    {
                        "location": "subscribe",
                        "label": "subscribe",
                        "hint": "[--channel]",
                        "description": "subscribes a channel to greet new users",
                        "call": {
                            "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/subscribe",
                            "values": {
                                "mode": "on"
                            }
                        }
                    },
                    {
                        "location": "unsubscribe",
                        "label": "unsubscribe",
                        "hint": "[--channel]",
                        "description": "unsubscribes a channel from greeting new users",
                        "call": {
                            "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/subscribe",
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
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "form",
    "expand": {
        "post": "All"
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
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
                "value": "",
                "description": "User to send the survey to",
                "label": "user",
                "hint": "enter user ID or @user",
                "modal_label": "User",
                "refresh": true
            },
            {
                "name": "other",
                "type": "dynamic_select",
                "description": "Some values",
                "label": "other",
                "hint": "Pick one",
                "modal_label": "Other",
                "refresh": true
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "ee",
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

<details><summary>Request to App server</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/hello/send

```json
{
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "form",
    "context": {
        "app_id": "hello",
        "location": "send",
        "acting_user_id": "mum5qskypidf3x3enkindgajrh",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "post": {
            "id": "jgbzgehjqirkb8mn38axjktufw",
            "create_at": 1608279829636,
            "update_at": 1608279829636,
            "edit_at": 0,
            "delete_at": 0,
            "is_pinned": false,
            "user_id": "mum5qskypidf3x3enkindgajrh",
            "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
            "root_id": "",
            "parent_id": "",
            "original_id": "",
            "message": "ee",
            "type": "",
            "props": {},
            "hashtags": "",
            "pending_post_id": "",
            "reply_count": 0
        }
    },
    "expand": {
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
                "value": "",
                "description": "User to send the survey to",
                "label": "user",
                "hint": "enter user ID or @user",
                "modal_label": "User",
                "refresh": true
            },
            {
                "name": "other",
                "type": "dynamic_select",
                "value": null,
                "description": "Some values",
                "label": "other",
                "hint": "Pick one",
                "modal_label": "Other",
                "refresh": true
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "ee",
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

## Selected user in modal

<details><summary>Request from Browser</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/api/v1/call

```json
{
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "form",
    "expand": {
        "post": "All"
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh"
    },
    "values": {
        "name": "userID",
        "values": {
            "userID": "mum5qskypidf3x3enkindgajrh",
            "other": {
                "label": "Option 1",
                "value": "option1",
                "icon_data": ""
            },
            "message": "some text "
        },
        "value": "mum5qskypidf3x3enkindgajrh"
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
                "value": "mum5qskypidf3x3enkindgajrh",
                "description": "User to send the survey to",
                "label": "user",
                "hint": "enter user ID or @user",
                "modal_label": "User",
                "refresh": true
            },
            {
                "name": "other",
                "type": "dynamic_select",
                "value": {
                    "label": "Option 1",
                    "value": "option1",
                    "icon_data": ""
                },
                "description": "Some values",
                "label": "other",
                "hint": "Pick one",
                "modal_label": "Other",
                "refresh": true
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "some text  Now sending to mum5qskypidf3x3enkindgajrh.",
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

<details><summary>Request to App server</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/hello/send

```json
{
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "form",
    "values": {
        "name": "userID",
        "value": "mum5qskypidf3x3enkindgajrh",
        "values": {
            "message": "some text ",
            "other": {
                "icon_data": "",
                "label": "Option 1",
                "value": "option1"
            },
            "userID": "mum5qskypidf3x3enkindgajrh"
        }
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "acting_user_id": "mum5qskypidf3x3enkindgajrh",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "post": {
            "id": "jgbzgehjqirkb8mn38axjktufw",
            "create_at": 1608279829636,
            "update_at": 1608279829636,
            "edit_at": 0,
            "delete_at": 0,
            "is_pinned": false,
            "user_id": "mum5qskypidf3x3enkindgajrh",
            "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
            "root_id": "",
            "parent_id": "",
            "original_id": "",
            "message": "ee",
            "type": "",
            "props": {},
            "hashtags": "",
            "pending_post_id": "",
            "reply_count": 0
        }
    },
    "expand": {
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
                "value": "mum5qskypidf3x3enkindgajrh",
                "description": "User to send the survey to",
                "label": "user",
                "hint": "enter user ID or @user",
                "modal_label": "User",
                "refresh": true
            },
            {
                "name": "other",
                "type": "dynamic_select",
                "value": {
                    "icon_data": "",
                    "label": "Option 1",
                    "value": "option1"
                },
                "description": "Some values",
                "label": "other",
                "hint": "Pick one",
                "modal_label": "Other",
                "refresh": true
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "some text  Now sending to mum5qskypidf3x3enkindgajrh.",
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

## Dynamic Lookup

<details><summary>Request from Browser</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/api/v1/call

```json
{
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "lookup",
    "expand": {
        "post": "All"
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh"
    },
    "values": {
        "user_input": "opt",
        "values": {
            "userID": "u9sww9qs9pboudgxgpzny7e9we",
            "other": {
                "label": "Option 1",
                "value": "option1",
                "icon_data": ""
            },
            "message": "ee Now sending to u9sww9qs9pboudgxgpzny7e9we."
        },
        "name": "other"
    }
}
```

</details>

<details><summary>Response to Browser</summary>

```json
{
    "data": {
        "items": [
            {
                "icon_data": "",
                "label": "Option 1",
                "value": "option1"
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
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "lookup",
    "values": {
        "name": "other",
        "user_input": "opt",
        "values": {
            "message": "ee Now sending to u9sww9qs9pboudgxgpzny7e9we.",
            "other": {
                "icon_data": "",
                "label": "Option 1",
                "value": "option1"
            },
            "userID": "u9sww9qs9pboudgxgpzny7e9we"
        }
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "acting_user_id": "mum5qskypidf3x3enkindgajrh",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "post": {
            "id": "jgbzgehjqirkb8mn38axjktufw",
            "create_at": 1608279829636,
            "update_at": 1608279829636,
            "edit_at": 0,
            "delete_at": 0,
            "is_pinned": false,
            "user_id": "mum5qskypidf3x3enkindgajrh",
            "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
            "root_id": "",
            "parent_id": "",
            "original_id": "",
            "message": "ee",
            "type": "",
            "props": {},
            "hashtags": "",
            "pending_post_id": "",
            "reply_count": 0
        }
    },
    "expand": {
        "post": "All"
    }
}
```

</details>

<details><summary>Response from App server</summary>

```json
{
    "data": {
        "items": [
            {
                "label": "Option 1",
                "value": "option1",
                "icon_data": ""
            }
        ]
    }
}
```

</details>

## Dynamic value selected

<details><summary>Request from Browser</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/api/v1/call

```json
{
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "form",
    "expand": {
        "post": "All"
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh"
    },
    "values": {
        "name": "other",
        "values": {
            "userID": null,
            "other": {
                "icon_data": "",
                "label": "Option 1",
                "value": "option1"
            },
            "message": "ee"
        },
        "value": {
            "icon_data": "",
            "label": "Option 1",
            "value": "option1"
        }
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
                "value": "",
                "description": "User to send the survey to",
                "label": "user",
                "hint": "enter user ID or @user",
                "modal_label": "User",
                "refresh": true
            },
            {
                "name": "other",
                "type": "dynamic_select",
                "value": {
                    "label": "Option 1",
                    "value": "option1",
                    "icon_data": ""
                },
                "description": "Some values",
                "label": "other",
                "hint": "Pick one",
                "modal_label": "Other",
                "refresh": true
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "ee",
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

<details><summary>Request to App server</summary>

`POST` http://localhost:8065/plugins/com.mattermost.apps/hello/send

```json
{
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "form",
    "values": {
        "name": "other",
        "value": {
            "icon_data": "",
            "label": "Option 1",
            "value": "option1"
        },
        "values": {
            "message": "ee Now sending to u9sww9qs9pboudgxgpzny7e9we.",
            "other": {
                "icon_data": "",
                "label": "Option 1",
                "value": "option1"
            },
            "userID": "u9sww9qs9pboudgxgpzny7e9we"
        }
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "acting_user_id": "mum5qskypidf3x3enkindgajrh",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "post": {
            "id": "jgbzgehjqirkb8mn38axjktufw",
            "create_at": 1608279829636,
            "update_at": 1608279829636,
            "edit_at": 0,
            "delete_at": 0,
            "is_pinned": false,
            "user_id": "mum5qskypidf3x3enkindgajrh",
            "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
            "root_id": "",
            "parent_id": "",
            "original_id": "",
            "message": "ee",
            "type": "",
            "props": {},
            "hashtags": "",
            "pending_post_id": "",
            "reply_count": 0
        }
    },
    "expand": {
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
                "value": "u9sww9qs9pboudgxgpzny7e9we",
                "description": "User to send the survey to",
                "label": "user",
                "hint": "enter user ID or @user",
                "modal_label": "User",
                "refresh": true
            },
            {
                "name": "other",
                "type": "dynamic_select",
                "value": {
                    "icon_data": "",
                    "label": "Option 1",
                    "value": "option1"
                },
                "description": "Some values",
                "label": "other",
                "hint": "Pick one",
                "modal_label": "Other",
                "refresh": true
            },
            {
                "name": "message",
                "type": "text",
                "is_required": true,
                "value": "ee Now sending to u9sww9qs9pboudgxgpzny7e9we.",
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
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "type": "",
    "expand": {
        "post": "All"
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh"
    },
    "values": {
        "userID": "u4bkq9ch67doxmix9namyk6qfe",
        "other": {
            "label": "Option 1",
            "value": "option1",
            "icon_data": ""
        },
        "message": "ee Now sending to u4bkq9ch67doxmix9namyk6qfe."
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
    "url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/hello/send",
    "values": {
        "message": "ee",
        "other": null,
        "userID": null
    },
    "context": {
        "app_id": "hello",
        "location": "send",
        "acting_user_id": "mum5qskypidf3x3enkindgajrh",
        "team_id": "qe5ken7k9f8rdp5bqnfhhs5nzy",
        "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
        "post_id": "jgbzgehjqirkb8mn38axjktufw",
        "post": {
            "id": "jgbzgehjqirkb8mn38axjktufw",
            "create_at": 1608279829636,
            "update_at": 1608279829636,
            "edit_at": 0,
            "delete_at": 0,
            "is_pinned": false,
            "user_id": "mum5qskypidf3x3enkindgajrh",
            "channel_id": "zanqhwfdtjfi8yqyapd5qh6udh",
            "root_id": "",
            "parent_id": "",
            "original_id": "",
            "message": "ee",
            "type": "",
            "props": {},
            "hashtags": "",
            "pending_post_id": "",
            "reply_count": 0
        }
    },
    "expand": {
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
