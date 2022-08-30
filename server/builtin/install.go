// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) installCommandBinding(loc *i18n.Localizer) apps.Binding {
	conf := a.conf.Get()
	if !conf.AllowHTTPApps {
		// No need for subcommands, only listed apps are installable.
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
			Form: &apps.Form{
				Submit: newUserCall(pInstallListed),
				Fields: []apps.Field{
					a.appIDField(LookupNotInstalledApps, 1, true, loc),
				},
			},
		}
	}

	// install http|listed
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.install.label",
			Other: "install",
		}),
		Location: "install",
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.install.hint",
			Other: "[ listed | http ]",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.install.description",
			Other: "Install an App, locally deployed or from a remote URL",
		}),
		Bindings: []apps.Binding{
			{
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.listed.label",
					Other: "listed",
				}),
				Location: "listed",
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.listed.hint",
					Other: "[app ID]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.listed.description",
					Other: "Install a listed App that has been locally deployed. (in the future, applicable Marketplace Apps will also be listed here).",
				}),
				Form: &apps.Form{
					Submit: newUserCall(pInstallListed),
					Fields: []apps.Field{
						a.appIDField(LookupNotInstalledApps, 1, true, loc),
					},
				},
			},
			{
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.http.label",
					Other: "http",
				}),
				Location: "http",
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.http.hint",
					Other: "[URL to manifest.json]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.http.description",
					Other: "Install an HTTP App from a URL",
				}),
				Form: &apps.Form{
					Fields: []apps.Field{
						{
							Name: fURL,
							Type: apps.FieldTypeText,
							Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.url.description",
								Other: "enter the HTTP URL for the app's manifest.json",
							}),
							Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.url.label",
								Other: "url",
							}),
							AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.url.hint",
								Other: "URL",
							}),
							AutocompletePosition: 1,
							IsRequired:           true,
						},
					},
					Submit: newUserCall(pInstallHTTP).WithLocale(),
				},
			},
		},
	}
}

func (a *builtinApp) installListed(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	m, err := a.proxy.GetManifest(appID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewFormResponse(a.newInstallConsentForm(*m, creq, "", loc))
}

func (a *builtinApp) installHTTP(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
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
	m, err = a.proxy.UpdateAppListing(r, appclient.UpdateAppListingRequest{
		Manifest:   *m,
		AddDeploys: apps.DeployTypes{apps.DeployHTTP},
	})
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewFormResponse(a.newInstallConsentForm(*m, creq, apps.DeployHTTP, loc))
}
