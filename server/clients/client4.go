package clients

import "github.com/mattermost/mattermost-server/v5/model"

type Client4 interface {
	SetToken(token string)
	GetUserByUsername(username, etag string) (*model.User, *model.Response)
	CreateBot(bot *model.Bot) (*model.Bot, *model.Response)
	EnableBot(userID string) (*model.Bot, *model.Response)
	GetUser(userID, etag string) (*model.User, *model.Response)
	UpdateUserRoles(userID, newRoles string) (bool, *model.Response)
	CreateUserAccessToken(userID, name string) (*model.UserAccessToken, *model.Response)
	GetOAuthApp(clientID string) (*model.OAuthApp, *model.Response)
	CreateOAuthApp(*model.OAuthApp) (*model.OAuthApp, *model.Response)
}

/*
	asAdmin := model.NewAPIv4Client(conf.MattermostSiteURL)
	asAdmin.SetToken(session.Token)
*/

type ClientService interface {
	NewClient4(siteURL string) Client4
}

type clientService struct{}

func NewClientService() ClientService {
	return clientService{}
}

func (c clientService) NewClient4(siteURL string) Client4 {
	return model.NewAPIv4Client(siteURL)
}
