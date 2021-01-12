package helloapp

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *helloapp) fInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.Call) (int, error) {
	if c.Type != apps.CallTypeSubmit {
		return http.StatusBadRequest, errors.New("not supported")
	}

	botAccessToken := c.GetValue(apps.PropBotAccessToken, "")
	oauth2ClientSecret := c.GetValue(apps.PropOAuth2ClientSecret, "")

	err := h.storeAppCredentials(&appCredentials{
		BotAccessToken:     botAccessToken,
		BotUserID:          c.Context.App.BotUserID,
		OAuth2ClientID:     c.Context.App.OAuth2ClientID,
		OAuth2ClientSecret: oauth2ClientSecret,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = h.initOAuther()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	connectURL, err := h.startOAuth2Connect(c.Context.ActingUserID, &apps.Call{
		URL:     h.appURL(PathConnectedInstall),
		Context: c.Context,
		Expand: &apps.Expand{
			App:    apps.ExpandAll,
			Config: apps.ExpandSummary,
		},
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type: apps.CallResponseTypeOK,
			Markdown: md.Markdownf(
				"**Hallo სამყარო** needs to continue its installation using your system administrator's credentials. Please [connect](%s) the application to your Mattermost account.",
				connectURL),
		})
	return http.StatusOK, nil
}
