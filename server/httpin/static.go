package httpin

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (s *Service) Static(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	if len(vars) == 0 {
		httputils.WriteErrorIfNeeded(w, utils.NewInvalidError("invalid URL format"))
		return
	}
	assetName, err := utils.CleanStaticPath(vars["name"])
	if err != nil {
		httputils.WriteErrorIfNeeded(w, err)
		return
	}

	body, status, err := s.Proxy.InvokeGetStatic(r, assetName)
	if err != nil {
		r.Log.WithError(err).Debugw("failed to get asset", "asset_name", assetName)
		httputils.WriteErrorIfNeeded(w, err)
		return
	}

	copyHeader(w.Header(), req.Header)
	w.WriteHeader(status)
	if _, err := io.Copy(w, body); err != nil {
		httputils.WriteErrorIfNeeded(w, err)
		return
	}
	if err := body.Close(); err != nil {
		httputils.WriteErrorIfNeeded(w, err)
		return
	}
}

func copyHeader(dst, src http.Header) {
	headerKey := "Content-Type"
	dst.Add(headerKey, src.Get(headerKey))
}
