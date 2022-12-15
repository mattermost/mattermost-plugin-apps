// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

// logs copy [--channel|--create-channel] --level
func (a *builtinApp) debugLogsCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "logs",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.logs.label",
			Other: "logs",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.logs.description",
			Other: "Stream the logs of the apps plugin to the selected channel.",
		}),
		Form: &apps.Form{
			Submit: apps.NewCall(pDebugLogs).WithExpand(apps.Expand{
				Team:       apps.ExpandID.Required(),
				ActingUser: apps.ExpandSummary.Required(),
				Locale:     apps.ExpandAll,
			}),
			Fields: []apps.Field{
				{
					Name: fChannel,
					Type: apps.FieldTypeChannel,
					Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.channel.description",
						Other: "Select an existing channel, or specify --create-channel to create a new channel.",
					}),
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.channel.label",
						Other: "channel",
					}),
				},
				{
					Name: fCreate,
					Type: apps.FieldTypeBool,
					Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.create.channel.description",
						Other: "Create a new channel for the plugin logs. Use --channel to specify an existing channel.",
					}),
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.create.channel.label",
						Other: "create-channel",
					}),
				},
				{
					Name: fLevel,
					Type: apps.FieldTypeStaticSelect,
					Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.log.level.description",
						Other: "Set minimum log severity (level) to output.",
					}),
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.log.level.label",
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
						ID:    "field.json.description",
						Other: "Include entry properties in the output, as JSON.",
					}),
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.json.label",
						Other: "json",
					}),
				},
			},
		},
	}
}

func (a *builtinApp) debugLogs(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	levelStr := creq.GetValue(fLevel, zapcore.InfoLevel.String())
	var level zapcore.Level
	if err := level.Set(levelStr); err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "invalid log level"))
	}

	outputJSON := creq.BoolValue(fJSON)
	create := creq.BoolValue(fCreate)
	channel := creq.Values[fChannel]
	channelID, channelLabel := "", ""

	switch {
	case !create && channel == "":
		// maybe set the level later.

	case create && channel != nil:
		return apps.NewErrorResponse(errors.New("cannot specify both --channel and --create-channel"))

	case create && channel == nil:
		name := "apps-plugin-logs"

		ch, _ := a.conf.MattermostAPI().Channel.GetByName(creq.Context.Team.Id, name, false)
		if ch == nil {
			ch = &model.Channel{
				Name: name,
				DisplayName: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.logs.channel.displayname",
					Other: "DEBUG: Apps Plugin Logs",
				}),
				Type:   model.ChannelTypePrivate,
				TeamId: creq.Context.Team.Id,
			}

			if err := a.conf.MattermostAPI().Channel.Create(ch); err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to create channel"))
			}
		}

		_, err := a.conf.MattermostAPI().Channel.AddMember(ch.Id, creq.Context.ActingUser.Id)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to add user to channel"))
		}
		channelID = ch.Id
		channelLabel = ch.Name

	case !create && channel != nil:
		if optMap, ok := channel.(map[string]interface{}); ok {
			channelID = optMap["value"].(string)
			channelLabel = optMap["label"].(string)
		}
	}

	changed := false
	storedConfig := a.conf.Get().StoredConfig
	if int(level) != storedConfig.LogChannelLevel {
		storedConfig.LogChannelLevel = int(level)
		changed = true
	}
	if outputJSON != storedConfig.LogChannelJSON {
		storedConfig.LogChannelJSON = outputJSON
		changed = true
	}
	if channelID != "" && channelID != storedConfig.LogChannelID {
		storedConfig.LogChannelID = channelID
		changed = true
	}

	if changed {
		if err := a.conf.StoreConfig(storedConfig, r.Log); err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to update the plugin configuration"))
		}
	}

	if storedConfig.LogChannelID == "" {
		return apps.NewTextResponse(a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.logs.response.no.channel",
			Other: "No channel is configured for the plugin logs. Use `/apps debug logs --create-channel` to create a new channel.",
		}))
	}

	if channelLabel == "" {
		if ch, err := a.conf.MattermostAPI().Channel.Get(storedConfig.LogChannelID); err == nil {
			channelLabel = ch.Name
		}
	}

	message := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "command.debug.logs.response",
			Other: "Logs above {{.Level}} will be sent to ~{{.Channel}}. Use `/apps debug logs --channel` to change the channel.",
		},
		TemplateData: map[string]string{
			"Level":   levelStr,
			"Channel": channelLabel,
		},
	})
	if storedConfig.LogChannelJSON {
		message += "\n" + a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.logs.json.on.response",
			Other: "JSON output: on.",
		})
	} else {
		message += "\n" + a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.logs.json.off.response",
			Other: "JSON output: off.",
		})
	}

	return apps.NewTextResponse(message)
}
