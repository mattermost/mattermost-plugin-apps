// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
const fetch = require('node-fetch');
const express = require('express');

const app = express();
const host = 'localhost';
const port = 8080;

app.get('/manifest.json', (req, res) => {
    res.json({
        app_id: 'hello-world',
        display_name: 'Hello, world!',
        app_type: 'http',
        root_url: 'http://localhost:8080',
        requested_permissions: [
            'act_as_bot',
        ],
        requested_locations: [
            '/channel_header',
            '/command',
        ],
    });
});

app.post('/bindings', (req, res) => {
    res.json({
        type: 'ok',
        data: [
            {
                location: '/channel_header',
                bindings: [
                    {
                        location: 'send-button',
                        icon: 'http://localhost:8080/static/icon.png',
                        label: 'send hello message',
                        call: {
                            path: '/send-modal',
                        },
                    },
                ],
            },
            {
                location: '/command',
                bindings: [
                    {
                        icon: 'http://localhost:8080/static/icon.png',
                        label: 'helloworld',
                        description: 'Hello World app',
                        hint: '[send]',
                        bindings: [
                            {
                                location: 'send',
                                label: 'send',
                                call: {
                                    path: '/send',
                                },
                            },
                        ],
                    },
                ],
            },
        ],
    });
});

app.post(['/send/form', '/send-modal/submit'], (req, res) => {
    res.json({
        type: 'form',
        form: {
            title: 'Hello, world!',
            icon: 'http://localhost:8080/static/icon.png',
            fields: [
                {
                    type: 'text',
                    name: 'message',
                    label: 'message',
                },
            ],
            call: {
                path: '/send',
            },
        },
    });
});

app.get('/static/icon.png', (req, res) => {
    res.sendFile(__dirname + '/icon.png');
});

app.use(express.json());

app.post('/send/submit', (req, res) => {
    const call = req.body;

    let message = 'Hello, world!';
    const user_message = call.values.message;
    if (user_message) {
        message += ' ...and ' + user_message + '!';
    }

    const users = [
        call.context.bot_user_id,
        call.context.acting_user_id,
    ];

    // Use the app bot to do API calls
    const options = {
        method: 'POST',
        headers: {
            Authorization: 'BEARER ' + call.context.bot_access_token,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(users),
    };

    // Get the DM channel between the user and the bot
    const mattermost_site_url = call.context.mattermost_site_url;

    fetch(mattermost_site_url + '/api/v4/channels/direct', options).
        then((mm_res) => mm_res.json()).
        then((channel) => {
            const post = {
                channel_id: channel.id,
                message,
            };

            // Create a post
            options.body = JSON.stringify(post);

            fetch(mattermost_site_url + '/api/v4/posts', options);
        });

    res.json({});
});

app.listen(port, host, () => {
    console.log(`hello-world app listening at http://${host}:${port}`);
});
