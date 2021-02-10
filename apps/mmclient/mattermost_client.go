package mmclient

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Client struct {
	*model.Client4
	*ClientPP
	userID string
}

func as(id, token string, cc *apps.Context) *Client {
	return NewClient(id, token, cc.MattermostSiteURL)
}

func AsBot(cc *apps.Context) *Client {
	return as(cc.BotUserID, cc.BotAccessToken, cc)
}

func AsActingUser(cc *apps.Context) *Client {
	return as(cc.ActingUserID, cc.ActingUserAccessToken, cc)
}

func AsAdmin(cc *apps.Context) *Client {
	return as(cc.ActingUserID, cc.AdminAccessToken, cc)
}

func NewClient(userID, token, mattermostSiteURL string) *Client {
	client := Client{
		userID:   userID,
		ClientPP: NewAPIClientPP(mattermostSiteURL),
		Client4:  model.NewAPIv4Client(mattermostSiteURL),
	}
	client.Client4.SetOAuthToken(token)
	return &client
}

func (client *Client) KVSet(id string, prefix string, in map[string]interface{}) (map[string]interface{}, error) {
	var mapRes map[string]interface{}
	var res *model.Response
	mapRes, res = client.ClientPP.KVSet(id, prefix, in)

	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return mapRes, nil
}

func (client *Client) KVGet(id string, prefix string) (map[string]interface{}, error) {
	var mapRes map[string]interface{}
	var res *model.Response

	mapRes, res = client.ClientPP.KVGet(id, prefix)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return mapRes, nil
}

func (client *Client) KVDelete(id string, prefix string) (bool, error) {
	var opRes bool
	var res *model.Response

	opRes, res = client.ClientPP.KVDelete(id, prefix)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return false, res.Error
		}
		return false, fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return opRes, nil
}

func (client *Client) Subscribe(sub *apps.Subscription) (*apps.SubscriptionResponse, error) {
	var subResponse *apps.SubscriptionResponse
	var res *model.Response

	subResponse, res = client.ClientPP.Subscribe(sub)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}

	return subResponse, nil
}

func (client *Client) Unsubscribe(sub *apps.Subscription) (*apps.SubscriptionResponse, error) {
	var subResponse *apps.SubscriptionResponse
	var res *model.Response

	subResponse, res = client.ClientPP.Unsubscribe(sub)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}

	return subResponse, nil
}

func (client *Client) CreatePost(post *model.Post) (*model.Post, error) {
	var createdPost *model.Post
	var res *model.Response
	post.UserId = client.userID

	createdPost, res = client.Client4.CreatePost(post)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return createdPost, nil
}

func (client *Client) DM(userID string, format string, args ...interface{}) {
	channel, err := client.getDirectChannelWith(userID)
	if err != nil {
		return
	}
	post := &model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(format, args...),
	}
	_, _ = client.CreatePost(post)
}

func (client *Client) DMPost(userID string, post *model.Post) (*model.Post, error) {
	channel, err := client.getDirectChannelWith(userID)
	if err != nil {
		return nil, errors.Wrap(err, "getDirectionChannel")
	}
	post.ChannelId = channel.Id
	return client.CreatePost(post)
}

func (client *Client) getDirectChannelWith(userID string) (*model.Channel, error) {
	var channel *model.Channel
	var res *model.Response

	channel, res = client.CreateDirectChannel(client.userID, userID)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}

	return channel, nil
}
