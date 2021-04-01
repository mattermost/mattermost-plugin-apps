// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
// <reference path="../support/index.d.ts" />

// ***************************************************************
// - [#] indicates a test step (e.g. # Go to a page)
// - [*] indicates an assertion (e.g. * Check the title)
// - Use element ID when selecting an element. Create one if none.
// ***************************************************************

import {verifyEphemeralMessage} from 'mattermost-webapp/e2e/cypress/integration/integrations/builtin_commands/helper';

const baseURL = Cypress.config('baseUrl');
const pluginID = Cypress.config('pluginID');
const helloManifestRoute = 'example/hello/mattermost-app.json';

const addManifestCommand = `/apps debug-add-manifest --url ${baseURL}/plugins/${pluginID}/${helloManifestRoute}`;
const installAppCommand = '/apps install --app-id http-hello --app-secret 1234';

describe('Apps bindings - Channel header', () => {
    before(() => {
        const newSettings = {
            PluginSettings: {
                Enable: true,
            },
            ServiceSettings: {
                EnableOAuthServiceProvider: true,
                EnableTesting: true,
                EnableDeveloper: true,
            },
        };

        cy.apiUpdateConfig(newSettings);
        cy.apiInitSetup();
    });

    it('MM-32330 Bindings - Channel header submit', () => {
        cy.visit('/');

        // # Install the http-hello app
        installHTTPHello();

        // # Open the apps modal by clicking on a channel header binding
        cy.get('#channel-header #send').first().click();

        // # Type into message field of modal form
        cy.findByTestId('message').type('The message');

        // # Submit modal form
        cy.get('#appsModalSubmit').click();

        // * Verify ephemeral message
        verifyEphemeralMessage('Successfully sent survey');
    });
});

const runCommand = (command: string) => {
    cy.get('#post_textbox').clear().type(command);
    cy.get('#post_textbox').type('{enter}');
}

const installHTTPHello = () => {
    runCommand(addManifestCommand);
    runCommand(installAppCommand);
    cy.get('#interactiveDialogSubmit').click();
}
