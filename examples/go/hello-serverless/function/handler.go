package function

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const DEPLOY_TYPE = "DEPLOY_TYPE"

// manifestData is preloaded with the Mattermost App manifest.
//go:embed static/manifest.json
var manifestData []byte

// StaticFS is preloaded with the contents of the ./static directory.
//go:embed static
var staticFS embed.FS

// DeployType is used to set, and then display how the app's instance is
// actually running (deployed as).
var DeployType apps.DeployType

// Init sets up the app's HTTP server, which is exactly the same for all of the
// deploy types.
//
// The app itself is very simple, registers a single /-command to send a DM back
// to the user. The DM includes the current DeployType of the app.
func init() {
	DeployType = apps.DeployType(os.Getenv(DEPLOY_TYPE))

	// Serve the manifest and the static assets, except in AWS Lambda, where
	// they are always served from S3.
	if DeployType != apps.DeployAWSLambda {
		http.HandleFunc("/manifest.json", httputils.DoHandleJSONData(manifestData))
		http.Handle("/static/", http.StripPrefix("/", http.FileServer(http.FS(staticFS))))
	}

	// Serve app's Calls. "/ping" is used to confirm successful deployment of an
	// App, specifically on AWS but we always make it available. Returns "PONG".
	http.HandleFunc("/ping", httputils.DoHandleJSON(
		apps.NewTextResponse("PONG")))

	// Returns the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", httputils.DoHandleJSON(
		apps.NewDataResponse(Bindings)))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send", send)
}

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
							Icon:  "icon.png",
							Fields: []apps.Field{
								{
									Type: apps.FieldTypeText,
									Name: "message",
								},
							},
							Submit: apps.NewCall("/send").WithExpand(apps.Expand{
								ActingUser: apps.ExpandID.Required(),
							}),
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
	asBot.DM(creq.Context.ActingUser.Id, message)

	httputils.WriteJSON(w,
		apps.NewTextResponse("Created a post in your DM channel."))
}

// Handle is the main entry point for OpenFAAS
func Handle(w http.ResponseWriter, req *http.Request) {
	http.DefaultServeMux.ServeHTTP(w, req)
}
