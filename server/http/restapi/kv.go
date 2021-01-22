package restapi

import (
	// nolint:gosec

	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

// TODO use raw byte API: for now all JSON is re-encoded to use api.Mattermost API

func (a *restapi) kvList(w http.ResponseWriter, r *http.Request, botUserID, prefix string) {
	// <><>TODO kvList
}

func (a *restapi) kvGet(w http.ResponseWriter, r *http.Request, botUserID, prefix string) {
	id := mux.Vars(r)["key"]
	out := map[string]interface{}{}

	client := a.getClient(botUserID, w, r)
	out, err := client.KVGet(prefix, id)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
	httputils.WriteJSON(w, out)
}

func (a *restapi) kvHead(w http.ResponseWriter, r *http.Request, botUserID, prefix string) {
	// TODO "HEAD"
}

func (a *restapi) kvPut(w http.ResponseWriter, r *http.Request, botUserID, prefix string) {
	id := mux.Vars(r)["key"]

	in := map[string]interface{}{}

	// TODO size limit
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	// <><>TODO atomic support
	// <><>TODO TTL support

	client := a.getClient(botUserID, w, r)
	changed, err := client.KVSet(prefix, id, in)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
	httputils.WriteJSON(w, map[string]interface{}{
		"changed": changed,
	})
}

func (a *restapi) kvDelete(w http.ResponseWriter, r *http.Request, botUserID, prefix string) {
	id := mux.Vars(r)["key"]

	client := a.getClient(botUserID, w, r)
	out, err := client.KVDelete(prefix, id)
	if err != nil || out == false {
		httputils.WriteInternalServerError(w, err)
		return
	}
}

func (a *restapi) handleKV(
	f func(http.ResponseWriter, *http.Request, string, string),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		botUserID := r.Header.Get("Mattermost-User-Id")
		if botUserID == "" {
			httputils.WriteUnauthorizedError(w, errors.New("not authorized"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			httputils.WriteBadRequestError(w, err)
			return
		}
		prefix := r.Form.Get("prefix")

		f(w, r, botUserID, prefix)
	}
}