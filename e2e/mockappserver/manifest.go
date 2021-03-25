package mockappserver

var manifest = []byte(`{
	"app_id": "e2e-testapp",
	"display_name": "E2E Test App",
	"app_type": "http",
	"root_url": "https://mickmister.ngrok.io/plugins/com.mattermost.apps/e2e-testapp",
	"homepage_url": "https://github.com/mattermost/mattermost-plugin-apps",
	"requested_permissions": [
		"act_as_bot"
	],
	"requested_locations": [
		"/channel_header",
		"/command",
		"/post_menu"
	]
}`)
