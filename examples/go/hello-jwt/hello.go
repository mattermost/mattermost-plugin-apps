package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/pkg/errors"
)

var secret = []byte("1234")

//go:embed icon.png
var iconData []byte

var manifest = apps.Manifest{
	AppID:       "hello-jwt",
	Version:     "v0.8.0",
	DisplayName: "Hello, JWT!",
	Icon:        "icon.png",
	HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/hello-jwt",
	RequestedPermissions: []apps.Permission{
		apps.PermissionActAsBot,
	},
	RequestedLocations: []apps.Location{
		apps.LocationCommand,
	},
	Deploy: apps.Deploy{
		HTTP: &apps.HTTP{
			RootURL: "http://localhost:8084",
			UseJWT:  true,
		},
	},
}

var bindings = []apps.Binding{
	{
		Location: "/command",
		Bindings: []apps.Binding{
			{
				Icon:        "icon.png",
				Label:       "hello-jwt",
				Description: "Hello JWT app",
				Hint:        "[send]",
				Bindings: []apps.Binding{
					{
						Location: "send",
						Label:    "send",
						Form: &apps.Form{
							Submit: &apps.Call{
								Path: "/send",
							},
						},
					},
				},
			},
		},
	},
}

func main() {
	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", httputils.HandleStaticJSON(manifest))

	// Ping to test the JWT connectivity upon install.
	http.HandleFunc("/ping", withJWT(
		httputils.HandleStaticJSON(apps.NewDataResponse(nil))))

	// Returns the Channel Header and Command bindings for the app.
	http.HandleFunc("/bindings", withJWT(
		httputils.HandleStaticJSON(apps.NewDataResponse(bindings))))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send/submit", withJWT(
		send))

	// Serves the icon for the app.
	http.HandleFunc("/static/icon.png",
		httputils.HandleStaticData("image/png", iconData))

	addr := ":8084" // matches manifest.json
	fmt.Println("Listening on", addr)
	fmt.Println("Use '/apps install http http://localhost" + addr + "/manifest.json' to install the app")
	fmt.Printf("Use %q as the app's JWT secret\n", secret)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func send(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	appclient.AsBot(c.Context).DM(c.Context.ActingUserID, "JWT check ok")

	httputils.WriteJSON(w,
		apps.NewTextResponse("Created a post in your DM channel."))
}

func withJWT(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if _, err := checkJWT(req); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		f(w, req)
	}
}

func checkJWT(req *http.Request) (*apps.JWTClaims, error) {
	authValue := req.Header.Get(apps.OutgoingAuthHeader)
	if !strings.HasPrefix(authValue, "Bearer ") {
		return nil, errors.Errorf("missing %s: Bearer header", apps.OutgoingAuthHeader)
	}

	jwtoken := strings.TrimPrefix(authValue, "Bearer ")
	claims := apps.JWTClaims{}
	_, err := jwt.ParseWithClaims(jwtoken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
