package restapi

import (
	// nolint:gosec

	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

// TODO use raw byte API: for now all JSON is re-encoded to use apps.Mattermost API

func (a *restapi) kvList(w http.ResponseWriter, req *http.Request, botUserID, prefix string) {
	// <><>TODO kvList
}

func (a *restapi) kvGet(w http.ResponseWriter, req *http.Request, botUserID, prefix string) {
	id := mux.Vars(req)["key"]
	out := map[string]interface{}{}
	err := a.apps.API.KVGet(botUserID, prefix, id, out)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
	httputils.WriteJSON(w, out)
}

func (a *restapi) kvHead(w http.ResponseWriter, req *http.Request, botUserID, prefix string) {
	// TODO "HEAD"
}

func (a *restapi) kvPut(w http.ResponseWriter, req *http.Request, botUserID, prefix string) {
	id := mux.Vars(req)["key"]

	in := map[string]interface{}{}
	// TODO size limit
	err := json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	// <><>TODO atomic support
	// <><>TODO TTL support
	changed, err := a.apps.API.KVSet(botUserID, prefix, id, in)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
	httputils.WriteJSON(w, map[string]interface{}{
		"changed": changed,
	})
}

func (a *restapi) kvDelete(w http.ResponseWriter, req *http.Request, botUserID, prefix string) {
	id := mux.Vars(req)["key"]
	err := a.apps.API.KVDelete(botUserID, prefix, id)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
}

func (a *restapi) handleKV(
	f func(http.ResponseWriter, *http.Request, string, string),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		botUserID := req.Header.Get("Mattermost-User-Id")
		if botUserID == "" {
			httputils.WriteUnauthorizedError(w, errors.New("not authorized"))
			return
		}

		err := req.ParseForm()
		if err != nil {
			httputils.WriteBadRequestError(w, err)
			return
		}
		prefix := req.Form.Get("prefix")

		f(w, req, botUserID, prefix)
	}
}
