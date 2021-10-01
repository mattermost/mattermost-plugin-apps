package function

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

//go:embed pong.json
var pongData []byte

//go:embed bindings.json
var bindingsData []byte

//go:embed send_form.json
var formData []byte

// Handler for OpenFaaS and faasd.
func Handle(w http.ResponseWriter, r *http.Request) {
	InitApp(apps.DeployOpenFAAS)
	http.DefaultServeMux.ServeHTTP(w, r)
}

var deployType apps.DeployType

func InitApp(dt apps.DeployType) {
	// Serve app's Calls. "/ping" is used in `appsctl test aws`
	// Returns "PONG". Used for `appsctl test aws`.
	http.HandleFunc("/ping", httputils.HandleJSONData(pongData))

	// Returns the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", httputils.HandleJSONData(bindingsData))

	// The form for sending a Hello message.
	http.HandleFunc("/send/form", httputils.HandleJSONData(formData))

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

	asBot := appclient.AsBot(creq.Context)
	asBot.DM(creq.Context.ActingUserID, message)

	httputils.WriteJSON(w,
		apps.NewTextResponse("Created a post in your DM channel."))
}
