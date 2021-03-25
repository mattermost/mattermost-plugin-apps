package mockappserver

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

const TestAppPath = "/e2e-testapp"
const runE2ETestApp = true

var bindingsRequest []byte
var bindingsResponse []byte

type CapturedData struct {
	Submit []byte
	Form   []byte
	Lookup []byte
}

type StoredRequestResponse struct {
	Requests  CapturedData
	Responses CapturedData
}

var storedRequestResponses = map[string]*StoredRequestResponse{}

func ensureStoredData(formName string) *StoredRequestResponse {
	if storedRequestResponses[formName] == nil {
		storedRequestResponses[formName] = &StoredRequestResponse{}
	}

	return storedRequestResponses[formName]
}

func Init(router *mux.Router, mm *pluginapi.Client, conf config.Service, _ proxy.Service, _ appservices.Service) {
	if !runE2ETestApp {
		return
	}

	r := router.PathPrefix(TestAppPath).Subrouter()
	r.HandleFunc("/manifest.json", writeJSON(manifest))

	r.HandleFunc("/bindings", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		bindingsRequest = b

		writeJSON(bindingsResponse)(w, r)
	})

	r.HandleFunc("/bindings/set-response", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		bindingsResponse = b
		w.Write([]byte("Success"))
	})

	r.HandleFunc("/bindings/get-request", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(bindingsRequest)(w, r)
	})

	r.HandleFunc("/{form_name}/submit", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		name := mux.Vars(r)["form_name"]
		stored := ensureStoredData(name)
		stored.Requests.Submit = b

		writeJSON(stored.Responses.Submit)(w, r)
	})

	r.HandleFunc("/{form_name}/set-submit-response", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		name := mux.Vars(r)["form_name"]
		stored := ensureStoredData(name)
		stored.Responses.Submit = b

		w.Write([]byte("Success"))
	})

	r.HandleFunc("/{form_name}/get-submit-request", func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["form_name"]
		stored := ensureStoredData(name)

		writeJSON(stored.Requests.Submit)(w, r)
	})

	r.HandleFunc("/clean", func(w http.ResponseWriter, r *http.Request) {
		bindingsRequest = nil
		bindingsResponse = nil
		storedRequestResponses = map[string]*StoredRequestResponse{}
	})
}
