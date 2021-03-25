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

func Hey() string {
	return "hey"
}

type testapp struct {
	conf config.Service
}

const TestAppPath = "/e2e-testapp"
const runE2ETestApp = true

var bindingsRequest []byte
var bindingsResponse []byte

var channelHeaderSubmitRequest []byte
var channelHeaderSubmitResponse []byte

var formSubmitRequest []byte
var formSubmitResponse []byte

func Init(router *mux.Router, mm *pluginapi.Client, conf config.Service, _ proxy.Service, _ appservices.Service) {
	if !runE2ETestApp {
		return
	}

	r := router.PathPrefix(TestAppPath).Subrouter()
	r.HandleFunc("/bindings", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		bindingsRequest = b

		writeJSON(bindingsResponse)
	})

	r.HandleFunc("/bindings/set-response", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		bindingsResponse = b
		w.Write([]byte("Success"))
	})

	r.HandleFunc("/bindings/get-request", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(bindingsRequest)
	})

	r.HandleFunc("/test-form/set-submit-response", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		formSubmitResponse = b
		w.Write([]byte("Success"))
	})

	r.HandleFunc("/test-form/submit", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		formSubmitRequest = b
		writeJSON(formSubmitResponse)
	})

	r.HandleFunc("/test-form/get-submit-request", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(formSubmitRequest)
	})

	r.HandleFunc("/channel-header-submit/set-submit-response", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		channelHeaderSubmitResponse = b
		w.Write([]byte("Success"))
	})

	r.HandleFunc("/channel-header-submit/submit", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		channelHeaderSubmitRequest = b
		writeJSON(channelHeaderSubmitResponse)
	})

	r.HandleFunc("/channel-header-submit/get-submit-request", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(channelHeaderSubmitRequest)
	})

	r.HandleFunc("/manifest.json", writeJSON(manifest))
}
