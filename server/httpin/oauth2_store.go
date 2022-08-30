package httpin

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (s *Service) OAuth2StoreApp(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	data, err := httputils.LimitReadAll(req.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteErrorIfNeeded(w, err)
		return
	}
	err = s.AppServices.StoreOAuth2App(r, data)
	if err != nil {
		httputils.WriteErrorIfNeeded(w, err)
		return
	}
}

func (s *Service) OAuth2StoreUser(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	data, err := httputils.LimitReadAll(req.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteErrorIfNeeded(w, err)
		return
	}
	err = s.AppServices.StoreOAuth2User(r, data)
	if err != nil {
		httputils.WriteErrorIfNeeded(w, err)
		return
	}
}

func (s *Service) OAuth2GetUser(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	data, err := s.AppServices.GetOAuth2User(r)
	if err != nil {
		httputils.WriteErrorIfNeeded(w, err)
		return
	}
	_, _ = w.Write(data)
}
