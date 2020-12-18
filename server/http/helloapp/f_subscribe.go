package helloapp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

func (h *helloapp) fSubscribe(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.Call) (int, error) {
	if c.Type != apps.CallTypeSubmit && c.Type != apps.CallTypeForm {
		return http.StatusBadRequest, errors.Errorf("call type not supported: %v", c.Type)
	}

	if c.Type == apps.CallTypeForm {
		out := apps.CallResponse{
			Type: apps.CallResponseTypeForm,
			Form: &apps.Form{
				Fields: []*apps.Field{
					{
						Name:             "channel",
						Type:             apps.FieldTypeChannel,
						Label:            "channel",
						ModalLabel:       "Channel",
						Description:      "The channel to subscribe to",
						AutocompleteHint: "",
					},
				},
			},
		}
		httputils.WriteJSON(w, out)
		return http.StatusOK, nil
	}

	channelName := c.GetValue("channel", "")
	if channelName == "" {
		out := apps.CallResponse{
			Type:  apps.CallResponseTypeError,
			Error: "Missing channel in form submission",
		}

		httputils.WriteJSON(w, out)
		return http.StatusBadRequest, nil
	}

	channelName = strings.TrimPrefix(channelName, "~")
	err := h.asUser(c.Context.ActingUserID, func(client *model.Client4) error {
		ch, res := client.GetChannelByName(channelName, c.Context.TeamID, "")
		if res.Error != nil {
			return errors.Wrapf(res.Error, "error fetching channel %v from team %v", channelName, c.Context.TeamID)
		}

		_, res = client.CreatePost(&model.Post{
			UserId:    c.Context.ActingUserID,
			ChannelId: ch.Id,
			Message:   "I set the subscription mode to " + c.GetValue("mode", ""),
		})
		if res.Error != nil {
			return errors.Wrap(res.Error, "error creating post")
		}

		return nil
	})

	if err != nil {
		out := apps.CallResponse{
			Type:  apps.CallResponseTypeError,
			Error: fmt.Sprintf("Error making post to channel %v. err=%v", channelName, err),
		}

		httputils.WriteJSON(w, out)
		return http.StatusBadRequest, nil
	}
	msg := md.Markdownf("Set subscription status to %v for channel %v", c.Values["mode"], channelName)
	out := apps.CallResponse{Markdown: msg}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}
