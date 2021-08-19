// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
)

func (p *Proxy) UninstallApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (string, error) {
	loc := p.conf.I18N().GetUserLocalizer(cc.ActingUserID)
	conf, _, log := p.conf.Basic()
	log = log.With("app_id", appID)
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", errors.Wrap(err, p.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.error.getApp",
				Other: "failed to get app. appID: {{.AppID}}",
			},
			TemplateData: map[string]string{
				"AppID": string(appID),
			},
		}))
	}

	var message string
	if app.OnUninstall != nil {
		creq := &apps.CallRequest{
			Call:    *app.OnUninstall,
			Context: cc,
		}
		resp := p.Call(sessionID, cc.ActingUserID, creq)
		if resp.Type == apps.CallResponseTypeError {
			log.WithError(err).Warnf("OnUninstall failed, uninstalling app anyway")
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = p.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.default",
				Other: "Uninstalled {{.DisplayName}}",
			},
			TemplateData: map[string]string{
				"DisplayName": app.DisplayName,
			},
		})
	}

	// delete oauth app
	if app.MattermostOAuth2.ClientID != "" {
		if err = client.DeleteOAuthApp(app.MattermostOAuth2.ClientID); err != nil {
			return "", errors.Wrap(err, p.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "apps.uninstall.error.deleteOAuth",
					Other: "failed to delete Mattermost OAuth2 for {{.AppID}}",
				},
				TemplateData: map[string]string{
					"AppID": string(app.AppID),
				},
			}))
		}
	}

	// revoke bot account token if there is one
	if app.BotAccessTokenID != "" {
		if err = client.RevokeUserAccessToken(app.BotAccessTokenID); err != nil {
			return "", errors.Wrap(err, p.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "apps.uninstall.error.revokeBot",
					Other: "failed to revoke bot access token for {{.AppID}}",
				},
				TemplateData: map[string]string{
					"AppID": string(app.AppID),
				},
			}))
		}
	}

	// disable the bot account
	if _, err = client.DisableBot(app.BotUserID); err != nil {
		return "", errors.Wrap(err, p.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.error.disableBot",
				Other: "failed to disable bot account for {{.AppID}}",
			},
			TemplateData: map[string]string{
				"AppID": string(app.AppID),
			},
		}))
	}

	// delete app
	if err = p.store.App.Delete(app.AppID); err != nil {
		return "", errors.Wrap(err, p.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.error.delete",
				Other: "can't delete app {{.AppID}}",
			},
			TemplateData: map[string]string{
				"AppID": string(app.AppID),
			},
		}))
	}

	// in on-prem mode the manifest need to be deleted as every install add a manifest anyway
	if !conf.MattermostCloudMode {
		if err = p.store.Manifest.DeleteLocal(app.AppID); err != nil {
			return "", errors.Wrap(err, p.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "apps.uninstall.error.deleteManifest",
					Other: "can't delete manifest for uninstalled app {{.AppID}}",
				},
				TemplateData: map[string]string{
					"AppID": string(app.AppID),
				},
			}))
		}
	}

	// remove data
	if err = p.store.AppKV.DeleteAll(app.BotUserID); err != nil {
		return "", errors.Wrap(err, p.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.error.deleteData",
				Other: "can't delete app data for {{.AppID}}",
			},
			TemplateData: map[string]string{
				"AppID": string(app.AppID),
			},
		}))
	}

	log.Infof("Uninstalled app.")

	p.dispatchRefreshBindingsEvent(cc.ActingUserID)

	return message, nil
}
