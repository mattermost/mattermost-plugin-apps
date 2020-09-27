package helloapp

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (h *helloapp) handleInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallData) (int, error) {
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
	err := h.asUser(data.Context.ActingUserID,
		func(client *model.Client4) error {
			teams, api4Resp := client.GetAllTeams("", 0, 100)
			fmt.Printf("<><> RESPONSE: %+v\n", api4Resp)
			fmt.Printf("<><> TEAMS: %+v\n", teams)
			return nil
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.ResponseTypeOK,
			Markdown: "<><> OK",
		})
	return http.StatusOK, nil
}
