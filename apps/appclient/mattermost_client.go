package appclient

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Client struct {
	*model.Client4
	*ClientPP
	userID string
}

func as(id, token string, cc apps.Context) *Client {
	return NewClient(id, token, cc.MattermostSiteURL)
}

func AsBot(cc apps.Context) *Client {
	return as(cc.BotUserID, cc.BotAccessToken, cc)
}

func AsActingUser(cc apps.Context) *Client {
	return as(cc.ActingUser.Id, cc.ActingUserAccessToken, cc)
}

func NewClient(userID, token, mattermostSiteURL string) *Client {
	c := Client{
		userID:   userID,
		ClientPP: NewAppsPluginAPIClient(mattermostSiteURL),
		Client4:  model.NewAPIv4Client(mattermostSiteURL),
	}
	c.Client4.SetOAuthToken(token)
	c.ClientPP.SetOAuthToken(token)
	return &c
}

func (c *Client) KVSet(prefix, id string, in interface{}) (bool, error) {
	changed, res, err := c.ClientPP.KVSet(prefix, id, in)

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return false, err
		}

		return false, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return changed, nil
}

func (c *Client) KVGet(prefix, id string, ref interface{}) error {
	res, err := c.ClientPP.KVGet(prefix, id, ref)
	if res.StatusCode != http.StatusOK {
		if err != nil {
			return err
		}

		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) KVDelete(prefix, id string) error {
	res, err := c.ClientPP.KVDelete(prefix, id)
	if res.StatusCode != http.StatusOK {
		if err != nil {
			return err
		}

		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) Subscribe(sub *apps.Subscription) error {
	res, err := c.ClientPP.Subscribe(sub)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return err
		}

		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) GetSubscriptions() ([]apps.Subscription, error) {
	subs, res, err := c.ClientPP.GetSubscriptions()
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return nil, err
		}

		return nil, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return subs, nil
}

func (c *Client) Unsubscribe(sub *apps.Subscription) error {
	res, err := c.ClientPP.Unsubscribe(sub)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return err
		}

		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) StoreOAuth2App(oauth2App apps.OAuth2App) error {
	res, err := c.ClientPP.StoreOAuth2App(oauth2App)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return err
		}

		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) StoreOAuth2User(ref interface{}) error {
	res, err := c.ClientPP.StoreOAuth2User(ref)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return err
		}
		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) GetOAuth2User(ref interface{}) error {
	res, err := c.ClientPP.GetOAuth2User(ref)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return err
		}

		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) Call(creq apps.CallRequest) (*apps.CallResponse, error) {
	cresp, res, err := c.ClientPP.Call(creq)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return nil, err
		}

		return nil, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return cresp, nil
}

func (c *Client) CreatePost(post *model.Post) (*model.Post, error) {
	post.UserId = c.userID

	createdPost, res, err := c.Client4.CreatePost(post)
	if res.StatusCode != http.StatusCreated {
		if err != nil {
			return nil, err
		}

		return nil, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return createdPost, nil
}

func (c *Client) DM(userID string, format string, args ...interface{}) (*model.Post, error) {
	channel, err := c.getDirectChannelWith(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get direct channel to post DM")
	}

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(format, args...),
	}
	return c.CreatePost(post)
}

func (c *Client) DMPost(userID string, post *model.Post) (*model.Post, error) {
	channel, err := c.getDirectChannelWith(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get direct channel")
	}

	post.ChannelId = channel.Id
	return c.CreatePost(post)
}

func (c *Client) getDirectChannelWith(userID string) (*model.Channel, error) {
	channel, res, err := c.CreateDirectChannel(c.userID, userID)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		if err != nil {
			return nil, err
		}

		return nil, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return channel, nil
}
