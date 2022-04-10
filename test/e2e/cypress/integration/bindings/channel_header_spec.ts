// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
// <reference path="../support/index.d.ts" />

// ***************************************************************
// - [#] indicates a test step (e.g. # Go to a page)
// - [*] indicates an assertion (e.g. * Check the title)
// - Use element ID when selecting an element. Create one if none.
// ***************************************************************

import {verifyEphemeralMessage} from 'mattermost-webapp/e2e/cypress/integration/integrations/builtin_commands/helper';

const testAppHost = Cypress.config('testAppHost');
const testAppManifestPath = `${testAppHost}/manifest.json`;

const installAppCommand = `/apps install http ${testAppManifestPath}`;

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
        const tt = cy.get('#channel-header img[src="http://localhost:8065/plugins/com.mattermost.apps/apps/test/static/icon.png"]').first();
        console.log('<>/<> ', tt);
        cy.get('#channel-header img[src="http://localhost:8065/plugins/com.mattermost.apps/apps/test/static/icon.png"]').first().click();

        // # Type into message field of modal form
        cy.findByTestId('message').type('the test message');

        // # Submit modal form
        cy.get('#appsModalSubmit').click();

        // * Verify ephemeral message
        verifyEphemeralMessage('Created a post in your DM channel.');

        // # Visit http-hello DM channel
        cy.get('a.SidebarLink[aria-label*="hello-world"]').click();

        // * Verify survey content
        cy.getLastPostId().then((postID) => {
            const postIDSelector = '#post_' + postID;
            cy.get(`${postIDSelector} .post__body`).should('have.text', 'Hello, world! ...and the test message!');
        });
    });
});

const runCommand = (command: string) => {
    cy.get('#post_textbox').clear().type(command + ' ');
    cy.get('#post_textbox').type('{enter}');
};

const installHTTPHello = () => {
    runCommand(installAppCommand);
    cy.get('#consent').click();
    cy.get('#appsModalSubmit').click();
};
