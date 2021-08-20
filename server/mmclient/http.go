package mmclient

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type httpClient struct {
	mm *model.Client4
}

func NewHTTPClient(config config.Service, sessionID, actingUserID string) (Client, utils.LocError, error) {
	conf, mm, _ := config.Basic()
	client, locError, err := utils.ClientFromSession(mm, conf.MattermostSiteURL, sessionID, actingUserID)
	if err != nil {
		return nil, locError, err
	}

	return &httpClient{client}, nil, nil
}

// User section

func (h *httpClient) GetUserByUsername(userName string) (*model.User, error) {
	user, resp := h.mm.GetUserByUsername(userName, "")
	if resp.Error != nil {
		return nil, resp.Error
	}

	return user, nil
}

func (h *httpClient) CreateUserAccessToken(userID, description string) (*model.UserAccessToken, error) {
	token, resp := h.mm.CreateUserAccessToken(userID, description)
	if resp.Error != nil {
		return nil, resp.Error
	}

	return token, nil
}

func (h *httpClient) RevokeUserAccessToken(tokenID string) error {
	_, resp := h.mm.RevokeUserAccessToken(tokenID)
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

// OAuth section

func (h *httpClient) CreateOAuthApp(app *model.OAuthApp) error {
	createdOauthApp, resp := h.mm.CreateOAuthApp(app)
	if resp.Error != nil {
		return resp.Error
	}

	*app = *createdOauthApp

	return nil
}

func (h *httpClient) GetOAuthApp(appID string) (*model.OAuthApp, error) {
	oauthApp, resp := h.mm.GetOAuthApp(appID)
	if resp.Error != nil {
		return nil, resp.Error
	}

	return oauthApp, nil
}

func (h *httpClient) DeleteOAuthApp(appID string) error {
	_, resp := h.mm.DeleteOAuthApp(appID)
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

// Bot section

func (h *httpClient) CreateBot(bot *model.Bot) error {
	createdBot, resp := h.mm.CreateBot(bot)
	if resp.Error != nil {
		return resp.Error
	}

	*bot = *createdBot

	return nil
}

func (h *httpClient) GetBot(botUserID string) (*model.Bot, error) {
	bot, resp := h.mm.GetBot(botUserID, "")
	if resp.Error != nil {
		return nil, resp.Error
	}

	return bot, nil
}

func (h *httpClient) EnableBot(botUserID string) (*model.Bot, error) {
	bot, resp := h.mm.EnableBot(botUserID)
	if resp.Error != nil {
		return nil, resp.Error
	}

	return bot, nil
}

func (h *httpClient) DisableBot(botUserID string) (*model.Bot, error) {
	bot, resp := h.mm.DisableBot(botUserID)
	if resp.Error != nil {
		return nil, resp.Error
	}

	return bot, nil
}
