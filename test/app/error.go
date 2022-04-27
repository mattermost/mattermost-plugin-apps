package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func initHTTPError(r *mux.Router) {
	handleCall(r, ErrorDefault, handleError("Error"))
	handleCall(r, ErrorEmpty, handleError(""))
	handleCall(r, ErrorMarkdownForm, handleErrorMarkdownForm)
	handleCall(r, ErrorMarkdownFormMissingField, handleErrorMarkdownFormMissingField)
	r.HandleFunc(Error404, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "TEST ERROR 404 ignored", http.StatusNotFound)
	})
	r.HandleFunc(Error500, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "TEST ERROR 500", http.StatusInternalServerError)
	})

	r.HandleFunc(InvalidUnknownType, httputils.DoHandleJSON(apps.CallResponse{
		Type: "unknown",
	}))

	r.HandleFunc(InvalidHTML, httputils.DoHandleData("text/html", []byte(`
<!DOCTYPE html>
<html>
	<head>
	</head>
	<body>
		<p>HTML example</p>
	</body>
</html>
`)))
}
