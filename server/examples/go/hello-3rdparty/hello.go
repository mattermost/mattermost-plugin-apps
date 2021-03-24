package main

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"path"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

//go:embed icon.png
var iconData []byte

//go:embed manifest.json
var manifestData []byte

//go:embed bindings.json
var bindingsData []byte

//go:embed send_form.json
var sendFormData []byte

//go:embed connect_form.json
var connectFormData []byte

func main() {
	// Static handlers

	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", writeJSON(manifestData))

	// Serve the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", writeJSON(bindingsData))

	// Serve the icon for the App.
	http.HandleFunc("/static/icon.png", writeData("image/png", iconData))

	// Remote OAuth2 handlers

	// Handle an OAuth2 connect request redirect.
	http.HandleFunc("/oauth2/redirect", oauth2Redirect)

	// Handle a successful OAuth2 connection.
	http.HandleFunc("/oauth2/success", oauth2Success)

	// Submit handlers

	// `send` command - send a Hello message.
	http.HandleFunc("/send/form", writeJSON(sendFormData))
	http.HandleFunc("/send/submit", send)

	// `connect` command - display the OAuth2 connect link.
	// <>/<> TODO: returning an empty form should be unnecessary, 404 should be
	// cached by the user agent as a {}
	http.HandleFunc("/connect/form", writeJSON(connectFormData))
	http.HandleFunc("/connect/submit", connect)

	http.ListenAndServe(":8080", nil)
}

func send(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	message := "Hello, world!"
	v, ok := c.Values["message"]
	if ok && v != nil {
		message += fmt.Sprintf(" ...and %s!", v)
	}
	mmclient.AsBot(c.Context).DM(c.Context.ActingUserID, message)

	json.NewEncoder(w).Encode(apps.CallResponse{})
}

func connect(w http.ResponseWriter, req *http.Request) {
	call := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&call)

	// <>/<> consider adding OAuth URLs to ExtendedContext, on request.
	// "/oauth2/remote/redirect" is hard-coded in the Apps proxy.
	txt := fmt.Sprintf("[Connect](%s%s) to Google.", call.Context.MattermostSiteURL, path.Join(call.Context.AppPath, apps.PathOAuthRedirect))

	mmclient.AsBot(call.Context).DM(call.Context.ActingUserID, txt)
	json.NewEncoder(w).Encode(apps.CallResponse{
		Markdown: md.MD(txt),
	})
}

func oauth2Redirect(w http.ResponseWriter, req *http.Request) {
	call := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&call)

	r := make([]byte, 10) // 20 hex digits
	rand.Read(r)
	random := hex.EncodeToString(r)
	state := fmt.Sprintf("%v_%s", random, call.Context.ActingUserID)

	asBot := mmclient.AsBot(call.Context)
	asBot.KVSet(state, "", state)

	json.NewEncoder(w).Encode(apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: "https://www.google.com",
	})
}

func oauth2Success(w http.ResponseWriter, req *http.Request) {
	fmt.Println("<>/<> OAuth2 success")
	json.NewEncoder(w).Encode(apps.CallResponse{})
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
