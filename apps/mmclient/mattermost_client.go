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
	c := Client{
		userID:   userID,
		ClientPP: NewAPIClientPP(mattermostSiteURL),
		Client4:  model.NewAPIv4Client(mattermostSiteURL),
	}
	c.Client4.SetOAuthToken(token)
	c.ClientPP.SetOAuthToken(token)
	return &c
}

func (c *Client) KVSet(id string, prefix string, in interface{}) (interface{}, error) {
	var mapRes interface{}
	var res *model.Response
	mapRes, res = c.ClientPP.KVSet(id, prefix, in)

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return mapRes, nil
}

func (c *Client) KVGet(id string, prefix string, ref interface{}) error {
	res := c.ClientPP.KVGet(id, prefix, ref)
	if res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return res.Error
		}
		return fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return nil
}

func (c *Client) KVDelete(id string, prefix string) (bool, error) {
	var opRes bool
	var res *model.Response

	opRes, res = c.ClientPP.KVDelete(id, prefix)
	if res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return false, res.Error
		}
		return false, fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return opRes, nil
}

func (c *Client) Subscribe(sub *apps.Subscription) (*apps.SubscriptionResponse, error) {
	var subResponse *apps.SubscriptionResponse
	var res *model.Response

	subResponse, res = c.ClientPP.Subscribe(sub)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}

	return subResponse, nil
}

func (c *Client) Unsubscribe(sub *apps.Subscription) (*apps.SubscriptionResponse, error) {
	var subResponse *apps.SubscriptionResponse
	var res *model.Response

	subResponse, res = c.ClientPP.Unsubscribe(sub)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}

	return subResponse, nil
}

func (c *Client) StoreOAuth2App(appID apps.AppID, clientID, clientSecret string) error {
	res := c.ClientPP.StoreOAuth2App(appID, clientID, clientSecret)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return res.Error
		}
		return fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return nil
}

func (c *Client) StoreOAuth2User(appID apps.AppID, ref interface{}) error {
	res := c.ClientPP.StoreOAuth2User(appID, ref)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return res.Error
		}
		return fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return nil
}

func (c *Client) GetOAuth2User(appID apps.AppID, ref interface{}) error {
	res := c.ClientPP.GetOAuth2User(appID, ref)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return res.Error
		}
		return fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return nil
}

func (c *Client) CreatePost(post *model.Post) (*model.Post, error) {
	var createdPost *model.Post
	var res *model.Response
	post.UserId = c.userID

	createdPost, res = c.Client4.CreatePost(post)
	if res.StatusCode != http.StatusCreated {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}
	return createdPost, nil
}

func (c *Client) DM(userID string, format string, args ...interface{}) {
	channel, err := c.getDirectChannelWith(userID)
	if err != nil {
		return
	}
	post := &model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(format, args...),
	}
	_, _ = c.CreatePost(post)
}

func (c *Client) DMPost(userID string, post *model.Post) (*model.Post, error) {
	channel, err := c.getDirectChannelWith(userID)
	if err != nil {
		return nil, errors.Wrap(err, "getDirectionChannel")
	}
	post.ChannelId = channel.Id
	return c.CreatePost(post)
}

func (c *Client) getDirectChannelWith(userID string) (*model.Channel, error) {
	var channel *model.Channel
	var res *model.Response

	channel, res = c.CreateDirectChannel(c.userID, userID)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if res.Error != nil {
			return nil, res.Error
		}
		return nil, fmt.Errorf("returned with status %d", res.StatusCode)
	}

	return channel, nil
}
