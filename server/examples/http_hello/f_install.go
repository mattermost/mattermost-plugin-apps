package http_hello

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *helloapp) fInstall(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	if c.Type != api.CallTypeSubmit {
		return http.StatusBadRequest, errors.New("not supported")
	}

	botAccessToken := c.GetValue(api.PropBotAccessToken, "")
	oauth2ClientSecret := c.GetValue(api.PropOAuth2ClientSecret, "")

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

	connectURL, err := h.startOAuth2Connect(c.Context.ActingUserID, &api.Call{
		URL:     h.appURL(PathConnectedInstall),
		Context: c.Context,
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
