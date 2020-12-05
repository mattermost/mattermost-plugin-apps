package examples

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

type Client struct {
	*model.Client4
	userID string
}

func as(id, token string, cc *api.Context) Client {
	if cc.Config == nil {
		return Client{}
	}
	return newClient(id, token, cc.Config.SiteURL)
}

func AsBot(cc *api.Context) Client {
	return as(cc.BotUserID, cc.BotAccessToken, cc)
}

func AsActingUser(cc *api.Context) Client {
	return as(cc.ActingUserID, cc.ActingUserAccessToken, cc)
}

func AsAdmin(cc *api.Context) Client {
	return as(cc.ActingUserID, cc.AdminAccessToken, cc)
}

func newClient(userID, token, mattermostSiteURL string) Client {
	i := Client{
		userID:  userID,
		Client4: model.NewAPIv4Client(mattermostSiteURL),
	}
	i.Client4.SetOAuthToken(token)
	return i
}

func (i *Client) CreatePost(post *model.Post) (*model.Post, error) {
	var createdPost *model.Post
	var res *model.Response
	post.UserId = i.userID

	createdPost, res = i.Client4.CreatePost(post)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return createdPost, nil
}

func (i *Client) DM(userID string, format string, args ...interface{}) {
	channel, err := i.getDirectChannelWith(userID)
	if err != nil {
		return
	}
	post := &model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(format, args...),
	}
	_, _ = i.CreatePost(post)
}

func (i *Client) DMPost(userID string, post *model.Post) (*model.Post, error) {
	channel, err := i.getDirectChannelWith(userID)
	if err != nil {
		return nil, errors.Wrap(err, "getDirectionChannel")
	}
	post.ChannelId = channel.Id
	return i.CreatePost(post)
}

func (i *Client) getDirectChannelWith(userID string) (*model.Channel, error) {
	var channel *model.Channel
	var res *model.Response

	channel, res = i.CreateDirectChannel(i.userID, userID)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}

	return channel, nil
}
