// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upkubeless

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
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

const Namespace = "mattermost-kubeless-apps"

type Upstream struct {
	kubelessClient versioned.Interface
}

var _ upstream.Upstream = (*Upstream)(nil)

func MakeUpstream() (*Upstream, error) {
	kubelessClient, err := kubelessutil.GetKubelessClientOutCluster()
	if os.IsNotExist(err) {
		return nil, errors.Wrap(utils.ErrNotFound, err.Error())
	}
	if err != nil {
		return nil, err
	}
	return &Upstream{
		kubelessClient: kubelessClient,
	}, nil
}

func (u *Upstream) Roundtrip(app apps.App, creq apps.CallRequest, async bool) (io.ReadCloser, error) {
	clientset := kubelessutil.GetClientOutOfCluster()

	url, err := resolvePath(clientset, &app.Manifest, creq.Path)
	if err != nil {
		return nil, err
	}

	// Build the JSON request
	creqData, err := upstream.ServerlessRequestFromCall(creq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert call into invocation payload")
	}

	crespData, err := u.invoke(clientset, url, http.MethodPost, creqData, async)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(crespData)), nil
}

// resolvePath resolved a call path into a fully-qualified URL.
func resolvePath(clientset kubernetes.Interface, m *apps.Manifest, path string) (string, error) {
	funcName := match(m, path)
	if funcName == "" {
		return "", utils.ErrNotFound
	}

	// Get the function's service URL
	svc, err := clientset.CoreV1().Services(Namespace).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "failed to find the kubernetes service for function %s", funcName)
	}
	port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	if svc.Spec.Ports[0].Name != "" {
		port = svc.Spec.Ports[0].Name
	}

	fURL := svc.ObjectMeta.SelfLink + ":" + port + "/proxy/" + strings.TrimPrefix(path, "/")
	return fURL, nil
}

func (u *Upstream) invoke(clientset kubernetes.Interface, url, method string, data []byte, async bool) ([]byte, error) {
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
	req.SetHeader("event-namespace", Namespace)

	// REST package removes trailing slash when building URLs
	// Causing POST requests to be redirected with an empty body
	// So we need to manually build the URL
	req = req.AbsPath(url)

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

func (u *Upstream) GetStatic(_ apps.App, path string) (io.ReadCloser, int, error) {
	return nil, 0, errors.New("not implemented")
}

func match(m *apps.Manifest, callPath string) string {
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
