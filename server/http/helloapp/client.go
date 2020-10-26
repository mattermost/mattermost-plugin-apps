package helloapp

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

func (h *helloapp) asUser(userID string, f func(*model.Client4) error) error {
	t, err := h.OAuther.GetToken(userID)
	if err != nil {
		return err
	}
	mmClient := model.NewAPIv4Client(h.apps.Configurator.GetConfig().MattermostSiteURL)
	mmClient.SetOAuthToken(t.AccessToken)
	return f(mmClient)
}

func (h *helloapp) asBot(f func(mmclient *model.Client4, botUserID string) error) error {
	creds, err := h.getAppCredentials()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve app bot credentials")
	}

	mmClient := model.NewAPIv4Client(h.apps.Configurator.GetConfig().MattermostSiteURL)
	mmClient.SetToken(creds.BotAccessToken)

	return f(mmClient, creds.BotUserID)
}

func (h *helloapp) getPost(postID string, actingUserID string) (*model.Post, error) {
	var post *model.Post
	// It would be better to make this as the user, so we don't have problems of users seeing messages they should not
	err := h.asBot(func(mmclient *model.Client4, botUserID string) error {
		var res *model.Response
		post, res = mmclient.GetPost(postID, "")
		if res.StatusCode != http.StatusOK {
			if res.Error != nil {
				return res.Error
			}
			return fmt.Errorf("returned with status %d", res.StatusCode)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (h *helloapp) postAsBot(post *model.Post) (*model.Post, error) {
	var createdPost *model.Post
	err := h.asBot(func(mmclient *model.Client4, botUserID string) error {
		var res *model.Response
		post.UserId = botUserID

		createdPost, res = mmclient.CreatePost(post)
		if res.StatusCode != http.StatusCreated {
			if res.Error != nil {
				return res.Error
			}
			return fmt.Errorf("returned with status %d", res.StatusCode)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return createdPost, nil
}

func (h *helloapp) dm(userID string, format string, args ...interface{}) (*model.Post, error) {
	channel, err := h.getDirectChannelWithBot(userID)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get direct channel")
	}

	post := &model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(format, args...),
	}

	createdPost, err := h.postAsBot(post)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create the post")
	}

	return createdPost, nil
}

func (h *helloapp) dmPost(userID string, post *model.Post) (*model.Post, error) {
	channel, err := h.getDirectChannelWithBot(userID)
	if err != nil {
		return nil, err
	}

	post.ChannelId = channel.Id

	createdPost, err := h.postAsBot(post)
	if err != nil {
		return nil, err
	}
	return createdPost, nil
}

func (h *helloapp) getDirectChannelWithBot(userID string) (*model.Channel, error) {
	var channel *model.Channel
	err := h.asBot(func(mmclient *model.Client4, botUserID string) error {
		var res *model.Response

		channel, res = mmclient.CreateDirectChannel(botUserID, userID)
		if res.StatusCode != http.StatusCreated {
			if res.Error != nil {
				return res.Error
			}
			return fmt.Errorf("returned with status %d", res.StatusCode)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (h *helloapp) getUser(userID string) (*model.User, error) {
	var user *model.User
	err := h.asBot(func(mmclient *model.Client4, botUserID string) error {
		var res *model.Response

		user, res = mmclient.GetUser(userID, "")
		if res.StatusCode != http.StatusOK {
			if res.Error != nil {
				return res.Error
			}
			return fmt.Errorf("returned with status %d", res.StatusCode)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}
