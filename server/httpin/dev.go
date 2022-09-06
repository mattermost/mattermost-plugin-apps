package httpin

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (s *Service) RefreshBindings(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	if !*s.Config.MattermostConfig().Config().ServiceSettings.EnableDeveloper {
		http.Error(w, "Development route unreachable due to disabled EnableDeveloper config setting", http.StatusBadRequest)
		return
	}

	userID := req.URL.Query().Get("user_id")
	if err := s.AppServices.RefreshBindings(r, userID); err != nil {
		http.Error(w, err.Error(), httputils.ErrorToStatus(err))
		return
	}
}
