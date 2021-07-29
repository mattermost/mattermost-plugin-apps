// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import express from 'express';

// Shim for mattermost-redux global fetch access
global.fetch = require('node-fetch');

import {AppBinding, AppCallRequest, AppCallResponse, AppForm, AppManifest} from 'mattermost-redux/types/apps';
import {Post} from 'mattermost-redux/types/posts';
import {Channel} from 'mattermost-redux/types/channels';

import Client4 from 'mattermost-redux/client/client4';

const host = process.env.NODE_HOST || 'localhost';
const port = process.env.PORT || 4000;

const app = express();
app.use(express.json());

// Uncomment these lines to enable verbose debugging of requests and responses
// import logger from './middleware/logger';
// app.use(logger);

app.use((req, res, next) => {
    const call: AppCallRequest = req.body;

    // This is used to interact with the Mattermost server in the docker-compose dev environment.
    // We ignore the site URL sent in call requests, and instead use the known site URL from the environment variable.
    if (call?.context?.mattermost_site_url && process.env.MATTERMOST_SITEURL) {
        call.context.mattermost_site_url = process.env.MATTERMOST_SITEURL;
    }

    next();
});

const manifest = {
    app_id: 'node-example',
    display_name: "I'm an App!",
    homepage_url: 'https://github.com/mattermost/mattermost-plugin-apps/dev',
    app_type: 'http',
    icon: 'icon.png',
    root_url: `http://${host}:${port}`,
    requested_permissions: [
        'act_as_bot',
    ],
    requested_locations: [
        '/channel_header',
        '/command',
    ],
} as AppManifest;

const form: AppForm = {
    title: "I'm a form!",
    icon: 'icon.png',
    fields: [
        {
            type: 'text',
            name: 'message',
            label: 'message',
            position: 1,
        },
    ],
    call: {
        path: '/send',
    },
};

const channelHeaderBindings = {
    location: '/channel_header',
    bindings: [
        {
            location: 'send-button',
            icon: 'icon.png',
            label: 'send hello message',
            call: {
                path: '/send',
            },
        },
    ],
} as AppBinding;

const commandBindings = {
    location: '/command',
    bindings: [
        {
            icon: 'icon.png',
            label: 'node-example',
            description: 'Example App written with Node.js',
            hint: '[send]',
            bindings: [
                {
                    location: 'send',
                    label: 'send',
                    form,
                },
            ],
        },
    ],
} as AppBinding;

// Serve icon.png and others from the static folder
app.use('/static', express.static('./static'));

app.get('/manifest.json', (req, res) => {
    res.json(manifest);
});

app.post('/bindings', (req, res) => {
    const callResponse: AppCallResponse<AppBinding[]> = {
        type: 'ok',
        data: [
            channelHeaderBindings,
            commandBindings,
        ],
    };

    res.json(callResponse);
});

type FormValues = {
    message: string;
}

app.post('/send/submit', async (req, res) => {
    const call = req.body as AppCallRequest;

    const botClient = new Client4();
    botClient.setUrl(call.context.mattermost_site_url);
    botClient.setToken(call.context.bot_access_token);

    const formValues = call.values as FormValues;

    let message = 'Hello, world!';
    const submittedMessage = formValues.message;
    if (submittedMessage) {
        message += ' ...and ' + submittedMessage + '!';
    }

    const users = [
        call.context.bot_user_id,
        call.context.acting_user_id,
    ] as string[];

    let channel: Channel;
    try {
        channel = await botClient.createDirectChannel(users);
    } catch (e) {
        res.json({
            type: 'error',
            error: 'Failed to create/fetch DM channel: ' + e.message,
        });
        return;
    }

    const post = {
        channel_id: channel.id,
        message,
    } as Post;

    try {
        await botClient.createPost(post)
    } catch (e) {
        res.json({
            type: 'error',
            error: 'Failed to create post in DM channel: ' + e.message,
        });
        return;
    }

    const callResponse: AppCallResponse = {
        type: 'ok',
        markdown: 'Created a post in your DM channel.',
    };

    res.json(callResponse);
});

app.listen(port, () => {
    console.log(`app listening on port ${port}`);
});
