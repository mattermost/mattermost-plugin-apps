package function

import (
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

//go:embed data/manifest.json
var manifestData []byte

//go:embed data/pong.json
var pongData []byte

//go:embed data/bindings.json
var bindingsData []byte

//go:embed data/send_form.json
var formData []byte

//go:embed static
var static embed.FS

// Handler for OpenFaaS and faasd.
func Handle(w http.ResponseWriter, r *http.Request) {
	InitApp(apps.DeployOpenFAAS)
	http.DefaultServeMux.ServeHTTP(w, r)
}

var deployType apps.DeployType

func InitApp(dt apps.DeployType) {
	// Serve static assets.
	http.Handle("/static/", http.FileServer(http.FS(static)))

	// Returns the manifest for the App.
	http.HandleFunc("/manifest.json", writeJSON(manifestData))

	// Serve app's Calls. "/ping" is used in `appsctl test aws`
	// Returns "PONG". Used for `appsctl test aws`.
	http.HandleFunc("/ping", writeJSON(pongData))

	// Returns the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", writeJSON(bindingsData))

	// The form for sending a Hello message.
	http.HandleFunc("/send/form", writeJSON(formData))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send/submit", send)

	deployType = dt
}

func send(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	message := fmt.Sprintf("Hello from a serververless app running as %s!", deployType)
	v, ok := creq.Values["message"]
	if ok && v != nil {
		message += fmt.Sprintf(" ...and %s!", v)
	}

	// Running on ngrok in development, need this to avoid getting "x509:
	// certificate signed by unknown authority" error when running in a fresh
	// ubuntu container.
	asBot := mmclient.AsBot(creq.Context)
	asBot.HttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	asBot.DM(creq.Context.ActingUserID, message)

	json.NewEncoder(w).Encode(apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: "Created a post in your DM channel.",
	})
}

func writeData(ct string, data []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", ct)
		w.Write(data)
	}
}

func writeJSON(data []byte) func(w http.ResponseWriter, r *http.Request) {
	return writeData("application/json", data)
}
