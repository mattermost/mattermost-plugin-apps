package mmclient

import (
	"io"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type httpClient struct {
	mm *model.Client4
}

func NewHTTPClient(conf config.Config, token string) Client {
	client := model.NewAPIv4Client(conf.MattermostLocalURL)
	client.SetToken(token)
	return &httpClient{client}
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

const (
	// MaxProfileImageSize is the maximum length in bytes of the profile image file.
	MaxProfileImageSize = 50 * 1024 * 1024 // 50Mb
)

func (h *httpClient) SetProfileImage(userID string, content io.Reader) error {
	data, err := httputils.LimitReadAll(content, MaxProfileImageSize)
	if err != nil {
		return err
	}
	_, err = h.mm.SetProfileImage(userID, data)
	if err != nil {
		return err
	}
	return nil
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

func (h *httpClient) GetChannelMember(channelID, userID string) (*model.ChannelMember, error) {
	channelMember, _, err := h.mm.GetChannelMember(channelID, userID, "")
	if err != nil {
		return nil, err
	}

	return channelMember, nil
}

// Team section

func (h *httpClient) GetTeam(teamID string) (*model.Team, error) {
	team, _, err := h.mm.GetTeam(teamID, "")
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (h *httpClient) GetTeamMember(teamID, userID string) (*model.TeamMember, error) {
	teamMember, _, err := h.mm.GetTeamMember(teamID, userID, "")
	if err != nil {
		return nil, err
	}

	return teamMember, nil
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
