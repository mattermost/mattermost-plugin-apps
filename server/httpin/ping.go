package httpin

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type VersionInfo struct {
	Version string `json:"version"`
}

func (s *Service) Ping(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	info := s.Config.Get().GetPluginVersionInfo()
	_ = httputils.WriteJSON(w, info)
}
