// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
const fetch = require('node-fetch');
const express = require('express');

const app = express();
app.use(express.json());
const host = 'localhost';
const port = 8080;

app.get('/manifest.json', (req, res) => {
    res.json({
        app_id: 'hello-world',
        display_name: 'Hello, world!',
        app_type: 'http',
        icon: 'icon.png',
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
                        icon: 'icon.png',
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
                        icon: 'icon.png',
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
                            {
                                location: 'issue',
                                label: 'issue',
                                call: {
                                    path: '/issue',
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
            icon: 'icon.png',
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

app.post('/issue/form', (req, res) => {
    res.json({
        type: 'form',
        form: {
            title: 'Reproduce MM-37429',
            icon: 'icon.png',
            call: {
                path: '/issue',
            },
        },
    });
});

app.get('/static/icon.png', (req, res) => {
    res.sendFile(__dirname + '/icon.png');
});

app.post('/send/submit', async (req, res) => {
    const call = req.body;

    let message = 'Hello, world!';
    const submittedMessage = call.values.message;
    if (submittedMessage) {
        message += ' ...and ' + submittedMessage + '!';
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
    const mattermostSiteURL = call.context.mattermost_site_url;

    const channel = await fetch(mattermostSiteURL + '/api/v4/channels/direct', options).
        then((r) => r.json());

    const post = {
        channel_id: channel.id,
        message,
    };

    // Create a post
    options.body = JSON.stringify(post);

    await fetch(mattermostSiteURL + '/api/v4/posts', options);


    res.json({
        type: 'ok',
        markdown: 'Created a post in your DM channel.'
    });
});

app.post('/issue/submit', async (req, res) => {
    const call = req.body;

    const users = [
        call.context.bot_user_id,
        call.context.acting_user_id,
    ];

    const message = `Users: ![Test User](https://avatars.slack-edge.com/2019-06-13/665302393639_6ee45a4c8e1342572d3e_192.jpg =25 "Test User") ![Test User2](https://secure.gravatar.com/avatar/bd9a02d5518b55c2a6b85a5dcda9f6e1.jpg?s=192&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-192.png =25 "Test User2")`;

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
    const mattermostSiteURL = call.context.mattermost_site_url;

    const channel = await fetch(mattermostSiteURL + '/api/v4/channels/direct', options).
        then((r) => r.json());

    const post = {
        channel_id: channel.id,
        message,
    };

    // Create a post
    options.body = JSON.stringify(post);

    await fetch(mattermostSiteURL + '/api/v4/posts', options);


    res.json({
        type: 'ok',
        markdown: message
    });
});

app.listen(port, host, () => {
    console.log(`hello-world app listening at http://${host}:${port}`);
});
