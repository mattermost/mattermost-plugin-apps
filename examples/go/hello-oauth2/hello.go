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
	rootURL    = "http://localhost:8082"
	listenAddr = ":8082"
)

//go:embed icon.png
var iconData []byte

var manifest = apps.Manifest{
	AppID:       "hello-oauth2",
	Version:     "0.8.0",
	DisplayName: "Hello, OAuth2!",
	Icon:        "icon.png",
	HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/hello-oauth2",
	RequestedPermissions: []apps.Permission{
		apps.PermissionActAsUser,
		apps.PermissionRemoteOAuth2,
	},
	RequestedLocations: []apps.Location{
		apps.LocationCommand,
	},
	Bindings: apps.NewCall("/bindings").WithExpand(apps.Expand{
		ActingUser: apps.ExpandAll,
		OAuth2User: apps.ExpandAll,
	}),
	Deploy: apps.Deploy{
		HTTP: &apps.HTTP{
			RootURL: rootURL,
		},
	},
}

func main() {
	// Static handlers

	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", httputils.DoHandleJSON(manifest))

	// Serve the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", bindings)

	// Serve the icon for the App.
	http.HandleFunc("/static/icon.png",
		httputils.DoHandleData("image/png", iconData))

	// Google OAuth2 handlers

	// Handle an OAuth2 connect URL request.
	http.HandleFunc("/oauth2/connect", oauth2Connect)
	// Handle a successful OAuth2 connection.
	http.HandleFunc("/oauth2/complete", oauth2Complete)

	// Command submit handlers

	// `configure` command - sets up Google OAuth client credentials.
	http.HandleFunc("/configure", configure)
	// `connect` command - display the OAuth2 connect link.
	http.HandleFunc("/connect", connect)
	// `disconnect` command - disconnect your account.
	http.HandleFunc("/disconnect", disconnect)
	// `send` command - send a Hello message.
	http.HandleFunc("/send", send)

	fmt.Printf("hello-oauth2 app listening on %q \n", listenAddr)
	fmt.Printf("Install via /apps install http %s/manifest.json \n", rootURL)
	panic(http.ListenAndServe(listenAddr, nil))
}

func bindings(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	commandBinding := apps.Binding{
		Icon:        "icon.png",
		Label:       "hello-oauth2",
		Description: "Hello remote (3rd party) OAuth2 App",
		Bindings:    []apps.Binding{},
	}

	token := oauth2.Token{}
	remarshal(&token, creq.Context.OAuth2.User)

	if token.AccessToken == "" {
		commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
			Location: "connect",
			Label:    "connect",
			Submit: apps.NewCall("/connect").WithExpand(apps.Expand{
				OAuth2App: apps.ExpandAll,
			}),
		})
	} else {
		commandBinding.Bindings = append(commandBinding.Bindings,
			apps.Binding{
				Location: "send",
				Label:    "send",
				Submit: apps.NewCall("/send").WithExpand(apps.Expand{
					OAuth2App:  apps.ExpandAll,
					OAuth2User: apps.ExpandAll,
				}),
			},
			apps.Binding{
				Location: "disconnect",
				Label:    "disconnect",
				Submit: apps.NewCall("/disconnect").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
				}),
			},
		)
	}

	if creq.Context.ActingUser.IsSystemAdmin() {
		configure := apps.Binding{
			Location: "configure",
			Label:    "configure",
			Form: &apps.Form{
				Title: "Configures Google OAuth2 App credentials",
				Icon:  "icon.png",
				Fields: []apps.Field{
					{
						Type:       "text",
						Name:       "client_id",
						Label:      "client-id",
						IsRequired: true,
					},
					{
						Type:       "text",
						Name:       "client_secret",
						Label:      "client-secret",
						IsRequired: true,
					},
				},
				Submit: apps.NewCall("/configure").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
				}),
			},
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
	asUser.StoreOAuth2App(creq.Context.AppID, apps.OAuth2App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

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
		Text: "Disconnected your Google account",
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
