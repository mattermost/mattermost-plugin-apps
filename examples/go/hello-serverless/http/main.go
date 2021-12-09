package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/function"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func main() {
	mpath := flag.String("manifest", "", "path to the manifest file to serve on /manifest.json")
	spath := flag.String("static", "", "path to the static folder to serve on /static")
	flag.Parse()

	if mpath != nil && *mpath != "" {
		mdata, err := os.ReadFile(*mpath)
		if err != nil {
			panic(err)
		}
		http.HandleFunc("/manifest.json", httputils.DoHandleJSONData(mdata))
	}

	if spath != nil && *spath != "" {
		http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(*spath))))
	}

	function.DeployType = apps.DeployHTTP
	fmt.Println("Listening on :8080")
	panic(http.ListenAndServe(":8080", nil))
}
