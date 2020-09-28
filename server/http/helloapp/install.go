package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (h *helloapp) handleInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallData) (int, error) {
	err := h.storeOAuth2AppCredentials(data.Expanded.App.OAuthAppID, data.Expanded.App.OAuthSecret)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = h.InitOAuther()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	connectURL, err := h.startOAuth2Connect(
		data.Context.ActingUserID,
		apps.Call{
			Wish: apps.NewWish(h.AppURL(PathWishConnectedInstall)),
			Data: data,
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type: apps.ResponseTypeOK,
			Markdown: md.Markdownf(
				"**Hallo სამყარო** needs to continue its installation using your system administrator's credentials. Please [connect](%s) the application to your Mattermost account.",
				connectURL),
		})
	return http.StatusOK, nil
}

func (h *helloapp) handleConnectedInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallData) (int, error) {
	var teams []*model.Team
	err := h.asUser(data.Context.ActingUserID,
		func(client *model.Client4) error {
			teams, _ = client.GetAllTeams("", 0, 100)
			return nil
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.ResponseTypeOK,
			Markdown: md.Markdownf("<><> OK: found %v teams", len(teams)),
		})
	return http.StatusOK, nil
}
