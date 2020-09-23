package dialog

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func NewInstallAppDialog(
	triggerID string,
	manifest *apps.Manifest,
	manifestURL string,
	pluginURL string,
	logChannelID, logRootPostID string,
) model.OpenDialogRequest {

	intro := md.Bold(
		md.Markdownf("Application %s requires the following permissions:", manifest.DisplayName)) + "\n"
	for _, permission := range manifest.RequestedPermissions {
		intro += md.Markdownf("- %s\n", permission.Markdown())
	}
	intro += "\n---\n"

	elements := []model.DialogElement{
		{
			DisplayName: "App secret:",
			Name:        "secret",
			Type:        "text",
			SubType:     "password",
			HelpText:    "TODO: How to obtain the App Secret",
		},
	}
	if manifest.RequestedPermissions.Contains(apps.PermissionActAsUser) {
		elements = append(elements, model.DialogElement{
			DisplayName: "Require user consent to use REST API first time they use the app:",
			Name:        "consent",
			Type:        "radio",
			Default:     "require",
			HelpText:    "please indicate if user consent is required to allow the app to act on their behalf",
			Options: []*model.PostActionOptions{
				{
					Text:  "Require user consent",
					Value: "require",
				},
				{
					Text:  "Do not require user consent",
					Value: "notrequire",
				},
			},
		})
	}

	return model.OpenDialogRequest{
		TriggerId: triggerID,
		URL:       pluginURL + constants.InteractiveDialogPath + InstallPath,
		Dialog: model.Dialog{
			CallbackId:       logChannelID + "/" + logRootPostID,
			Title:            "Install App - " + manifest.DisplayName,
			IntroductionText: intro.String(),
			Elements:         elements,
			SubmitLabel:      "Approve and Install",
			NotifyOnCancel:   true,
			State:            manifestURL,
		},
	}
}

func (d *dialog) handleInstall(w http.ResponseWriter, req *http.Request) {
	var err error
	logChannelID, logRootPostID := "", ""
	message := ""
	actingUserID := ""
	status := http.StatusOK

	defer func() {
		conf := d.apps.Configurator.GetConfig()
		resp := model.SubmitDialogResponse{}
		if err != nil {
			resp.Error = errors.Wrap(err, "failed to install").Error()
			message = "Error: " + resp.Error
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(resp)

		if actingUserID != "" {
			_ = d.apps.Mattermost.Post.DM(conf.BotUserID, actingUserID, &model.Post{
				RootId:   logRootPostID,
				ParentId: logRootPostID,
				Message:  message,
			})
		}
	}()

	actingUserID = req.Header.Get("Mattermost-User-Id")
	if actingUserID == "" {
		err = errors.New("user not logged in")
		status = http.StatusUnauthorized
		return
	}
	// <><> TODO check for sysadmin

	var dialogRequest model.SubmitDialogRequest
	err = json.NewDecoder(req.Body).Decode(&dialogRequest)
	if err != nil {
		status = http.StatusBadRequest
		return
	}
	if dialogRequest.Type != "dialog_submission" {
		err = errors.New("expected dialog_submission, got " + dialogRequest.Type)
		status = http.StatusBadRequest
		return
	}

	ss := strings.Split(dialogRequest.CallbackId, "/")
	if len(ss) != 2 {
		err = errors.New("expected channelId/postId as CallbackId, got " + dialogRequest.CallbackId)
		status = http.StatusBadRequest
		return
	}
	logChannelID, logRootPostID = ss[0], ss[1]

	if dialogRequest.Cancelled {
		message = "Installation was canceled by the user"
		return
	}

	v := dialogRequest.Submission["consent"]
	consentValue, _ := v.(string)
	noUserConsentForOAuth2 := false
	if consentValue == "require" {
		noUserConsentForOAuth2 = true
	}

	v = dialogRequest.Submission["secret"]
	secret, _ := v.(string)

	manifest, err := d.apps.AppClient.GetManifest(dialogRequest.State)
	if err != nil {
		err = errors.Wrapf(err, "failed to get manifest from %s", dialogRequest.State)
		return
	}

	out, err := d.apps.API.InstallApp(&apps.InInstallApp{
		ActingMattermostUserID: actingUserID,
		App: &apps.App{
			Manifest:               manifest,
			NoUserConsentForOAuth2: noUserConsentForOAuth2,
			Secret:                 secret,
		},
		LogChannelID:  logChannelID,
		LogRootPostID: logRootPostID,
	})
	if err != nil {
		status = http.StatusInternalServerError
		return
	}

	message = out.String()
}
