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
	AppID:       "example-jwt",
	Version:     "v1.0.0",
	DisplayName: "Example app with JWT",
	Icon:        "icon.png",
	HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/jwt",
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
				Label:       "example-jwt",
				Description: "Example HTTP JWT app",
				Bindings: []apps.Binding{
					{
						Label: "test",
						Submit: &apps.Call{
							Path: "/test",
						},
					},
				},
			},
		},
	},
}

func main() {
	http.HandleFunc("/manifest.json", httputils.DoHandleJSON(manifest))

	// Ping is invoke to confirm connectivity immediately after install.
	http.HandleFunc("/ping", requireJWT(httputils.DoHandleJSON(apps.NewDataResponse(nil))))

	// Returns the bindings for the app.
	http.HandleFunc("/bindings", requireJWT(httputils.DoHandleJSON(apps.NewDataResponse(bindings))))

	// Sends a test message.
	http.HandleFunc("/test", requireJWT(test))

	// Serves the icon for the app.
	http.HandleFunc("/static/icon.png", httputils.DoHandleData("image/png", iconData))

	addr := ":8084" // matches manifest.json
	fmt.Println("Listening on", addr)
	fmt.Println("Use '/apps install http http://localhost" + addr + "/manifest.json' to install the app")
	fmt.Printf("Use %q as the app's JWT secret\n", secret)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func test(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	appclient.AsBot(c.Context).DM(c.Context.ActingUser.Id, "JWT tested ok")

	httputils.WriteJSON(w, apps.NewTextResponse("Created a post in your DM channel."))
}

func requireJWT(f http.HandlerFunc) http.HandlerFunc {
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
