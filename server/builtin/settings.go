// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-server/v6/model"
)

var settingsModalSourceCall = apps.NewCall(pSettingsModalSource).WithExpand(apps.Expand{
	ActingUser: apps.ExpandSummary.Required(),
	Locale:     apps.ExpandAll,
	Team:       apps.ExpandSummary.Required(),
})

var settingsModalSaveCall = apps.NewCall(pSettingsModalSave).WithExpand(apps.Expand{
	ActingUser: apps.ExpandSummary.Required(),
	Locale:     apps.ExpandAll,
	Team:       apps.ExpandSummary.Required(),
})

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
			Other: "Configure system-wide apps settings",
		}),
		Submit: settingsModalSourceCall,
	}
}

func (a *builtinApp) settingsForm(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	conf := a.conf.Get()

	haveOverrides := conf.DeveloperModeOverride != nil || conf.AllowHTTPAppsOverride != nil
	wantOverrides := creq.GetValue(fOverrides, "")
	useOverrides := wantOverrides == "use" || (wantOverrides == "" && haveOverrides)
	defaultDevMode, defaultAllowHTTP := a.conf.SystemDefaultFlags()
	optUseOverrides := apps.SelectOption{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.use_overrides",
			Other: "Use overrides for developer mode and HTTP apps",
		}),
		Value: "use",
	}
	optNoOverrides := apps.SelectOption{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.reset_overrides",
			Other: "Do not use overrides for developer mode and HTTP apps",
		}),
		Value: "none",
	}
	overrideSelectorField := apps.Field{
		Name:          fOverrides,
		Type:          apps.FieldTypeStaticSelect,
		IsRequired:    true,
		SelectRefresh: true,
		ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.overrides.modal_label",
			Other: "Overrides for developer mode and HTTP apps",
		}),
		SelectStaticOptions: []apps.SelectOption{
			optUseOverrides,
			optNoOverrides,
		},
		Value: optNoOverrides,
		Description: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "modal.overrides.description",
				Other: "Current system settings: developer mode: **{{.DeveloperMode}}**, allow HTTP apps: **{{.AllowHTTPApps}}**",
			},
			TemplateData: map[string]interface{}{
				"DeveloperMode": defaultDevMode,
				"AllowHTTPApps": defaultAllowHTTP,
			},
		}),
	}

	conf.DeveloperMode, conf.AllowHTTPApps = a.conf.SystemDefaultFlags()
	if useOverrides {
		if conf.DeveloperModeOverride == nil {
			conf.DeveloperModeOverride = &conf.DeveloperMode
		} else {
			conf.DeveloperMode = *conf.DeveloperModeOverride
		}
		if conf.AllowHTTPAppsOverride == nil {
			conf.AllowHTTPAppsOverride = &conf.AllowHTTPApps
		} else {
			conf.AllowHTTPApps = *conf.AllowHTTPAppsOverride
		}
		overrideSelectorField.Value = optUseOverrides
		overrideSelectorField.Description = ""
	}

	fields := []apps.Field{overrideSelectorField}
	if useOverrides {
		fields = append(fields,
			apps.Field{
				Name:  fDeveloperMode,
				Type:  apps.FieldTypeBool,
				Value: conf.DeveloperMode,
				Description: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "modal.developer_mode.description",
						Other: "Enables various development tools. Apps developer mode can lead to performance degradation of the Mattermost server and should not be used in a production environment.",
					},
				}),
				ModalLabel: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "modal.developer_mode.modal_label",
						Other: "Enable developer mode",
					},
				}),
			},
			apps.Field{
				Name:  fAllowHTTPApps,
				Type:  apps.FieldTypeBool,
				Value: conf.AllowHTTPApps,
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
			})
	}

	optUseChannelLog := apps.SelectOption{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.channel_log.use",
			Other: "Copy apps logs to a channel",
		}),
		Value: "use",
	}
	optNoChannelLog := apps.SelectOption{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.channel_log.none",
			Other: "Do not copy apps logs to a channel",
		}),
		Value: "none",
	}
	optCreateChannelLog := apps.SelectOption{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.channel_log.create",
			Other: "Create a new channel for apps logs",
		}),
		Value: "create",
	}
	optSelectChannelLog := apps.SelectOption{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.channel_log.select",
			Other: "Select an existing channel for apps logs",
		}),
		Value: "select",
	}

	var otherFields []apps.Field
	logSelectorField := apps.Field{
		Name:          fLog,
		Type:          apps.FieldTypeStaticSelect,
		SelectRefresh: true,
		IsRequired:    true,
		ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.channel_log.modal_label",
			Other: "Apps logs settings",
		}),
	}

	logDeveloperModeNeededField := apps.Field{
		Name:                fLog,
		IsRequired:          true,
		ReadOnly:            true,
		SelectStaticOptions: []apps.SelectOption{optNoChannelLog},
		Type:                apps.FieldTypeStaticSelect,
		Value:               optNoChannelLog,
		ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.channel_log.dev_mode_needed.modal_label",
			Other: "Apps logs settings",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.settings.channel_log.dev_mode_needed.description",
			Other: "Developer mode is required to use apps logs. Please enable developer mode in the settings above.",
		}),
	}

	haveLog := conf.LogChannelID != ""
	channelOpt := apps.SelectOption{}
	if haveLog {
		channelOpt = apps.SelectOption{
			Label: conf.LogChannelID + " (unavailable)",
			Value: conf.LogChannelID,
		}
		ch, _ := a.conf.MattermostAPI().Channel.Get(conf.LogChannelID)
		if ch != nil {
			channelOpt.Label = ch.DisplayName
		}
	}
	wantLog := creq.GetValue(fLog, "")

	logSettingsFields := []apps.Field{
		{
			Name:     fChannel,
			Type:     apps.FieldTypeChannel,
			Value:    channelOpt,
			ReadOnly: true,
			Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.description",
				Other: "Channel where apps logs are copied to",
			}),
			ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.label",
				Other: "channel",
			}),
		},
		{
			Name: fLevel,
			Type: apps.FieldTypeStaticSelect,
			Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.level.description",
				Other: "Set minimum log severity (level) to output.",
			}),
			ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.level.label",
				Other: "level",
			}),
			SelectStaticOptions: []apps.SelectOption{
				{
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "log.level.debug.label",
						Other: "Debug",
					}),
					Value: zapcore.DebugLevel.String(),
				},
				{
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "log.level.info.label",
						Other: "Info",
					}),
					Value: zapcore.InfoLevel.String(),
				},
				{
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "log.level.warn.label",
						Other: "Warning",
					}),
					Value: zapcore.WarnLevel.String(),
				},
				{
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "log.level.error.label",
						Other: "Error",
					}),
					Value: zapcore.ErrorLevel.String(),
				},
			},
			Value: zapcore.InfoLevel.String(),
		},
		{
			Name: fJSON,
			Type: apps.FieldTypeBool,
			Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.json.description",
				Other: "Include entry properties in the output, as JSON.",
			}),
			ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.json.label",
				Other: "json",
			}),
		},
	}

	createChannelFields := []apps.Field{
		{
			Name: fChannelName,
			Type: apps.FieldTypeText,
			Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.name.description",
				Other: "Name (short) a new channel where apps logs will be copied to (in the current team).",
			}),
			ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.name,label",
				Other: "Name",
			}),
		},
		{
			Name: fChannelDisplayName,
			Type: apps.FieldTypeText,
			Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.display_name.description",
				Other: "Display name for the logs channel",
			}),
			ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.display_name,label",
				Other: "Display Name",
			}),
		},
	}

	selectChannelFields := []apps.Field{
		{
			Name: fChannel,
			Type: apps.FieldTypeChannel,
			Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.channel.description",
				Other: "Select an existing channel where apps logs will be copied to.",
			}),
			ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.channel_log.channel.channel,label",
				Other: "Channel",
			}),
		},
	}

	switch {
	case !conf.DeveloperMode:
		logSelectorField = logDeveloperModeNeededField

	case haveLog:
		logSelectorField.SelectStaticOptions = []apps.SelectOption{
			optNoChannelLog,
			optUseChannelLog,
		}
		switch wantLog {
		case "", "use":
			logSelectorField.Value = optUseChannelLog
			otherFields = logSettingsFields
		case "none":
			// no other fields to display
			logSelectorField.Value = optNoChannelLog
		default:
			return apps.NewErrorResponse(utils.NewInvalidError("invalid input %s: already have log settings, reset to none first", wantLog))
		}

	case !haveLog:
		logSelectorField.SelectStaticOptions = []apps.SelectOption{
			optNoChannelLog,
			optCreateChannelLog,
			optSelectChannelLog,
		}
		switch wantLog {
		case "", "none":
			// no other fields to display
			logSelectorField.Value = optNoChannelLog
		case "create":
			logSelectorField.Value = optCreateChannelLog
			otherFields = createChannelFields
		case "select":
			logSelectorField.Value = optSelectChannelLog
			otherFields = selectChannelFields
		default:
			return apps.NewErrorResponse(utils.NewInvalidError("invalid input %s: create or select a channel first", wantLog))
		}
	}
	fields = append(fields, logSelectorField)
	fields = append(fields, otherFields...)

	form := apps.Form{
		Title: a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "modal.settings.title",
				Other: "Configure systemwide apps settings",
			},
		}),
		Fields: fields,
		Submit: settingsModalSaveCall,
		Source: settingsModalSourceCall,
	}
	return apps.NewFormResponse(form)
}

