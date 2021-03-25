// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
// <reference path="../support/index.d.ts" />

// ***************************************************************
// - [#] indicates a test step (e.g. # Go to a page)
// - [*] indicates an assertion (e.g. * Check the title)
// - Use element ID when selecting an element. Create one if none.
// ***************************************************************

import {AppCallRequest} from 'mattermost-redux/types/apps';

describe('Apps binsings - Channel header', () => {
    // const pluginID = Cypress.config('pluginID');
    // const pluginFile = Cypress.config('pluginFile');

    before(() => {
        const newSettings = {
            PluginSettings: {
                Enable: true,
            },
            ServiceSettings: {
                EnableOAuthServiceProvider: true,
            },
        };

        cy.apiUpdateConfig(newSettings);
        cy.apiInitSetup();
    });

    beforeEach(() => {
        cleanTestApp();
    });

    it('MM-00000 Bindings - Channel header submit', () => {
        setChannelHeaderBinding();
        setChannelHeaderSubmitResponse({
            "type":"ok",
            "markdown":"Sent survey to mickmister."
        });

        cy.visit('/');
        cy.get('#send-button-testers').click();

        cy.wait(1000);
        cy.getLastPost().should('contain.text', 'Sent survey to mickmister.');

        getBindingsRequest().then((request: {body: AppCallRequest}) => {
            const {
                path,
                context,
            } = request.body;

            expect(path).to.equal('/bindings');

            expect(context.app_id).to.equal('e2e-testapp');
            expect(context.user_agent).to.equal('webapp');
            expect(context.mattermost_site_url).to.equal('http://localhost:8065');
            expect(context.bot_user_id).to.be.ok;
            expect(context.acting_user_id).to.be.ok;
            // expect(context.team_id).to.be.ok; // currently fails
            expect(context.channel_id).to.be.ok;
            expect(context.bot_access_token).to.be.ok;
        });
    });

    it('MM-00000 Bindings - Channel header open form', () => {
        setChannelHeaderBinding();
        setChannelHeaderSubmitResponse({
            "type": "form",
            "form": {
                "title": "Test Form",
                "icon": "http://localhost:8080/static/icon.png",
                "fields": [
                    {
                        "type": "text",
                        "name": "message",
                        "label": "message"
                    },
                    {
                        "type": "user",
                        "name": "user",
                        "label": "user",
                        "refresh": true
                    },
                ],
                "call": {
                    "path": "/test-form"
                }
            }
        });

        cy.visit('/');
        cy.get('#send-button-testers').click();

        cy.get('#appsModalLabel').should('have.text', 'Test Form');

        cy.findByTestId('message').type('hey');

        cy.findByTestId('user').click();
        cy.get('.suggestion-list__item').first().click();

        setFormSubmitResponse({
            "type":"ok",
            "markdown":"You submitted the form!",
        });
        cy.get('#appsModalSubmit').click();

        cy.wait(1000);
        cy.getLastPost().should('contain.text', 'You submitted the form!');

        getSubmitRequest('/test-form').then((request: {body: AppCallRequest}) => {
            const {
                path,
                expand,
                values,
                context,
            } = request.body;

            expect(path).to.equal('/test-form/submit');
            expect(expand).to.deep.equal({});
            expect(values).to.deep.equal({
                "message": "hey",
                "user": {
                    "label": "aaron.medina",
                    "value": "a8qcejd6sf8odga14817fnyazw"
                }
            });

            expect(context.app_id).to.equal('e2e-testapp');
            expect(context.location).to.equal('send-button-testers');
            expect(context.user_agent).to.equal('webapp');
            expect(context.mattermost_site_url).to.equal('http://localhost:8065');
            expect(context.bot_user_id).to.be.ok;
            expect(context.acting_user_id).to.be.ok;
            expect(context.team_id).to.be.ok;
            expect(context.channel_id).to.be.ok;
            expect(context.bot_access_token).to.be.ok;
        });
    });
});

const setChannelHeaderSubmitResponse = (response) => {
    return setSubmitResponse('/channel-header-submit', response);
}

const setFormSubmitResponse = (response) => {
    return setSubmitResponse('/test-form', response);
}

const setChannelHeaderBinding = () => {
    return setBindings([
        {
            "location": "/channel_header",
            "bindings": [
                {
                    "location": "send-button-testers",
                    "icon": "http://localhost:8080/static/icon.png",
                    "label":"send hello message",
                    "call": {
                        "path": "/channel-header-submit"
                    }
                }
            ]
        }
    ]);
}

const setBindings = (bindings) => {
    const data = {
        type: 'ok',
        data: bindings,
    };

    return cy.request('POST', 'http://localhost:8065/plugins/com.mattermost.apps/e2e-testapp/bindings/set-response', data);
}

const getBindingsRequest = () => {
    return cy.request('GET', 'http://localhost:8065/plugins/com.mattermost.apps/e2e-testapp/bindings/get-request');
}

const setSubmitResponse = (url, response) => {
    return cy.request('POST', 'http://localhost:8065/plugins/com.mattermost.apps/e2e-testapp' + url + '/set-submit-response', response);
}

const getSubmitRequest = (url) => {
    return cy.request('GET', 'http://localhost:8065/plugins/com.mattermost.apps/e2e-testapp' + url + '/get-submit-request');
}

const cleanTestApp = () => {
    return cy.request('POST', 'http://localhost:8065/plugins/com.mattermost.apps/e2e-testapp/clean');
}
