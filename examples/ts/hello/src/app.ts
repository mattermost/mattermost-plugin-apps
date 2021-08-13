// Copyright (c) 2021-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import express from 'express';
import serverless from 'serverless-http';

const app = new express();
app.use(express.json());
app.use('/static', express.static('./static'));

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

export default app;