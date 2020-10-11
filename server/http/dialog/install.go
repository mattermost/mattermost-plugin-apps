package dialog

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type installDialogState struct {
	Manifest      *store.Manifest
	TeamID        string
	LogRootPostID string
	LogChannelID  string
}

func NewInstallAppDialog(manifest *store.Manifest, secret, pluginURL string, commandArgs *model.CommandArgs) model.OpenDialogRequest {
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
	if manifest.RequestedPermissions.Contains(store.PermissionActAsUser) {
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
	actingUserID := req.Header.Get("Mattermost-User-Id")
	if actingUserID == "" {
		respondWithError(w, http.StatusUnauthorized, errors.New("user not logged in"))
		return
	}
	// TODO check for sysadmin

	sessionID := req.Header.Get("MM_SESSION_ID")
	if sessionID == "" {
		respondWithError(w, http.StatusUnauthorized, errors.New("no session"))
		return
	}
	session, err := d.apps.Mattermost.Session.Get(sessionID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err)
	}

	var dialogRequest model.SubmitDialogRequest
	err = json.NewDecoder(req.Body).Decode(&dialogRequest)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	if dialogRequest.Type != "dialog_submission" {
		respondWithError(w, http.StatusBadRequest,
			errors.New("expected dialog_submission, got "+dialogRequest.Type))
		return
	}
	if dialogRequest.Cancelled {
		return
	}

	v := dialogRequest.Submission["consent"]
	consentValue, _ := v.(string)
	noUserConsentForOAuth2 := false
	if consentValue == "notrequire" {
		noUserConsentForOAuth2 = true
	}

	v = dialogRequest.Submission["secret"]
	secret, _ := v.(string)

	stateData := installDialogState{}
	err = json.Unmarshal([]byte(dialogRequest.State), &stateData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	app, out, err := d.apps.API.InstallApp(
		&apps.InInstallApp{
			NoUserConsentForOAuth2: noUserConsentForOAuth2,
			AppSecret:              secret,
			GrantedPermissions:     stateData.Manifest.RequestedPermissions,
		},
		&apps.Context{
			ActingUserID: actingUserID,
			AppID:        stateData.Manifest.AppID,
			TeamID:       stateData.TeamID,
		},
		apps.SessionToken(session.Token),
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	_ = d.apps.Mattermost.Post.DM(app.BotUserID, actingUserID, &model.Post{
		Message: out.String(),
	})
}
