package httpin

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const (
	// MaxKVStoreValueLength is the maximum length in bytes that a value in the KV store of an app can contain
	MaxKVStoreValueLength = 8192
)

// KVGet returns a value stored by the App in the KV store.
//   Path: /api/v1/kv/[{prefix}/]{key}
//   Method: GET
//   Input: none
//   Output: a JSON object
func (s *Service) KVGet(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["key"]
	prefix := mux.Vars(req)["prefix"]
	data, err := s.AppServices.KVGet(r, prefix, id)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	_, _ = w.Write(data)
}

// KVPut stores an App-provided JSON document in the KV store.
//   Path: /api/v1/kv/[{prefix}/]{key}
//   Methods: POST, PUT
//   Output: a JSON object
//   Output:
//     changed: set to true if the key value was changed.
func (s *Service) KVPut(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["key"]
	prefix := mux.Vars(req)["prefix"]
	data, err := httputils.LimitReadAll(req.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	changed, err := s.AppServices.KVSet(r, prefix, id, data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	_ = httputils.WriteJSON(w, map[string]interface{}{
		"changed": changed,
	})
}

// KVDelete removes a (App-specific) value from the KV store.
//   Path: /api/v1/kv/[{prefix}/]{key}
//   Methods: DELETE
//   Input: none
//   Output: none
func (s *Service) KVDelete(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["key"]
	prefix := mux.Vars(req)["prefix"]

	err := s.AppServices.KVDelete(r, prefix, id)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
