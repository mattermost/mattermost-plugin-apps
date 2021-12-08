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

// DeployType is used to set, and then display how the app's instance is
// actually running (deployed as).
var DeployType apps.DeployType

// Handler is used exclusively for OpenFaaS and faasd, as the main entry-point.
// The name `Handler` appears hardcoded in the OpenFaas template used to build
// the image.
func Handle(w http.ResponseWriter, r *http.Request) {
	DeployType = apps.DeployOpenFAAS
	http.DefaultServeMux.ServeHTTP(w, r)
}

// Init sets up the app's HTTp server, which is exactly the same for all of the
// deploy types. Including this package as `_ ".../function"` is sufficient to
// initialize the app's server.
//
// The app itself is very simple, registers a single /-command to send a DM back
// to the user. The DM includes the current DeployType of the app.
func Init() {
	// Serve app's Calls. "/ping" is used in `appsctl test aws`
	// Returns "PONG". Used for `appsctl test aws`.
	http.HandleFunc("/ping", httputils.DoHandleJSONData(PongData))

	// Returns the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", httputils.DoHandleJSON(apps.NewDataResponse(Bindings)))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send", send)
}

// PongData is used to answer the "/ping" response from the app, to confirm its
// successful installation.
//
//go:embed pong.json
var PongData []byte

var Bindings = []apps.Binding{
	{
		Location: apps.LocationCommand,
		Bindings: []apps.Binding{
			{
				Icon:        "icon.png",
				Label:       "hello-serverless",
				Description: "Hello Serverless app",
				Hint:        "[send]",
				Bindings: []apps.Binding{
					{
						Label: "send",
						Form: &apps.Form{
							Title: "Hello, serverless!",
							Icon:  "/static/icon.png",
							Fields: []apps.Field{
								{
									Type: apps.FieldTypeText,
									Name: "message",
								},
							},
							Submit: apps.NewCall("/send"),
						},
					},
				},
			},
		},
	},
}

func send(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	message := fmt.Sprintf("Hello from a serverless app running as %s!", DeployType)
	v, ok := creq.Values["message"]
	if ok && v != nil {
		message += fmt.Sprintf(" ...and %s!", v)
	}

	asBot := appclient.AsBot(creq.Context)
	asBot.DM(creq.Context.ActingUserID, message)

	httputils.WriteJSON(w,
		apps.NewTextResponse("Created a post in your DM channel."))
}
