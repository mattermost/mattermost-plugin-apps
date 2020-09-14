package dialog

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/utils"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/apps"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/constants"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func NewInstallAppDialog(triggerID string, manifest *apps.Manifest, pluginURL string, postID string) model.OpenDialogRequest {
	intro := md.Bold(
		md.Markdownf("Application %s requires the following permissions:", manifest.DisplayName)) + "\n"
	for _, permission := range manifest.RequestedPermissions {
		intro += md.Markdownf("- %s\n", permission.Markdown())
	}
	intro += "\n---\n"

	var elements []model.DialogElement
	if manifest.RequestedPermissions.Contains(apps.PermissionActAsUser) {
		elements = []model.DialogElement{
			{
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
			},
		}
	}

	return model.OpenDialogRequest{
		TriggerId: triggerID,
		URL:       pluginURL + constants.InteractiveDialogPath + InstallPath,
		Dialog: model.Dialog{
			CallbackId:       postID,
			Title:            "Install App - " + manifest.DisplayName,
			IntroductionText: intro.String(),
			Elements:         elements,
			SubmitLabel:      "Approve and Install",
			NotifyOnCancel:   true,
			State:            utils.ToJSON(manifest),
		},
	}
}

func (d *dialog) handleInstall(w http.ResponseWriter, req *http.Request) {
	var err error
	rootID := ""
	message := ""
	actingUserID := ""
	status := http.StatusOK
	defer func() {
		conf := d.apps.Config.GetConfig()
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
				RootId:   rootID,
				ParentId: rootID,
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

	v := dialogRequest.Submission["consent"]
	consentValue, _ := v.(string)
	noUserConsentForOAuth2 := false
	if consentValue == "require" {
		noUserConsentForOAuth2 = true
	}

	rootID = dialogRequest.CallbackId

	var manifest apps.Manifest
	err = json.Unmarshal([]byte(dialogRequest.State), &manifest)
	if err != nil {
		err = errors.Wrap(err, "failed to unmarshal manifest as state")
		return
	}

	_, err = d.apps.Registry.InstallApp(&apps.InInstallApp{
		ActingMattermostUserID: actingUserID,
		NoUserConsentForOAuth2: noUserConsentForOAuth2,
		Manifest:               &manifest,
	})
	if err != nil {
		status = http.StatusInternalServerError
		return
	}

	message = "Installed " + manifest.DisplayName
}
