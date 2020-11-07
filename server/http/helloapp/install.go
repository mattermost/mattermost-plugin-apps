package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *helloapp) fInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *api.Call) (int, error) {
	if call.Type != api.CallTypeSubmit {
		return http.StatusBadRequest, errors.New("not supported")
	}

	botAccessToken := call.GetValue(constants.BotAccessToken, "")
	oauth2ClientSecret := call.GetValue(constants.OAuth2ClientSecret, "")

	err := h.storeAppCredentials(&appCredentials{
		BotAccessToken:     botAccessToken,
		BotUserID:          call.Context.App.BotUserID,
		OAuth2ClientID:     call.Context.App.OAuth2ClientID,
		OAuth2ClientSecret: oauth2ClientSecret,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = h.InitOAuther()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	connectURL, err := h.startOAuth2Connect(call.Context.ActingUserID, &api.Call{
		URL:     h.AppURL(PathConnectedInstall),
		Context: call.Context,
		Expand: &api.Expand{
			App:    api.ExpandAll,
			Config: api.ExpandSummary,
		},
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSON(w,
		api.CallResponse{
			Type: api.CallResponseTypeOK,
			Markdown: md.Markdownf(
				"**Hallo სამყარო** needs to continue its installation using your system administrator's credentials. Please [connect](%s) the application to your Mattermost account.",
				connectURL),
		})
	return http.StatusOK, nil
}
