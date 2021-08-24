package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// Starts the default http server on localhost used in all "hello" examples. It
// prefers the port specified in the manifest if there is one, or 8080. If those
// fail, it starts on a random port.
func Run(manifestData []byte) {
	m := apps.Manifest{}
	err := json.Unmarshal(manifestData, &m)
	if err != nil {
		panic(err)
	}

	port := 0
	if m.HTTP != nil {
		u, _ := url.Parse(m.HTTP.RootURL)
		if u != nil {
			addr, _ := net.ResolveTCPAddr("tcp", u.Host)
			if addr != nil {
				port = addr.Port
			}
		}
	}
	if port == 0 {
		port = 8080
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		fmt.Printf("Port %v is unavailable, using a random port: %v\n", port, err)
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}
	}

	port = listener.Addr().(*net.TCPAddr).Port
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
