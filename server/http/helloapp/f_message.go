package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const FieldUserID = "userID"
const FieldMessage = "message"

func (h *helloapp) fMessageMeta(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, cc *api.Context) (int, error) {
	httputils.WriteJSON(w, api.Function{
		Form: &api.Form{
			Title:  "Message to user",
			Header: "Message modal form header",
			Footer: "Message modal form footer",
			Fields: []*api.Field{
				{
					Name:              FieldUserID,
					Type:              api.FieldTypeUser,
					Description:       "User to send the message to",
					AutocompleteLabel: "user",
					AutocompleteHint:  "enter user ID or @user",
					ModalLabel:        "User",
				}, {
					Name:              FieldMessage,
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
		Expand: &api.Expand{},
	})
	return http.StatusOK, nil
}

func (h *helloapp) fMessage(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *api.Call) (int, error) {
	userID := call.Values[FieldUserID]
	if userID == "" {
		userID = call.Context.ActingUserID
	}

	h.message(userID, call.Values[FieldMessage])

	httputils.WriteJSON(w, api.CallResponse{
		Type: api.CallResponseTypeOK,
	})
	return http.StatusOK, nil
}

func (h *helloapp) message(userID, message string) {
	h.dm(userID, "PING %s", message)
}
