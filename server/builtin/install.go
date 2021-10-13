// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var installURLCall = apps.Call{
	Path: pInstallURL,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll, // ensure sysadmin
	},
}

var installListedCall = apps.Call{
	Path: pInstallListed,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll, // ensure sysadmin
	},
}

func (a *builtinApp) installCommandBinding(loc *i18n.Localizer) apps.Binding {
	if a.conf.Get().MattermostCloudMode {
		return apps.Binding{
			Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.enable.install_cloud.label",
				Other: "install",
			}),
			Location: "install",
			Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.enable.install_cloud.hint",
				Other: "[app ID]",
			}),
			Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.enable.install_cloud.description",
				Other: "Installs an App from the Marketplace",
			}),
			Call: &installListedCall,
			Form: a.appIDForm(installListedCall, loc),
		}
	} else {
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
					Call: &installListedCall,
					Form: a.appIDForm(installListedCall, loc),
				},
				{
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "command.enable.install.url.label",
						Other: "url",
					}),
					Location: "url",
					Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "command.enable.install.url.hint",
						Other: "[manifest.json URL]",
					}),
					Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "command.enable.install.url.description",
						Other: "Installs an App from an HTTP URL",
					}),
					Call: &installURLCall,
					Form: &apps.Form{
						Fields: []apps.Field{
							{
								Name: fURL,
								Type: apps.FieldTypeText,
								Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
									ID:    "command.enable.install.url.form.description",
									Other: "enter the URL for the app's manifest.json",
								}),
								Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
									ID:    "command.enable.install.url.form.label",
									Other: "url",
								}),
								AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
									ID:    "command.enable.install.url.form.autocompleteHint",
									Other: "URL",
								}),
								AutocompletePosition: 1,
								IsRequired:           true,
							},
						},
						Call: &installURLCall,
					},
				},
			},
		}
	}
}

func (a *builtinApp) installListed() handler {
	return handler{
		requireSysadmin: true,

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			res, err := a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return !app.Installed
			})
			return res, err
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			m, err := a.proxy.GetManifest(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewFormResponse(*a.newInstallConsentForm(*m, creq))
		},
	}
}

func (a *builtinApp) installURL() handler {
	return handler{
		requireSysadmin: true,

		submitf: func(creq apps.CallRequest) apps.CallResponse {
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
			_, err = a.proxy.StoreLocalManifest(*m)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewFormResponse(*a.newInstallConsentForm(*m, creq))
		},
	}
}
