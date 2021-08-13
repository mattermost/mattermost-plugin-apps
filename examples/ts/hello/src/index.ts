// Copyright (c) 2021-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import express from 'express';
import serverless from 'serverless-http';

import app from './app';

if (isRunningInHTTPMode()) {
    // Listen to http port
    const port = getPort();
    app.listen(port, () => console.log('Listening on ' + port));
}

export function isRunningInHTTPMode(): boolean {
    return process.env.LOCAL === 'true';
}

function getPort(): number {
    return Number(process.env.PORT) || 4000;
}

export function getHTTPPath(): string {
    return 'http://localhost:' + getPort();
}

const shandler = serverless(app);

module.exports.kubeless = async (event, context) => {
    return shandler(event.data, context);
};

module.exports.aws = async (event, context) => {
    return shandler(event, context);
};