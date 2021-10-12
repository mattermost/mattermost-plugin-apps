package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const (
	host = "localhost"
	port = 8082
)

//go:embed icon.png
var iconData []byte

//go:embed manifest.json
var manifestData []byte

//go:embed send_form.json
var sendFormData []byte

//go:embed connect_form.json
var connectFormData []byte

//go:embed disconnect_form.json
var disconnectFormData []byte

//go:embed configure_form.json
var configureFormData []byte

func main() {
	// Static handlers

	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", httputils.HandleStaticJSONData(manifestData))

	// Serve the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", bindings)

	// Serve the icon for the App.
	http.HandleFunc("/static/icon.png",
		httputils.HandleStaticData("image/png", iconData))

	// Google OAuth2 handlers

	// Handle an OAuth2 connect URL request.
	http.HandleFunc("/oauth2/connect", oauth2Connect)

	// Handle a successful OAuth2 connection.
	http.HandleFunc("/oauth2/complete", oauth2Complete)

	// Submit handlers

	// `configure` command - sets up Google OAuth client credentials.
	http.HandleFunc("/configure/form", httputils.HandleStaticJSONData(configureFormData))
	http.HandleFunc("/configure/submit", configure)

	// `connect` command - display the OAuth2 connect link.
	// <>/<> TODO: returning an empty form should be unnecessary, 404 should be
	// cached by the user agent as a {}
	http.HandleFunc("/connect/form", httputils.HandleStaticJSONData(connectFormData))
	http.HandleFunc("/connect/submit", connect)

	// `disconnect` command - disconnect your account.
	http.HandleFunc("/disconnect/form", httputils.HandleStaticJSONData(disconnectFormData))
	http.HandleFunc("/disconnect/submit", disconnect)

	// `send` command - send a Hello message.
	http.HandleFunc("/send/form", httputils.HandleStaticJSONData(sendFormData))
	http.HandleFunc("/send/submit", send)

	addr := fmt.Sprintf(":%v", port)
	rootURL := fmt.Sprintf("http://%v:%v", host, port)
	fmt.Printf("hello-oauth2 app listening on %q \n", addr)
	fmt.Printf("Install via /apps install url %s/manifest.json \n", rootURL)
	panic(http.ListenAndServe(addr, nil))
}

func bindings(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	commandBinding := apps.Binding{
		Icon:        "icon.png",
		Label:       "hello-oauth2",
		Description: "Hello remote (3rd party) OAuth2 App",
		Hint:        "",
		Bindings:    []apps.Binding{},
	}

	token := oauth2.Token{}
	remarshal(&token, creq.Context.OAuth2.User)

	if token.AccessToken == "" {
		connect := apps.Binding{
			Location: "connect",
			Label:    "connect",
			Form:     apps.NewBlankForm(apps.NewCall("/connect")),
		}

		commandBinding.Bindings = append(commandBinding.Bindings, connect)
	} else {
		send := apps.Binding{
			Location: "send",
			Label:    "send",
			Form:     apps.NewBlankForm(apps.NewCall("/send")),
		}

		disconnect := apps.Binding{
			Location: "disconnect",
			Label:    "disconnect",
			Form:     apps.NewBlankForm(apps.NewCall("/disconnect")),
		}
		commandBinding.Bindings = append(commandBinding.Bindings, send, disconnect)
	}

	if creq.Context.ActingUser.IsSystemAdmin() {
		configure := apps.Binding{
			Location: "configure",
			Label:    "configure",
			Form:     apps.NewBlankForm(apps.NewCall("/configure")),
		}
		commandBinding.Bindings = append(commandBinding.Bindings, configure)
	}

	json.NewEncoder(w).Encode(apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: []apps.Binding{{
			Location: apps.LocationCommand,
			Bindings: []apps.Binding{
				commandBinding,
			},
		}},
	})
}

func configure(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)
	clientID, _ := creq.Values["client_id"].(string)
	clientSecret, _ := creq.Values["client_secret"].(string)

	asUser := appclient.AsActingUser(creq.Context)
	asUser.StoreOAuth2App(creq.Context.AppID, clientID, clientSecret)

	json.NewEncoder(w).Encode(
		apps.NewTextResponse("updated OAuth client credentials"))
}

func connect(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	json.NewEncoder(w).Encode(
		apps.NewTextResponse("[Connect](%s) to Google.", creq.Context.OAuth2.ConnectURL))
}

func disconnect(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	asActingUser := appclient.AsActingUser(creq.Context)
	err := asActingUser.StoreOAuth2User(creq.Context.AppID, nil)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(apps.CallResponse{
		Markdown: "Disconnected your Google account",
	})
}

func oauth2Config(creq *apps.CallRequest) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     creq.Context.OAuth2.ClientID,
		ClientSecret: creq.Context.OAuth2.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  creq.Context.OAuth2.CompleteURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}
}

func oauth2Connect(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)
	state, _ := creq.Values["state"].(string)

	url := oauth2Config(&creq).AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	httputils.WriteJSON(w, apps.NewDataResponse(url))
}

func oauth2Complete(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)
	code, _ := creq.Values["code"].(string)

	token, _ := oauth2Config(&creq).Exchange(context.Background(), code)

	asActingUser := appclient.AsActingUser(creq.Context)
	asActingUser.StoreOAuth2User(creq.Context.AppID, token)

	httputils.WriteJSON(w, apps.NewDataResponse(nil))
}

func send(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	oauthConfig := oauth2Config(&creq)
	token := oauth2.Token{}
	remarshal(&token, creq.Context.OAuth2.User)
	ctx := context.Background()
	tokenSource := oauthConfig.TokenSource(ctx, &token)
	oauth2Service, _ := oauth2api.NewService(ctx, option.WithTokenSource(tokenSource))
	calService, _ := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	uiService := oauth2api.NewUserinfoService(oauth2Service)

	ui, _ := uiService.V2.Me.Get().Do()
	message := fmt.Sprintf("Hello from Google, [%s](mailto:%s)!", ui.Name, ui.Email)
	cl, _ := calService.CalendarList.List().Do()
	if cl != nil && len(cl.Items) > 0 {
		message += " You have the following calendars:\n"
		for _, item := range cl.Items {
			message += "- " + item.Summary + "\n"
		}
	} else {
		message += " You have no calendars.\n"
	}

	httputils.WriteJSON(w, apps.NewTextResponse(message))

	// Store new token if refreshed
	newToken, err := tokenSource.Token()
	if err != nil && newToken.AccessToken != token.AccessToken {
		appclient.AsActingUser(creq.Context).StoreOAuth2User(creq.Context.AppID, newToken)
	}
}

func remarshal(dst, src interface{}) {
	data, _ := json.Marshal(src)
	json.Unmarshal(data, dst)
}
