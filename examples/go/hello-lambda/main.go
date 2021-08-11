package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

//go:embed manifest-http.json
var manifestHTTPData []byte

//go:embed manifest.json
var manifestAWSData []byte

//go:embed pong.json
var pongData []byte

//go:embed bindings.json
var bindingsData []byte

//go:embed send_form.json
var formData []byte

const (
	host = "localhost"
	port = 8080
)

//go:embed static
var static embed.FS

func main() {
	localMode := os.Getenv("LOCAL") == "true"

	// Serve its own manifest as HTTP for convenience in dev. mode.

	manifestData := manifestAWSData
	if localMode {
		manifestData = manifestHTTPData
	}
	http.HandleFunc("/manifest.json", writeJSON(manifestData))

	// Returns "PONG". Used for `appsctl test aws`.
	http.HandleFunc("/ping", writeJSON(pongData))

	// Returns the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", writeJSON(bindingsData))

	// The form for sending a Hello message.
	http.HandleFunc("/send/form", writeJSON(formData))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send/submit", send)

	if localMode {
		addr := fmt.Sprintf("%v:%v", host, port)
		fmt.Printf(`hello-world app listening at http://%s`, addr)
		http.ListenAndServe(":8080", nil)
	} else {
		lambda.Start(httpadapter.New(http.DefaultServeMux).Proxy)
	}
}

func send(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	message := "Hello from AWS Lambda!"
	v, ok := c.Values["message"]
	if ok && v != nil {
		message += fmt.Sprintf(" ...and %s!", v)
	}
	mmclient.AsBot(c.Context).DM(c.Context.ActingUserID, message)

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
