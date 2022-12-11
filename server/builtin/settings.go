// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) settingsCommandBinding(loc *i18n.Localizer) apps.Binding {
	// Open settings modal
	return apps.Binding{
		Location: "setttings",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.settings.label",
			Other: "settings",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.settings.description",
			Other: "Configure systemwide apps settings",
		}),
		Submit: newUserCall(pSettings),
	}
}

func (a *builtinApp) settings(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	conf := a.conf.Get()

	opts := []apps.SelectOption{
		{
			Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.on",
				Other: "On",
			}),
			Value: "true",
		},
		{
			Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.off",
				Other: "Off",
			}),
			Value: "false",
		},
	}

	fields := []apps.Field{
		{
			Name:       fDeveloperMode,
			Type:       apps.FieldTypeStaticSelect,
			IsRequired: true,
			Value:      getSelectedValue(opts[0], opts[1], conf.DeveloperMode),
			Description: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "modal.developer_mode.description",
					Other: "Enables various development tools.\n\nShould not be enabled in production.",
				},
			}),
			ModalLabel: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "modal.developer_mode.modal_label",
					Other: "Enable developer mode",
				},
			}),
			SelectStaticOptions: opts,
		}, {
			Name:       fAllowHTTPApps,
			Type:       apps.FieldTypeStaticSelect,
			IsRequired: true,
			Value:      getSelectedValue(opts[0], opts[1], conf.AllowHTTPApps),
			Description: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "modal.allow_http_apps.description",
					Other: "Allow apps, which run as an http server, to be installed.",
				},
			}),
			ModalLabel: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "modal.developer_mode.modal_label",
					Other: "Enable HTTP apps",
				},
			}),
			SelectStaticOptions: opts,
		},
	}

	form := apps.Form{
		Title: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "modal.settings.title",
				Other: "Configure systemwide apps settings",
			},
		}),
		Fields: fields,
		Submit: newUserCall(pSettingsSave),
	}
	return apps.NewFormResponse(form)
}

func (a *builtinApp) settingsSave(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	sc := a.conf.Get().StoredConfig

	developerModeOverride := apps.DeployType(creq.GetValue(fDeveloperMode, "")) == "true"
	sc.DeveloperModeOverride = &developerModeOverride

	allowHTTPAppsOverride := apps.DeployType(creq.GetValue(fAllowHTTPApps, "")) == "true"
	sc.AllowHTTPAppsOverride = &allowHTTPAppsOverride

	err := a.conf.StoreConfig(sc, r.Log)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to store configuration"))
	}

	resp := apps.NewTextResponse("Saved settings.")
	resp.RefreshBindings = true

	return resp
}

func getSelectedValue(trueOtion, falseOption apps.SelectOption, value bool) apps.SelectOption {
	if value {
		return trueOtion
	}

	return falseOption
}
