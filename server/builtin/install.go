// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var installHTTPCall = apps.Call{
	Path: pInstallHTTP,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

var installListedCall = apps.Call{
	Path: pInstallListed,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) installCommandBinding(loc *i18n.Localizer) apps.Binding {
	if a.conf.Get().MattermostCloudMode {
		return apps.Binding{
			Location:    "install",
			Label:       a.conf.Local(loc, "command.install.cloud.label"),
			Description: a.conf.Local(loc, "command.install.cloud.description"),
			Hint:        a.conf.Local(loc, "command.install.cloud.hint"),
			Call:        &installListedCall,
			Form:        a.appIDForm(installListedCall, loc),
		}
	} else {
		return apps.Binding{
			Location:    "install",
			Label:       a.conf.Local(loc, "command.install.label"),
			Description: a.conf.Local(loc, "command.install.description"),
			Hint:        a.conf.Local(loc, "command.install.hint"),
			Bindings: []apps.Binding{
				{
					Location:    "listed",
					Label:       a.conf.Local(loc, "command.install.listed.label"),
					Description: a.conf.Local(loc, "command.install.listed.description"),
					Hint:        a.conf.Local(loc, "command.install.listed.hint"),
					Call:        &installListedCall,
					Form:        a.appIDForm(installListedCall, loc),
				},
				{
					Location:    "http",
					Label:       a.conf.Local(loc, "command.install.http.label"),
					Hint:        a.conf.Local(loc, "command.install.http.hint"),
					Description: a.conf.Local(loc, "command.install.http.description"),
					Call:        &installHTTPCall,
					Form: &apps.Form{
						Fields: []apps.Field{
							{
								Name:                 fURL,
								Type:                 apps.FieldTypeText,
								Label:                a.conf.Local(loc, "field.url.label"),
								Description:          a.conf.Local(loc, "field.url.description"),
								AutocompleteHint:     a.conf.Local(loc, "field.url.hint"),
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
