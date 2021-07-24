package kubeless

import (
	"github.com/kubeless/kubeless/pkg/functions"
	"github.com/sirupsen/logrus"
)

const (
	host = "localhost"
	port = 8080
)

func Handler(event functions.Event, context functions.Context) (string, error) {
	logrus.Error(event.Data)
	return event.Data, nil
}
