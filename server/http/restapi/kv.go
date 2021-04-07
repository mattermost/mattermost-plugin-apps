package restapi

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) kvGet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]
	var out interface{}
	err := a.appServices.KVGet(actingID(r), prefix, id, &out)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, out)
}

func (a *restapi) kvPut(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]

	data, err := io.ReadAll(r.Body)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	// <>/<> TODO: atomic support
	// <>/<> TODO: TTL support

	changed, err := a.appServices.KVSet(actingID(r), prefix, id, data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, map[string]interface{}{
		"changed": changed,
	})
}

func (a *restapi) kvDelete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["key"]
	prefix := mux.Vars(r)["prefix"]

	err := a.appServices.KVDelete(actingID(r), prefix, id)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
