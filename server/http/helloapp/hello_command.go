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

type SubCommand struct {
	AppID       string        `json:"app_id"`
	Pretext     string        `json:"pretext"`
	Description string        `json:"description"`
	SubCommands []SubCommand  `json:"sub_commands"`
	Args        []interface{} `json:"args"`
	FormURL     string        `json:"form_url"`
}

func (h *helloapp) handleHelloCommandSubmission(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *apps.Call) (int, error) {
	var e = func(err error, msg string) (int, error) {
		httputils.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeError,
			Markdown: md.MD(errors.Wrap(err, msg).Error()),
		})
		return http.StatusInternalServerError, err
	}

	username := call.Values.Get("user")
	message := call.Values.Get("message")
	u, err := h.apps.Mattermost.User.GetByUsername(username)
	if err != nil {
		return e(err, "get by username "+username)
	}

	fullMessage := md.Markdownf("Hey there! Your teammate %s sent you a message: \"%s\"", call.Context.ActingUserID, message)
	h.DM(u.Id, fullMessage.String())
	h.Ephemeral(call.Context.ActingUserID, call.Context.ChannelID, fmt.Sprintf("Sent \"%s\" to @%s", message, username))
	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: md.Markdownf("Looks like you've submitted the form!"),
		})

	return 200, nil
}

func (h *helloapp) handleHelloCommandDefinition(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *apps.Call) (int, error) {
	data := SubCommand{
		Pretext: "/hello",
		AppID:   AppID,
		SubCommands: []SubCommand{
			{
				AppID:       AppID,
				Pretext:     "sendto",
				FormURL:     PathCommandSubmission,
				Description: "Send a message to someone in this channel",
				Args: []interface{}{
					apps.AutocompleteDynamicSelect{
						AutocompleteProps: apps.AutocompleteProps{
							ElementProps: &apps.ElementProps{
								Name:        "user",
								Description: "User to send the message to",
								Type:        apps.ElementTypeDynamicSelect,
							},
							AutocompleteElementProps: &apps.AutocompleteElementProps{
								Hint:       "[username]",
								Positional: true,
							},
						},
						DynamicSelectElementProps: apps.DynamicSelectElementProps{
							RefreshURL: "/lookup/users",
						},
					},
					apps.AutocompleteText{
						AutocompleteProps: apps.AutocompleteProps{
							ElementProps: &apps.ElementProps{
								Name:        "message",
								Description: "The message to send",
								Type:        apps.ElementTypeText,
							},
							AutocompleteElementProps: &apps.AutocompleteElementProps{
								FlagName:   "message",
								Positional: false,
							},
						},
					},
				},
			},
		},
	}

	res := apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: data,
	}

	httputils.WriteJSON(w, res)

	return 200, nil
}

func (h *helloapp) handleLookupUsers(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *apps.Call) (int, error) {
	cmdStr := md.CodeBlock(call.Values.Raw)
	cj := md.JSONBlock(call)

	var e = func(err error) (int, error) {
		h.DM(call.Context.ActingUserID, err.Error())
		httputils.WriteJSON(w, apps.CallResponse{
			Type:  apps.CallResponseTypeError,
			Error: err,
		})
		return http.StatusInternalServerError, err
	}

	users, err := h.apps.Mattermost.User.ListInChannel(call.Context.ChannelID, "username", 0, 50)
	if err != nil {
		return e(err)
	}

	search := call.Values.Get("user")

	data := []*apps.SelectOption{}
	for _, user := range users {
		if !strings.HasPrefix(user.Username, search) {
			continue
		}

		data = append(data, &apps.SelectOption{
			Label:    user.GetDisplayName(model.SHOW_FULLNAME),
			Value:    user.Username,
			IconData: fmt.Sprintf("/api/v4/users/%s/image?_=0", user.Id),
		})
	}

	res := apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: data,
	}

	httputils.WriteJSON(w, res)

	if false {
		s := md.Markdownf("We ran\n%s\n%s\n%s", cmdStr, cj, md.JSONBlock(res)).String()
		h.DM(call.Context.ActingUserID, s)
	}

	return 200, nil
}
