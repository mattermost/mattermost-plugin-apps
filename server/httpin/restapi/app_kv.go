package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const (
	// MaxKVStoreValueLength is the maximum length in bytes that a value in the KV store of an app can contain
	MaxKVStoreValueLength = 8192
)

func (a *restapi) initKV(h *httpin.Handler) {
	h.HandleFunc(path.KV+"/{prefix}/{key}",
		a.KVGet, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodGet)
	h.HandleFunc(path.KV+"/{key}",
		a.KVGet, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodGet)

	h.HandleFunc(path.KV+"/{prefix}/{key}",
		a.KVPut, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodPut, http.MethodPost)
	h.HandleFunc(path.KV+"/{key}",
		a.KVPut, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodPut, http.MethodPost)

	h.HandleFunc(path.KV+"/{prefix}/{key}",
		a.KVDelete, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodDelete)
	h.HandleFunc(path.KV+"/{key}",
		a.KVDelete, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodDelete)
}

// KVGet returns a value stored by the App in the KV store.
//   Path: /api/v1/kv/[{prefix}/]{key}
//   Method: GET
//   Input: none
//   Output: a JSON object
func (a *restapi) KVGet(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]
	data, err := a.appServices.KVGet(req, req.AppID(), req.ActingUserID(), prefix, id)
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
func (a *restapi) KVPut(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]

	data, err := httputils.LimitReadAll(r.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	changed, err := a.appServices.KVSet(req, req.AppID(), req.ActingUserID(), prefix, id, data)
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
func (a *restapi) KVDelete(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]

	err := a.appServices.KVDelete(req, req.AppID(), req.ActingUserID(), prefix, id)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
