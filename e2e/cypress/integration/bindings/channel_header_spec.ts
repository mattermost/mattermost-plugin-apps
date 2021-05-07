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
const installAppCommand = '/apps install http-hello --app-secret 1234';

describe('Apps bindings - Channel header', () => {
    let testTeam;

    before(() => {
        const newSettings = {
            PluginSettings: {
                Enable: true,
            },
            ServiceSettings: {
                EnableOAuthServiceProvider: true,
                EnableBotAccountCreation: true,
                EnableTesting: true,
                EnableDeveloper: true,
            },
        };

        cy.apiUpdateConfig(newSettings);
        cy.apiInitSetup().then(({team}) => {
            testTeam = team;
        });
    });

    it('MM-32330 Bindings - Channel header submit', () => {
        cy.visit(`/${testTeam.name}/channels/town-square`);

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

        // # Visit http-hello DM channel
        cy.get('a.SidebarLink[aria-label*="http-hello"]').click();

        // * Verify survey content
        cy.getLastPostId().then((postID) => {
            const postIDSelector = '#post_' + postID;
            cy.get(`${postIDSelector} .attachment .attachment__title`).should('have.text', 'Survey');
            cy.get(`${postIDSelector} .attachment .post-message__text-container`).should('have.text', 'The message');
        });
    });
});

const runCommand = (command: string) => {
    cy.get('#post_textbox').clear().type(command);
    cy.get('#post_textbox').type('{enter}');
};

const installHTTPHello = () => {
    runCommand(addManifestCommand);
    runCommand(installAppCommand);
    cy.get('#interactiveDialogSubmit').click();
};
