package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var PreferredPorts = map[apps.AppID]int{
	"hello-world":     8080,
	"hello-lifecycle": 8081,
	"hello-oauth2":    8082,
	"hello-webhooks":  8083,
}

// Starts the default http server used in all "hello" examples. It prefers
// localhost://8080, but if that fails, it starts on a random port.
func Run(manifestData []byte) {
	m := apps.Manifest{}
	err := json.Unmarshal(manifestData, &m)
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", PreferredPorts[m.AppID]))
	if err != nil {
		fmt.Printf("Port 8080 is unavailable, going to use a random port: %v\n", err)
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}
	}

	port := listener.Addr().(*net.TCPAddr).Port
	if m.HTTP == nil {
		m.HTTP = &apps.HTTP{}
	}
	m.HTTP.RootURL = fmt.Sprintf("http://localhost:%v", port)

	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", func(w http.ResponseWriter, req *http.Request) {
		err = json.NewEncoder(w).Encode(m)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to serve manifest.json: "+err.Error())
		}
	})

	fmt.Printf("App %s started. Manifest at %s\n", m.AppID, m.HTTP.RootURL+"/manifest.json")
	panic(http.Serve(listener, nil))
}
