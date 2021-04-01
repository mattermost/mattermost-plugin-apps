package main

import (
	"context"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

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

//go:embed configure_form.json
var configureFormData []byte

const configKey = "config"

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
	http.HandleFunc("/oauth2/complete", oauth2Complete)

	// Submit handlers

	// `configure` command - sets up Google OAuth client credentials.
	http.HandleFunc("/configure/form", writeJSON(configureFormData))
	http.HandleFunc("/configure/submit", configure)

	// `connect` command - display the OAuth2 connect link.
	// <>/<> TODO: returning an empty form should be unnecessary, 404 should be
	// cached by the user agent as a {}
	http.HandleFunc("/connect/form", writeJSON(connectFormData))
	http.HandleFunc("/connect/submit", connect)

	// `send` command - send a Hello message.
	http.HandleFunc("/send/form", writeJSON(sendFormData))
	http.HandleFunc("/send/submit", send)

	http.ListenAndServe(":8080", nil)
}

func configure(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	clientID, _ := creq.Values["client_id"].(string)
	clientSecret, _ := creq.Values["client_secret"].(string)
	mmclient.AsBot(creq.Context).KVSet(configKey, "", map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
	})

	json.NewEncoder(w).Encode(apps.CallResponse{
		Markdown: "updated OAuth client credentials",
	})
}

func oauth2Config(asBot *mmclient.Client, creq *apps.CallRequest) *oauth2.Config {
	v, _ := asBot.KVGet(configKey, "")
	m, _ := v.(map[string]interface{})
	clientID, _ := m["client_id"].(string)
	clientSecret, _ := m["client_secret"].(string)

	completeURL := creq.Context.MattermostSiteURL +
		path.Join(creq.Context.AppPath, apps.PathOAuthComplete)

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  completeURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}
}

func connect(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	// <>/<> consider adding OAuth URLs to ExtendedContext, on request.
	// "/oauth2/remote/redirect" is hard-coded in the Apps proxy.
	json.NewEncoder(w).Encode(apps.CallResponse{
		Markdown: md.Markdownf("[Connect](%s%s) to Google.",
			creq.Context.MattermostSiteURL,
			path.Join(creq.Context.AppPath, apps.PathOAuthRedirect)),
	})
}

func oauth2Redirect(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	r := make([]byte, 10) // 20 hex digits
	rand.Read(r)
	random := hex.EncodeToString(r)
	state := fmt.Sprintf("%v_%s", random, creq.Context.ActingUserID)

	asBot := mmclient.AsBot(creq.Context)
	asBot.KVSet(state, "", state)

	oauthConfig := oauth2Config(asBot, &creq)
	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	json.NewEncoder(w).Encode(apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: url,
	})
}

func oauth2Complete(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	q, _ := creq.Values["q"].(string)
	values, _ := url.ParseQuery(q)
	state := values.Get("state")
	code := values.Get("code")
	userId := strings.Split(state, "_")[1]

	asBot := mmclient.AsBot(creq.Context)
	v, _ := asBot.KVGet(state, "")

	storedState, _ := v.(string)
	if storedState != state {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	asBot.KVDelete(state, "")

	oauthConfig := oauth2Config(asBot, &creq)
	token, _ := oauthConfig.Exchange(context.Background(), code)

	// TODO: token needs to be stored double-encoded 'cause KV doesn't have a get into a struct. Change KV?
	tokenData, _ := json.Marshal(token)
	asBot.KVSet("token"+userId, "", string(tokenData))

	json.NewEncoder(w).Encode(apps.CallResponse{})
}

func send(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	asBot := mmclient.AsBot(creq.Context)
	v, _ := asBot.KVGet("token"+creq.Context.ActingUserID, "")

	tokenData, _ := v.(string)
	var token oauth2.Token
	_ = json.Unmarshal([]byte(tokenData), &token)

	oauthConfig := oauth2Config(asBot, &creq)
	ctx := context.Background()
	tokenSource := oauthConfig.TokenSource(ctx, &token)

	oauth2Service, _ := oauth2api.NewService(ctx, option.WithTokenSource(tokenSource))
	uiService := oauth2api.NewUserinfoService(oauth2Service)
	ui, _ := uiService.V2.Me.Get().Do()
	message := fmt.Sprintf("Hello from Google, [%s](mailto:%s)!", ui.Name, ui.Email)

	calService, _ := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	cl, _ := calService.CalendarList.List().Do()
	if cl != nil && len(cl.Items) > 0 {
		message += " You have the following calendars:\n"
		for _, item := range cl.Items {
			message += "- " + item.Summary + "\n"
		}
	} else {
		message += " You have no calendars.\n"
	}

	json.NewEncoder(w).Encode(apps.CallResponse{
		Markdown: md.MD(message),
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
