package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const fieldUserID = "userID"
const fieldMessage = "message"

func (h *helloapp) Message(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *api.Call) (int, error) {
	userID := c.Values[fieldUserID]
	message := c.Values[fieldMessage]
	var out *api.CallResponse

	switch c.Type {
	case api.CallTypeForm:
		out = &api.CallResponse{
			Type: api.CallResponseTypeForm,
			Form: &api.Form{
				Title:  "Send a message to user",
				Header: "Message modal form header",
				Footer: "Message modal form footer",
				Fields: []*api.Field{
					{
						Name:              fieldUserID,
						Type:              api.FieldTypeUser,
						Description:       "User to send the message to",
						AutocompleteLabel: "user",
						AutocompleteHint:  "enter user ID or @user",
						ModalLabel:        "User",
					}, {
						Name:              fieldMessage,
						Type:              api.FieldTypeText,
						IsRequired:        true,
						Description:       "Message that will be sent to the user",
						AutocompleteLabel: "$1",
						AutocompleteHint:  "Anything you want to say",
						ModalLabel:        "Message to send",
						TextMinLength:     2,
						TextMaxLength:     1024,
					},
				},
			},
		}

	case api.CallTypeSubmit:
		if userID == "" {
			userID = c.Context.ActingUserID
		}
		if message == "" {
			message = "Hello"
		}
		h.message(userID, message)
		out = &api.CallResponse{}
	}
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) message(userID, message string) {
	h.dm(userID, "PING %s", message)
}
