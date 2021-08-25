package dialog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type installDialogState struct {
	AppID         apps.AppID
	ChannelID     string
	TeamID        string
	LogRootPostID string
	LogChannelID  string
}

func NewInstallAppDialog(m *apps.Manifest, secret string, conf config.Config, commandArgs *model.CommandArgs) model.OpenDialogRequest {
	consent := ""
	if m.AppType == apps.AppTypeHTTP {
		consent += fmt.Sprintf("- Access **Remote HTTP API** at `%s` \n", m.HTTPRootURL)
	}
	if len(m.RequestedPermissions) != 0 {
		consent += "- Access **Mattermost API** with the following permissions:\n"
		for _, permission := range m.RequestedPermissions {
			consent += fmt.Sprintf("  - %s\n", permission.String())
		}
	}
	if len(m.RequestedLocations) != 0 {
		consent += "\n- Add the following elements to the **Mattermost User Interface**:\n"
		for _, l := range m.RequestedLocations {
			consent += fmt.Sprintf("  - %s\n", l.Markdown())
		}
	}
	if consent != "" {
		header := fmt.Sprintf("Application **%s** requires system administrator's consent to:\n\n", m.DisplayName)
		consent = header + consent + "---\n"
	}

	elements := []model.DialogElement{}
	if m.AppType == apps.AppTypeHTTP {
		elements = append(elements, model.DialogElement{
			DisplayName: "JWT Secret:",
			Name:        "secret",
			Type:        "text",
			SubType:     "password",
			HelpText: fmt.Sprintf("The JWT Secret authenticates HTTP messages sent to the App. "+
				"It should be obtained from the App itself, %s.",
				m.HomepageURL),
			Default:  secret,
			Optional: true,
		})
	}

	// if m.RequestedPermissions.Contains(apps.PermissionActAsUser) {
	// 	elements = append(elements, model.DialogElement{
	// 		DisplayName: "Require user consent to use REST API first time they use the app:",
	// 		Name:        "consent",
	// 		Type:        "radio",
	// 		Default:     "require",
	// 		HelpText:    "please indicate if user consent is required to allow the app to act on their behalf",
	// 		Options: []*model.PostActionOptions{
	// 			{
	// 				Text:  "Require user consent",
	// 				Value: "require",
	// 			},
	// 			{
	// 				Text:  "Do not require user consent",
	// 				Value: "notrequire",
	// 			},
	// 		},
	// 	})
	// }

	var iconURL string
	if m.Icon != "" {
		iconURL = conf.StaticURL(m.AppID, m.Icon)
	}

	stateData, _ := json.Marshal(installDialogState{
		AppID:     m.AppID,
		TeamID:    commandArgs.TeamId,
		ChannelID: commandArgs.ChannelId,
	})

	return model.OpenDialogRequest{
		TriggerId: commandArgs.TriggerId,
		URL:       conf.PluginURL + path.Join(config.InteractiveDialogPath, InstallPath),
		Dialog: model.Dialog{
			Title:            "Install App - " + m.DisplayName,
			IntroductionText: consent,
			IconURL:          iconURL,
			Elements:         elements,
			SubmitLabel:      "Approve and Install",
			NotifyOnCancel:   true,
			State:            string(stateData),
		},
	}
}

func (d *dialog) handleInstall(w http.ResponseWriter, req *http.Request) {
	_, mm, log := d.conf.Basic()
	actingUserID := req.Header.Get("Mattermost-User-Id")
	if actingUserID == "" {
		respondWithError(w, http.StatusUnauthorized, errors.New("user not logged in"))
		return
	}

	if err := utils.EnsureSysAdmin(mm, actingUserID); err != nil {
		respondWithError(w, http.StatusForbidden, err)
		return
	}

	sessionID := req.Header.Get("MM_SESSION_ID")
	if sessionID == "" {
		respondWithError(w, http.StatusUnauthorized, errors.New("no session"))
		return
	}
	var dialogRequest model.SubmitDialogRequest
	err := json.NewDecoder(req.Body).Decode(&dialogRequest)
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

	client, err := mmclient.NewHTTPClientFromSessionID(d.conf, sessionID, actingUserID)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "invalid session"))
		return
	}

	cc := &apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID:    stateData.TeamID,
			ChannelID: stateData.ChannelID,
		},
		ActingUserID: actingUserID,
		UserID:       actingUserID,
	}
	cc = d.conf.Get().SetContextDefaultsForApp(stateData.AppID, cc)

	_, out, err := d.proxy.InstallApp(client, sessionID, cc, noUserConsentForOAuth2, secret)
	if err != nil {
		log.WithError(err).Warnw("Failed to install app", "app_id", cc.AppID)
		respondWithError(w, http.StatusInternalServerError, err)

		out = fmt.Sprintf("Install failed. Error: **%s**\n", err.Error())
	}

	mm.Post.SendEphemeralPost(actingUserID, &model.Post{
		ChannelId: dialogRequest.ChannelId,
		Message:   out,
	})
}
