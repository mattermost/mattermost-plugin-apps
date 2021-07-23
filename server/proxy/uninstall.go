// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

func (p *Proxy) UninstallApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (md.MD, error) {
	loc := p.i18n.GetUserLocalizer(cc.ActingUserID)
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", errors.Wrap(err, p.i18n.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.error.getApp",
				Other: "failed to get app. appID: {{.AppID}}",
			},
			TemplateData: map[string]string{
				"AppID": string(appID),
			},
		}))
	}

	var message md.MD
	if app.OnUninstall != nil {
		creq := &apps.CallRequest{
			Call:    *app.OnUninstall,
			Context: cc,
		}
		resp := p.Call(sessionID, cc.ActingUserID, creq)
		if resp.Type == apps.CallResponseTypeError {
			p.mm.Log.Warn("OnUninstall failed, uninstalling app anyway", "err", resp.Error(), "app_id", app.AppID)
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = md.MD(p.i18n.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.default",
				Other: "Uninstalled {{.AppID}}",
			},
			TemplateData: map[string]string{
				"AppID": app.DisplayName,
			},
		}))
	}

	// delete oauth app
	if app.MattermostOAuth2.ClientID != "" {
		if err = client.DeleteOAuthApp(app.MattermostOAuth2.ClientID); err != nil {
			return "", errors.Wrap(err, p.i18n.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
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
			return "", errors.Wrap(err, p.i18n.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
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
		return "", errors.Wrap(err, p.i18n.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
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
		return "", errors.Wrap(err, p.i18n.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.error.delete",
				Other: "can't delete app - {{.AppID}}",
			},
			TemplateData: map[string]string{
				"AppID": string(app.AppID),
			},
		}))
	}

	// in on-prem mode the manifest need to be deleted as every install add a manifest anyway
	conf := p.conf.GetConfig()
	if !conf.MattermostCloudMode {
		if err = p.store.Manifest.DeleteLocal(app.AppID); err != nil {
			return "", errors.Wrap(err, p.i18n.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "apps.uninstall.error.deleteManifest",
					Other: "can't delete manifest for uninstalled app - {{.AppID}}",
				},
				TemplateData: map[string]string{
					"AppID": string(app.AppID),
				},
			}))
		}
	}

	// remove data
	if err = p.store.AppKV.DeleteAll(app.BotUserID); err != nil {
		return "", errors.Wrap(err, p.i18n.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.uninstall.error.deleteData",
				Other: "can't delete app data - {{.AppID}}",
			},
			TemplateData: map[string]string{
				"AppID": string(app.AppID),
			},
		}))
	}

	p.mm.Log.Info("Uninstalled app", "app_id", app.AppID)

	p.dispatchRefreshBindingsEvent(cc.ActingUserID)

	return message, nil
}
