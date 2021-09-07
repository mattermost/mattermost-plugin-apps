package mmclient

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type httpClient struct {
	mm *model.Client4
}

func NewHTTPClientFromSessionID(config config.Service, sessionID, actingUserID string) (Client, error) {
	conf, mm, _ := config.Basic()
	client, err := utils.ClientFromSession(mm, conf.MattermostLocalURL, sessionID, actingUserID)
	if err != nil {
		return nil, err
	}

	return &httpClient{client}, nil
}

func NewHTTPClientFromToken(config config.Service, token, actingUserID string) (Client, error) {
	conf := config.Get()

	client := model.NewAPIv4Client(conf.MattermostLocalURL)
	client.SetToken(token)

	return &httpClient{client}, nil
}

// User section

func (h *httpClient) GetUser(userID string) (*model.User, error) {
	user, _, err := h.mm.GetUser(userID, "")
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *httpClient) GetUserByUsername(userName string) (*model.User, error) {
	user, _, err := h.mm.GetUserByUsername(userName, "")
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *httpClient) CreateUserAccessToken(userID, description string) (*model.UserAccessToken, error) {
	token, _, err := h.mm.CreateUserAccessToken(userID, description)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (h *httpClient) RevokeUserAccessToken(tokenID string) error {
	_, err := h.mm.RevokeUserAccessToken(tokenID)
	if err != nil {
		return err
	}

	return nil
}

// Channel section

func (h *httpClient) GetChannel(channelID string) (*model.Channel, error) {
	channel, _, err := h.mm.GetChannel(channelID, "")
	if err != nil {
		return nil, err
	}

	return channel, nil
}

// Team section

func (h *httpClient) GetTeam(teamID string) (*model.Team, error) {
	team, _, err := h.mm.GetTeam(teamID, "")
	if err != nil {
		return nil, err
	}

	return team, nil
}

// Post section

func (h *httpClient) GetPost(postID string) (*model.Post, error) {
	post, _, err := h.mm.GetPost(postID, "")
	if err != nil {
		return nil, err
	}

	return post, nil
}

// OAuth section

func (h *httpClient) CreateOAuthApp(app *model.OAuthApp) error {
	createdOauthApp, _, err := h.mm.CreateOAuthApp(app)
	if err != nil {
		return err
	}

	*app = *createdOauthApp

	return nil
}

func (h *httpClient) GetOAuthApp(appID string) (*model.OAuthApp, error) {
	oauthApp, _, err := h.mm.GetOAuthApp(appID)
	if err != nil {
		return nil, err
	}

	return oauthApp, nil
}

func (h *httpClient) DeleteOAuthApp(appID string) error {
	_, err := h.mm.DeleteOAuthApp(appID)
	if err != nil {
		return err
	}

	return nil
}

// Bot section

func (h *httpClient) CreateBot(bot *model.Bot) error {
	createdBot, _, err := h.mm.CreateBot(bot)
	if err != nil {
		return err
	}

	*bot = *createdBot

	return nil
}

func (h *httpClient) GetBot(botUserID string) (*model.Bot, error) {
	bot, _, err := h.mm.GetBot(botUserID, "")
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func (h *httpClient) EnableBot(botUserID string) (*model.Bot, error) {
	bot, _, err := h.mm.EnableBot(botUserID)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func (h *httpClient) DisableBot(botUserID string) (*model.Bot, error) {
	bot, _, err := h.mm.DisableBot(botUserID)
	if err != nil {
		return nil, err
	}

	return bot, nil
}
