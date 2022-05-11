package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const (
	// MaxKVStoreValueLength is the maximum length in bytes that a value in the KV store of an app can contain
	MaxKVStoreValueLength = 8192
)

func (a *restapi) initKV(h *handler.Handler) {
	h.HandleFunc(path.KV+"/{prefix}/{key}",
		a.KVGet, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodGet)
	h.HandleFunc(path.KV+"/{key}",
		a.KVGet, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodGet)

	h.HandleFunc(path.KV+"/{prefix}/{key}",
		a.KVPut, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodPut, http.MethodPost)
	h.HandleFunc(path.KV+"/{key}",
		a.KVPut, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodPut, http.MethodPost)

	h.HandleFunc(path.KV+"/{prefix}/{key}",
		a.KVDelete, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodDelete)
	h.HandleFunc(path.KV+"/{key}",
		a.KVDelete, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodDelete)
}

// KVGet returns a value stored by the App in the KV store.
//   Path: /api/v1/kv/[{prefix}/]{key}
//   Method: GET
//   Input: none
//   Output: a JSON object
func (a *restapi) KVGet(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["key"]
	prefix := mux.Vars(req)["prefix"]
	data, err := a.appServices.KVGet(r, r.SourceApp().AppID, r.ActingUserID(), prefix, id)
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
func (a *restapi) KVPut(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["key"]
	prefix := mux.Vars(req)["prefix"]
	data, err := httputils.LimitReadAll(req.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	changed, err := a.appServices.KVSet(r, r.SourceApp().AppID, r.ActingUserID(), prefix, id, data)
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
func (a *restapi) KVDelete(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["key"]
	prefix := mux.Vars(req)["prefix"]

	err := a.appServices.KVDelete(r, r.SourceApp().AppID, r.ActingUserID(), prefix, id)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
