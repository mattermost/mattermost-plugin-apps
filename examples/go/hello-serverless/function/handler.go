package function

import (
	"net/http"

	_ "github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/hello"
)

// Handle is the main entry point for OpenFAAS
func Handle(w http.ResponseWriter, req *http.Request) {
	http.DefaultServeMux.ServeHTTP(w, req)
}
