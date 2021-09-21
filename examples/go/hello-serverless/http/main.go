package main

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"

	_ "github.com/mattermost/mattermost-example-apps/hello-serverless"
	"github.com/mattermost/mattermost-example-apps/hello-serverless/function"
)

func main() {
	function.InitApp(apps.DeployHTTP)
	fmt.Println("Listening on :8080")
	panic(http.ListenAndServe(":8080", nil))
}
