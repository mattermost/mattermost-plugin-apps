package kubeless

import (
	_ "embed"

	"github.com/kubeless/kubeless/pkg/functions"
	"github.com/sirupsen/logrus"
)

//go:embed manifest-http.json
var manifestHTTPData []byte

//go:embed manifest.json
var manifestAWSData []byte

//go:embed pong.json
var pongData []byte

//go:embed bindings.json
var bindingsData []byte

//go:embed send_form.json
var formData []byte

const (
	host = "localhost"
	port = 8080
)

func Handler(event functions.Event, context functions.Context) (string, error) {
	logrus.Error(event.Data)
	return event.Data, nil
}
