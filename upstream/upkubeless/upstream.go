// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upkubeless

import (
	"bytes"
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
	"k8s.io/client-go/kubernetes"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Upstream struct {
	kubelessClient versioned.Interface
}

var _ upstream.Upstream = (*Upstream)(nil)

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
	if app.Manifest.Kubeless == nil {
		return nil, errors.New("no 'kubeless' section in manifest.json")
	}
	name := match(creq.Path, &app.Manifest)
	if name == "" {
		return nil, utils.ErrNotFound
	}

	// Build the JSON request
	creqData, err := upstream.ServerlessRequestFromCall(creq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert call into invocation payload")
	}

	crespData, err := u.invoke(app, name, creq.Path, http.MethodPost, creqData, async)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(crespData)), nil
}

func FunctionURL(appID apps.AppID, funcName string) (string, error) {
	clientset := kubelessutil.GetClientOutOfCluster()
	return functionURL(clientset, appID, funcName)
}

func functionURL(clientset kubernetes.Interface, appID apps.AppID, funcName string) (string, error) {
	// Get the function's service URL
	svc, err := clientset.CoreV1().Services(namespace(appID)).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "failed to find the kubernetes service for function %s", funcName)
	}
	port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	if svc.Spec.Ports[0].Name != "" {
		port = svc.Spec.Ports[0].Name
	}

	serviceURL := svc.ObjectMeta.SelfLink + ":" + port + "/proxy/"
	return serviceURL, nil
}

func (u *Upstream) invoke(app *apps.App, funcName string, requestPath, method string, data []byte, async bool) ([]byte, error) {
	clientset := kubelessutil.GetClientOutOfCluster()
	fURL, err := functionURL(clientset, app.AppID, funcName)
	if err != nil {
		return nil, err
	}
	fURL += strings.TrimPrefix(requestPath, "/")

	timestamp := time.Now().UTC()
	eventID, err := kubelessutil.GetRandString(11)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate ID")
	}

	req := clientset.CoreV1().RESTClient().Post().Body(bytes.NewBuffer(data))

	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("event-type", "application/json")
	req.SetHeader("event-id", eventID)
	req.SetHeader("event-time", timestamp.Format(time.RFC3339))
	req.SetHeader("event-namespace", "mattermost")

	// REST package removes trailing slash when building URLs
	// Causing POST requests to be redirected with an empty body
	// So we need to manually build the URL
	req = req.AbsPath(fURL)

	received, err := req.Do().Raw()
	if err != nil {
		// Properly interpret line breaks
		if strings.Contains(err.Error(), "status code 408") {
			// Give a more meaninful error for timeout errors
			return nil, errors.Wrap(err, "request timeout exceeded")
		}
		return nil, errors.New(strings.ReplaceAll(err.Error(), `\n`, "\n"))
	}
	resp, err := upstream.ServerlessResponseFromJSON(received)
	if err != nil {
		return nil, err
	}
	return []byte(resp.Body), nil
}

func (u *Upstream) GetStatic(_ *apps.App, path string) (io.ReadCloser, int, error) {
	return nil, 0, errors.New("not implemented")
}

func match(callPath string, m *apps.Manifest) string {
	matchedName := ""
	matchedPath := ""
	for _, f := range m.Kubeless.Functions {
		if strings.HasPrefix(callPath, f.CallPath) {
			if len(f.CallPath) > len(matchedPath) {
				matchedName = FunctionName(m.AppID, m.Version, f.Handler)
				matchedPath = f.CallPath
			}
		}
	}

	return matchedName
}

func FunctionName(appID apps.AppID, version apps.AppVersion, function string) string {
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")
	sanitizedFunction := strings.ReplaceAll(function, " ", "-")
	sanitizedFunction = strings.ReplaceAll(sanitizedFunction, "_", "-")
	sanitizedFunction = strings.ReplaceAll(sanitizedFunction, ".", "-")
	sanitizedFunction = strings.ToLower(sanitizedFunction)
	return fmt.Sprintf("%s-%s-%s", sanitizedAppID, sanitizedVersion, sanitizedFunction)
}

func namespace(appID apps.AppID) string {
	sanitized := string(appID)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-")
	return "mattermost-app-" + sanitized
}
