// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

func (a *builtinApp) installCommandBinding(loc *i18n.Localizer) apps.Binding {
	if a.conf.Get().MattermostCloudMode {
		return apps.Binding{
			Location: "install",
			Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.install.cloud.label",
				Other: "install",
			}),
			Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.install.cloud.hint",
				Other: "[ app ID ]",
			}),
			Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.install.cloud.description",
				Other: "Install an App from the Marketplace",
			}),
			Form: a.appIDForm(newAdminCall(pInstallListed).WithLocale(), newAdminCall(pInstallListedLookup), loc),
		}
	}
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.install.label",
			Other: "install",
		}),
		Location: "install",
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.install.hint",
			Other: "[ listed | url ]",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.install.description",
			Other: "Installs an App, locally deployed or from a remote URL",
		}),
		Bindings: []apps.Binding{
			{
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.install.listed.label",
					Other: "listed",
				}),
				Location: "listed",
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.install.listed.hint",
					Other: "[app ID]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.install.listed.description",
					Other: "Installs a listed App that has been locally deployed. (in the future, applicable Marketplace Apps will also be listed here).",
				}),
				Form: a.appIDForm(newAdminCall(pInstallListed).WithLocale(), newAdminCall(pInstallListedLookup), loc),
			},
			{
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.install.http.label",
					Other: "http",
				}),
				Location: "http",
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.install.http.hint",
					Other: "[URL to manifest.json]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.install.http.description",
					Other: "Installs an HTTP App from a URL",
				}),
				Form: &apps.Form{
					Fields: []apps.Field{
						{
							Name: fURL,
							Type: apps.FieldTypeText,
							Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "command.enable.install.http.form.description",
								Other: "enter the HTTP URL for the app's manifest.json",
							}),
							Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "command.enable.install.http.form.label",
								Other: "url",
							}),
							AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "command.enable.install.http.form.autocompleteHint",
								Other: "URL",
							}),
							AutocompletePosition: 1,
							IsRequired:           true,
						},
					},
					Submit: newAdminCall(pInstallHTTP).WithLocale(),
				},
			},
		},
	}
}

func (a *builtinApp) installListedLookup(creq apps.CallRequest) apps.CallResponse {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return !app.Installed
	})
}

func (a *builtinApp) installListed(creq apps.CallRequest) apps.CallResponse {
	loc := i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	m, err := a.proxy.GetManifest(appID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewFormResponse(a.newInstallConsentForm(*m, creq, "", loc))
}

func (a *builtinApp) installHTTP(creq apps.CallRequest) apps.CallResponse {
	loc := i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
	manifestURL := creq.GetValue(fURL, "")
	conf := a.conf.Get()
	data, err := a.httpOut.GetFromURL(manifestURL, conf.DeveloperMode, apps.MaxManifestSize)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	m, err := apps.DecodeCompatibleManifest(data)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	m, err = a.proxy.UpdateAppListing(appclient.UpdateAppListingRequest{
		Manifest:   *m,
		AddDeploys: apps.DeployTypes{apps.DeployHTTP},
	})
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewFormResponse(a.newInstallConsentForm(*m, creq, apps.DeployHTTP, loc))
}
