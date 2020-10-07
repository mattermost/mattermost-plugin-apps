package dialog

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type installDialogState struct {
	Manifest      *apps.Manifest
	TeamID        string
	LogRootPostID string
	LogChannelID  string
}

func NewInstallAppDialog(manifest *apps.Manifest, secret, pluginURL string, commandArgs *model.CommandArgs) model.OpenDialogRequest {
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
			Default:     secret,
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

	stateData, _ := json.Marshal(installDialogState{
		Manifest: manifest,
		TeamID:   commandArgs.TeamId,
	})

	return model.OpenDialogRequest{
		TriggerId: commandArgs.TriggerId,
		URL:       pluginURL + constants.InteractiveDialogPath + InstallPath,
		Dialog: model.Dialog{
			Title:            "Install App - " + manifest.DisplayName,
			IntroductionText: intro.String(),
			Elements:         elements,
			SubmitLabel:      "Approve and Install",
			NotifyOnCancel:   true,
			State:            string(stateData),
		},
	}
}

func (d *dialog) handleInstall(w http.ResponseWriter, req *http.Request) {
	var err error
	stateData := installDialogState{}
	logMessage := ""
	actingUserID := ""
	status := http.StatusInternalServerError

	defer func() {
		resp := model.SubmitDialogResponse{}
		if err != nil {
			resp.Error = errors.Wrap(err, "failed to install").Error()
			logMessage = "Error: " + resp.Error
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(resp)

		conf := d.apps.Configurator.GetConfig()
		_ = d.apps.Mattermost.Post.DM(conf.BotUserID, actingUserID, &model.Post{
			Message: logMessage,
		})
	}()

	actingUserID = req.Header.Get("Mattermost-User-Id")
	if actingUserID == "" {
		err = errors.New("user not logged in")
		status = http.StatusUnauthorized
		return
	}
	// <><> TODO check for sysadmin

	sessionID := req.Header.Get("MM_SESSION_ID")
	if sessionID == "" {
		err = errors.New("no session")
		status = http.StatusUnauthorized
		return
	}

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

	if dialogRequest.Cancelled {
		logMessage = "Installation was canceled by the user"
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

	err = json.Unmarshal([]byte(dialogRequest.State), &stateData)
	if err != nil {
		status = http.StatusBadRequest
		return
	}

	_, out, err := d.apps.API.InstallApp(&apps.InInstallApp{
		Context: apps.CallContext{
			ActingUserID: actingUserID,
			AppID:        stateData.Manifest.AppID,
			TeamID:       stateData.TeamID,
			LogTo: &apps.Thread{
				ChannelID:  stateData.LogChannelID,
				RootPostID: stateData.LogRootPostID,
			},
		},
		App: apps.App{
			Manifest:               stateData.Manifest,
			NoUserConsentForOAuth2: noUserConsentForOAuth2,
			Secret:                 secret,
		},
		GrantedPermissions: stateData.Manifest.RequestedPermissions,
	})
	if err != nil {
		return
	}

	status = http.StatusOK
	logMessage = out.String()
}
