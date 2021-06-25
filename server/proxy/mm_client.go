package proxy

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type MMClient interface {
	GetUserByUsername(userName string) (*model.User, error)
	CreateUserAccessToken(userID, description string) (*model.UserAccessToken, error)
	RevokeUserAccessToken(tokenID string) error

	CreateOAuthApp(app *model.OAuthApp) error
	GetOAuthApp(appID string) (*model.OAuthApp, error)
	DeleteOAuthApp(appID string) error

	GetBot(botUserID string) (*model.Bot, error)
	CreateBot(bot *model.Bot) error
	EnableBot(botUserID string) (*model.Bot, error)
	DisableBot(botUserID string) (*model.Bot, error)
}

type httpClient struct {
	mm *model.Client4
}

func (p *Proxy) GetMMHTTPClient(sessionID, actingUserID string) (MMClient, error) {
	conf := p.conf.GetConfig()

	client, err := utils.ClientFromSession(p.mm, conf.MattermostSiteURL, sessionID, actingUserID)
	if err != nil {
		return nil, err
	}

	return &httpClient{client}, nil
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

type rpcClient struct {
	mm *pluginapi.Client
}

func (p *Proxy) GetMMPRCClient() MMClient {
	return &rpcClient{p.mm}
}

// User section

func (r *rpcClient) GetUserByUsername(userName string) (*model.User, error) {
	return r.mm.User.GetByUsername(userName)
}

func (r *rpcClient) CreateUserAccessToken(userID, description string) (*model.UserAccessToken, error) {
	return r.mm.User.CreateAccessToken(userID, description)
}

func (r *rpcClient) RevokeUserAccessToken(tokenID string) error {
	return r.mm.User.RevokeAccessToken(tokenID)
}

// OAuth section

func (r *rpcClient) CreateOAuthApp(app *model.OAuthApp) error {
	return r.mm.OAuth.Create(app)
}

func (r *rpcClient) GetOAuthApp(appID string) (*model.OAuthApp, error) {
	return r.mm.OAuth.Get(appID)
}

func (r *rpcClient) DeleteOAuthApp(appID string) error {
	return r.mm.OAuth.Delete(appID)
}

// Bpt section

func (r *rpcClient) CreateBot(bot *model.Bot) error {
	return r.mm.Bot.Create(bot)
}

func (r *rpcClient) GetBot(botUserID string) (*model.Bot, error) {
	return r.mm.Bot.Get(botUserID, false)
}

func (r *rpcClient) EnableBot(botUserID string) (*model.Bot, error) {
	return r.mm.Bot.UpdateActive(botUserID, true)
}

func (r *rpcClient) DisableBot(botUserID string) (*model.Bot, error) {
	return r.mm.Bot.UpdateActive(botUserID, false)
}
