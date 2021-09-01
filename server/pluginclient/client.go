package pluginclient

import (
	"io"

	"github.com/mattermost/mattermost-server/v5/model"
)

type Client interface {
	GetUser(userID string) (*model.User, error)
	GetUserByUsername(userName string) (*model.User, error)
	CreateUserAccessToken(userID, description string) (*model.UserAccessToken, error)
	RevokeUserAccessToken(tokenID string) error
	SetProfileImage(userID string, content io.Reader) error

	GetChannel(channelID string) (*model.Channel, error)

	GetTeam(teamID string) (*model.Team, error)

	GetPost(postID string) (*model.Post, error)

	CreateOAuthApp(app *model.OAuthApp) error
	GetOAuthApp(appID string) (*model.OAuthApp, error)
	DeleteOAuthApp(appID string) error

	GetBot(botUserID string) (*model.Bot, error)
	CreateBot(bot *model.Bot) error
	EnableBot(botUserID string) (*model.Bot, error)
	DisableBot(botUserID string) (*model.Bot, error)
}