func (a *builtinApp) settingsSave(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	r.Log.Debugf("<>/<> Save 0: values: %v", creq.Values)

	wantOverrides := creq.GetValue(fOverrides, "")
	developerModeOverride := creq.BoolValue(fDeveloperMode)
	allowHTTPAppsOverride := creq.BoolValue(fAllowHTTPApps)
	wantLog := creq.GetValue(fLog, "")
	channelID := creq.GetValue(fChannel, "")
	channelName := creq.GetValue(fChannelName, "")
	channelDisplayName := creq.GetValue(fChannelDisplayName, "")
	levelStr := creq.GetValue(fLevel, zapcore.InfoLevel.String())
	var level zapcore.Level
	if err := level.Set(levelStr); err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "invalid log level"))
	}
	outputJSON := creq.BoolValue(fJSON)

	sc := a.conf.Get().StoredConfig
	haveOverrides := sc.DeveloperModeOverride != nil || sc.AllowHTTPAppsOverride != nil
	haveLog := sc.LogChannelID != ""

	r.Log.Debugf("<>/<> Save 0: fChannel %T %v", creq.Values[fChannel], creq.Values[fChannel])

	r.Log.Debugf("<>/<> Save 1: haveOverrides=%v, wantOverrides=%s, haveLog=%v, wantLog=%s", haveOverrides, wantOverrides, haveLog, wantLog)

	if !haveOverrides && wantOverrides == "none" && !haveLog && wantLog == "none" {
		loc := a.newLocalizer(creq)
		return apps.CallResponse{
			Type: apps.CallResponseTypeOK,
			Text: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "modal.settings.save.nothing_to_do",
				Other: "No changes to save",
			}),
			Data: sc,
		}
	}

	switch {
	case wantOverrides == "use":
		sc.DeveloperModeOverride = &developerModeOverride
		sc.AllowHTTPAppsOverride = &allowHTTPAppsOverride
		r.Log.Debugf("<>/<> Save 2: setting overrides to %v, %v", developerModeOverride, allowHTTPAppsOverride)

	case wantOverrides == "none":
		sc.DeveloperModeOverride = nil
		sc.AllowHTTPAppsOverride = nil
		r.Log.Debugf("<>/<> Save 3: resetting overrides")

	default:
		return apps.NewErrorResponse(utils.NewInvalidError("invalid input %s:%s: must be 'use' or 'none'", fOverrides, wantOverrides))
	}

	redirect := false
	switch {
	case haveLog && wantLog == "use":
		sc.LogChannelLevel = int(level)
		sc.LogChannelJSON = outputJSON
		r.Log.Debugf("<>/<> Save 4: setting log settings to %s, %v", level, outputJSON)

	case wantLog == "none":
		sc.LogChannelID = ""
		sc.LogChannelLevel = 0
		sc.LogChannelJSON = false
		r.Log.Debugf("<>/<> Save 5: resetting log settings")

	case !haveLog && wantLog == "create":
		ch, _ := a.conf.MattermostAPI().Channel.GetByName(creq.Context.Team.Id, channelName, false)
		if ch == nil {
			ch = &model.Channel{
				Name:        channelName,
				DisplayName: channelDisplayName,
				Type:        model.ChannelTypePrivate,
				TeamId:      creq.Context.Team.Id,
			}
			if err := a.conf.MattermostAPI().Channel.Create(ch); err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to create channel"))
			}
			r.Log.Debugf("<>/<> Save 6: created channel %s %s %s", ch.Id, ch.Name, ch.DisplayName)
		}
		_, err := a.conf.MattermostAPI().Channel.AddMember(ch.Id, creq.Context.ActingUser.Id)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to add user to channel"))
		}
		sc.LogChannelID = ch.Id
		r.Log.Debugf("<>/<> Save 7: forwarding to settings modal to set up log settings")
		// Forward to the settings modal to set up the options
		redirect = true

	case !haveLog && wantLog == "select":
		sc.LogChannelID = channelID
		r.Log.Debugf("<>/<> Save 8: using channel %s", channelID)
		// Forward to the settings modal to set up the options
		r.Log.Debugf("<>/<> Save 9: forwarding to settings modal to set up log settings")
		redirect = true

	default:
		return apps.NewErrorResponse(utils.NewInvalidError("invalid input %s:%s: must be 'use', 'select', 'create', or 'none'", fLog, wantLog))
	}

	err := a.conf.StoreConfig(sc, r.Log)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to store configuration"))
	}
	r.Log.Debugf("<>/<> Save 100: stored")

	if !redirect {
		resp := apps.NewTextResponse("Saved settings.")
		resp.RefreshBindings = true
		return resp
	}

	creq.Values[fLog] = "use"
	return a.settingsForm(r, creq)
}

func getSelectedValue(trueOtion, falseOption apps.SelectOption, value bool) apps.SelectOption {
	if value {
		return trueOtion
	}

	return falseOption
}
