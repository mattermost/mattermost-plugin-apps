package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/constants"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

// Install function metadata is not necessary, but fillint it out (minimally)
// for demo purposes. Install does not bind to any locations, it's Expand is
// pre-determined by the server.
func (h *helloapp) fInstallMeta(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, cc *api.Context) (int, error) {
	httputils.WriteJSON(w,
		&api.Function{
			Form: &api.Form{
				Fields: []*api.Field{
					{
						Name:       constants.BotAccessToken,
						Type:       api.FieldTypeText,
						IsRequired: true,
					}, {
						Name:       constants.OAuth2ClientSecret,
						Type:       api.FieldTypeText,
						IsRequired: true,
					},
				},
			},
		})
	return http.StatusOK, nil
}

func (h *helloapp) fInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *api.Call) (int, error) {
	err := h.storeAppCredentials(&appCredentials{
		BotAccessToken:     call.Values[constants.BotAccessToken],
		BotUserID:          call.Context.App.BotUserID,
		OAuth2ClientID:     call.Context.App.OAuth2ClientID,
		OAuth2ClientSecret: call.Values[constants.OAuth2ClientSecret],
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = h.initOAuther()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	connectURL, err := h.startOAuth2Connect(call.Context.ActingUserID, &api.Call{
		URL:     h.appURL(pathConnectedInstall),
		Context: call.Context,
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
