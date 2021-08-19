// Copyright (c) 2021-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

const fetch = require('node-fetch');
const express = require('express');
const serverless = require('serverless-http');

const app = new express();
app.use(express.json());
const shandler = serverless(app);

module.exports.handler = async (event, context) => {
    return await shandler(event.data, context);
};

app.post('/ping', (req, res) => {
    res.json({
        type: 'ok',
        markdown: 'PONG',
    })
})

// app.get('/manifest.json', (req, res) => {
//     res.json({
//         app_id: "hello-kubeless",
//         display_name: "Hello, kubeless world!",
//         app_type: "kubeless",
//         version: "demo",
//         homepage_url: "https://www.mattermost.com",
//         root_url: "http://localhost:8080",
//         kubeless_functions: [
//             {
//                 call_path: "/",
//                 handler: "app.Handler",
//                 file: "app.js",
//                 deps_file: "package.json",
//                 runtime: "nodejs14"
//             }
//         ],
//         requested_permissions: [
//             "act_as_bot"
//         ],
//         requested_locations: [
//             "/command"
//         ]
//     });
// });

app.post('/bindings', (req, res) => {
    res.json({
        type: 'ok',
        data: [
            {
                location: '/command',
                bindings: [
                    {
                        label: 'hello-kubeless',
                        description: 'Hello Kubeless World app',
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

app.post(['/send/form'], (req, res) => {
    res.json({
        type: 'form',
        form: {
            title: 'Hello, world!',
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
    const submittedMessage = call.values?.message;
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