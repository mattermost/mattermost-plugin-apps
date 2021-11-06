// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

var installHTTPCall = apps.Call{
	Path: pInstallHTTP,
	Expand: &apps.Expand{
		ActingUser:            apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
	},
}

var installListedCall = apps.Call{
	Path: pInstallListed,
	Expand: &apps.Expand{
		ActingUser:            apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
	},
}

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
			Call: &installListedCall,
			Form: a.appIDForm(installListedCall, loc),
		}
	}

	return apps.Binding{
		Location: "install",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.install.label",
			Other: "install",
		}),
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.install.hint",
			Other: "[ listed | url ]",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.install.description",
			Other: "Install an App, locally deployed or from a remote URL",
		}),
		Bindings: []apps.Binding{
			{
				Location: "listed",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.listed.label",
					Other: "listed",
				}),
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.listed.hint",
					Other: "[ app ID ]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.listed.description",
					Other: "Install a listed App that has been locally deployed. (in the future, applicable Marketplace Apps will also be listed here).",
				}),
				Call: &installListedCall,
				Form: a.appIDForm(installListedCall, loc),
			},
			{
				Location: "http",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.http.label",
					Other: "http",
				}),
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.http.hint",
					Other: "[ manifest.json URL ]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.install.http.description",
					Other: "Install an App from an HTTP URL",
				}),
				Call: &installHTTPCall,
				Form: &apps.Form{
					Fields: []apps.Field{
						{
							Name: fURL,
							Type: apps.FieldTypeText,
							Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.url.label",
								Other: "url",
							}),
							Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.url.description",
								Other: "enter the URL for the app's manifest.json",
							}),
							AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.url.hint",
								Other: "HTTP(s) URL",
							}),
							AutocompletePosition: 1,
							IsRequired:           true,
						},
					},
					Call: &installHTTPCall,
				},
			},
		},
	}
}

func (a *builtinApp) installListed() handler {
	return handler{
		requireSysadmin: true,

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			res, err := a.lookupAppID(creq, nil)
			return res, err
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			loc := a.newLocalizer(creq)
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			m, err := a.proxy.GetManifest(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewFormResponse(*a.newInstallConsentForm(*m, creq, "", loc))
		},
	}
}

func (a *builtinApp) installHTTP() handler {
	return handler{
		requireSysadmin: true,

		submitf: func(creq apps.CallRequest) apps.CallResponse {
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
			m, err = a.proxy.UpdateAppListing(appclient.UpdateAppListingRequest{
				Manifest:   *m,
				AddDeploys: apps.DeployTypes{apps.DeployHTTP},
			})
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewFormResponse(*a.newInstallConsentForm(*m, creq, apps.DeployHTTP, loc))
		},
	}
}
