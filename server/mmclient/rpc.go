package mmclient

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
)

type rpcClient struct {
	mm *pluginapi.Client
}

func NewRPCClient(c *pluginapi.Client) Client {
	return &rpcClient{c}
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

// Bot section

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
