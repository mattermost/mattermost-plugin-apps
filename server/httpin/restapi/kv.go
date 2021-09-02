package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const (
	// MaxKVStoreValueLength is the maximum length in bytes that a value in the KV store of an app can contain
	MaxKVStoreValueLength = 8192
)

func (a *restapi) kvGet(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]
	var out interface{}
	err := a.appServices.KVGet(in.ActingUserID, prefix, id, &out)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, out)
}

func (a *restapi) kvPut(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]

	data, err := httputils.LimitReadAll(r.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	// <>/<> TODO: atomic support
	// <>/<> TODO: TTL support

	changed, err := a.appServices.KVSet(in.ActingUserID, prefix, id, data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, map[string]interface{}{
		"changed": changed,
	})
}

func (a *restapi) kvDelete(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]

	err := a.appServices.KVDelete(in.ActingUserID, prefix, id)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
