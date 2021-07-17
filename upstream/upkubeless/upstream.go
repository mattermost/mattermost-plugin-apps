// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upkubeless

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kubelessutil "github.com/kubeless/kubeless/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Upstream struct {
	kubelessClient versioned.Interface
}

type Logger interface {
	Error(message string, keyValuePairs ...interface{})
	Warn(message string, keyValuePairs ...interface{})
	Info(message string, keyValuePairs ...interface{})
	Debug(message string, keyValuePairs ...interface{})
}

var _ upstream.Upstream = (*Upstream)(nil)

type invocationPayload struct {
	Path       string            `json:"path"`
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type invocationResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func X() (versioned.Interface, error) {
	return kubelessutil.GetKubelessClientOutCluster()
}

func MakeUpstream() (*Upstream, error) {
	kubelessClient, err := kubelessutil.GetKubelessClientOutCluster()
	if err != nil {
		return nil, err
	}
	return &Upstream{
		kubelessClient: kubelessClient,
	}, nil
}

func (u *Upstream) Roundtrip(app *apps.App, creq *apps.CallRequest, async bool) (io.ReadCloser, error) {
	name := match(creq.Path, &app.Manifest)
	if name == "" {
		return nil, utils.ErrNotFound
	}

	data, err := u.InvokeFunction(app, name, creq, async)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

// InvokeFunction is a public method used in appsctl, but is not a part of the
// upstream.Upstream interface. It invokes a function with a specified name,
// with no conversion.
func (u *Upstream) InvokeFunction(app *apps.App, funcName string, creq *apps.CallRequest, async bool) ([]byte, error) {
	payload, err := callToInvocationPayload(creq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert call into invocation payload")
	}

	clientset := kubelessutil.GetClientOutOfCluster()
	svc, err := clientset.CoreV1().Services(namespace(app.AppID)).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find the service for %s", funcName)
	}

	port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	if svc.Spec.Ports[0].Name != "" {
		port = svc.Spec.Ports[0].Name
	}

	req := clientset.CoreV1().RESTClient().Post().Body(bytes.NewBuffer(payload))
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("event-type", "application/json")
	// REST package removes trailing slash when building URLs
	// Causing POST requests to be redirected with an empty body
	// So we need to manually build the URL
	req = req.AbsPath(svc.ObjectMeta.SelfLink + ":" + port + "/proxy/")
	timestamp := time.Now().UTC()
	eventID, err := kubelessutil.GetRandString(11)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate ID")
	}
	req.SetHeader("event-id", eventID)
	req.SetHeader("event-time", timestamp.Format(time.RFC3339))
	req.SetHeader("event-namespace", "mattermost")
	received, err := req.Do().Raw()
	if err != nil {
		// Properly interpret line breaks
		// logrus.Error(string(res))
		if strings.Contains(err.Error(), "status code 408") {
			// Give a more meaninful error for timeout errors
			return nil, errors.Wrap(err, "request timeout exceeded")
		} else {
			return nil, errors.New(strings.Replace(err.Error(), `\n`, "\n", -1))
		}
	}
	return received, nil
}

func (u *Upstream) GetStatic(app *apps.App, path string) (io.ReadCloser, int, error) {
	return nil, 0, errors.New("not implemented")
}

func callToInvocationPayload(cr *apps.CallRequest) ([]byte, error) {
	body, err := json.Marshal(cr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal call for lambda payload")
	}

	request := invocationPayload{
		Path:       cr.Path,
		HTTPMethod: http.MethodPost,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal lambda payload")
	}
	return payload, nil
}

func match(callPath string, m *apps.Manifest) string {
	matchedName := ""
	matchedPath := ""
	for _, f := range m.KubelessFunctions {
		if strings.HasPrefix(callPath, f.CallPath) {
			if len(f.CallPath) > len(matchedPath) {
				matchedName = FunctionName(m.AppID, m.Version, f.Name)
				matchedPath = f.CallPath
			}
		}
	}

	return matchedName
}

func FunctionName(appID apps.AppID, version apps.AppVersion, function string) string {
	// Sanitized any dots used in appID and version as lambda function names can not contain dots
	// While there are other non-valid characters, a dots is the most commonly used one
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")
	sanitizedFunction := strings.ReplaceAll(function, " ", "-")
	return fmt.Sprintf("%s_%s_%s", sanitizedAppID, sanitizedVersion, sanitizedFunction)
}

func namespace(appID apps.AppID) string {
	return "mattermost_app_" + string(appID)
}
