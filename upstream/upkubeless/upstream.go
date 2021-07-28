// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upkubeless

import (
	"bytes"
	"fmt"
	"io"
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
	clientset := kubelessutil.GetClientOutOfCluster()

	// Build the JSON request
	sreq, err := upstream.ServerlessRequestFromCall(creq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert call into invocation payload")
	}
	fmt.Printf("<>/<> ServerlessRequest:%s\n", string(sreq))

	req := clientset.CoreV1().RESTClient().Post().Body(bytes.NewBuffer(sreq))
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("event-type", "application/json")

	// Get the function's service URL
	svc, err := clientset.CoreV1().Services(namespace(app.AppID)).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find the service for %s", funcName)
	}
	port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	if svc.Spec.Ports[0].Name != "" {
		port = svc.Spec.Ports[0].Name
	}

	serviceURL := svc.ObjectMeta.SelfLink + ":" + port + "/proxy/" +
		strings.TrimPrefix(creq.Path, "/")
	// REST package removes trailing slash when building URLs
	// Causing POST requests to be redirected with an empty body
	// So we need to manually build the URL
	req = req.AbsPath(serviceURL)

	// Set event metadata
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
		if strings.Contains(err.Error(), "status code 408") {
			// Give a more meaninful error for timeout errors
			return nil, errors.Wrap(err, "request timeout exceeded")
		}
		return nil, errors.New(strings.Replace(err.Error(), `\n`, "\n", -1))
	}
	fmt.Printf("<>/<> ServerlessResponse:%s\n", string(received))
	resp, err := upstream.ServerlessResponseFromJSON(received)
	if err != nil {
		return nil, err
	}
	return []byte(resp.Body), nil
}

func (u *Upstream) GetStatic(_ *apps.Manifest, path string) (io.ReadCloser, int, error) {
	return nil, 0, errors.New("not implemented")
}

func match(callPath string, m *apps.Manifest) string {
	matchedName := ""
	matchedPath := ""
	for _, f := range m.KubelessFunctions {
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
