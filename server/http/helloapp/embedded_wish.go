package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (h *helloapp) handleCreateEmbedded(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallRequest) (int, error) {
	err := h.asBot(func(mmclient *model.Client4, botUserID string) error {
		post := &model.Post{
			Message:   "Debug form",
			ChannelId: data.Context.ChannelID,
			UserId:    botUserID,
		}
		post.AddProp("appID", AppID)
		post.AddProp("dialog", h.getDialogSmallSample())

		_, resp := mmclient.CreatePost(post)
		if resp.Error != nil {
			return resp.Error
		}
		return nil
	})

	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSONStatus(w, http.StatusOK, nil)
	return http.StatusOK, nil
}

func (h *helloapp) handleSubmitEmbedded(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallRequest) (int, error) {
	response := apps.CallResponse{
		Type: apps.ResponseTypeOK,
		Data: make(map[string]interface{}),
	}
	post := &model.Post{
		Message: "Submitted",
		Props:   model.StringInterface{},
	}
	response.Data["post"] = post
	// response := apps.CallResponse{
	// 	Type:  apps.ResponseTypeError,
	// 	Data:  make(map[string]interface{}),
	// 	Error: "Some error",
	// }

	// errors := map[string]string{}
	// for key := range data.Values.Data {
	// 	errors[key] = "Some other error"
	// }
	// response.Data["errors"] = errors

	httputils.WriteJSON(w, response)
	return http.StatusOK, nil
}

func (h *helloapp) getDialogSmallSample() model.OpenDialogRequest {
	return model.OpenDialogRequest{
		URL: h.AppURL(PathWishSubmitEmbedded),
		Dialog: model.Dialog{
			Title:   "Title for Small Dialog Test",
			IconURL: "http://www.mattermost.org/wp-content/uploads/2016/04/icon.png",
			Elements: []model.DialogElement{
				{
					DisplayName: "Display Name",
					Name:        "realname",
					Type:        "text",
					Default:     "default text",
					Placeholder: "placeholder",
					HelpText:    "This a regular input in an interactive dialog triggered by a test integration.",
				},
			},
		},
	}
}

func (h *helloapp) getDialogFullSample() model.OpenDialogRequest {
	return model.OpenDialogRequest{
		URL: h.AppURL(PathWishSubmitEmbedded),
		Dialog: model.Dialog{
			Title:   "Title for Full Dialog Test",
			IconURL: "http://www.mattermost.org/wp-content/uploads/2016/04/icon.png",
			Elements: []model.DialogElement{
				{
					DisplayName: "Display Name",
					Name:        "realname",
					Type:        "text",
					Default:     "default text",
					Placeholder: "placeholder",
					HelpText:    "This a regular input in an interactive dialog triggered by a test integration.",
				},
				{
					DisplayName: "Email",
					Name:        "someemail",
					Type:        "text",
					SubType:     "email",
					Placeholder: "placeholder@bladekick.com",
					HelpText:    "This a regular email input in an interactive dialog triggered by a test integration.",
				},
				{
					DisplayName: "Number",
					Name:        "somenumber",
					Type:        "text",
					SubType:     "number",
				},
				{
					DisplayName: "Password",
					Name:        "somepassword",
					Type:        "text",
					SubType:     "password",
					Default:     "p@ssW0rd",
					Placeholder: "placeholder",
					HelpText:    "This a password input in an interactive dialog triggered by a test integration.",
				},
				{
					DisplayName: "Display Name Long Text Area",
					Name:        "realnametextarea",
					Type:        "textarea",
					Placeholder: "placeholder",
					Optional:    true,
					MinLength:   5,
					MaxLength:   100,
				},
				{
					DisplayName: "User Selector",
					Name:        "someuserselector",
					Type:        "select",
					Placeholder: "Select a user...",
					DataSource:  "users",
				},
				{
					DisplayName: "Channel Selector",
					Name:        "somechannelselector",
					Type:        "select",
					Placeholder: "Select a channel...",
					HelpText:    "Choose a channel from the list.",
					DataSource:  "users",
					Optional:    true,
				},
				{
					DisplayName: "Option Selector",
					Name:        "someoptionselector",
					Type:        "select",
					Placeholder: "Select an option...",
					HelpText:    "Choose a channel from the list.",
					Options: []*model.PostActionOptions{
						{
							Text:  "Option1",
							Value: "opt1",
						},
						{
							Text:  "Option2",
							Value: "opt2",
						},
						{
							Text:  "Option3",
							Value: "opt3",
						},
					},
				},
				{
					DisplayName: "Radio Option Selector",
					Name:        "someradiooptions",
					Type:        "radio",
					Options: []*model.PostActionOptions{
						{
							Text:  "Engineering",
							Value: "engineering",
						},
						{
							Text:  "Sales",
							Value: "sales",
						},
					},
				},
				{
					DisplayName: "Boolean Selector",
					Placeholder: "Was this modal helpful?",
					Name:        "boolean_input",
					Type:        "bool",
					Default:     "True",
					Optional:    true,
					HelpText:    "This is the help text",
				},
			},
			SubmitLabel:    "Submit",
			NotifyOnCancel: true,
			State:          "somestate",
		},
	}
}
