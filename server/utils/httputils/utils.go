// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package httputils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func NormalizeRemoteBaseURL(mattermostSiteURL, remoteURL string) (string, error) {
	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		ss := strings.Split(u.Path, "/")
		if len(ss) > 0 && ss[0] != "" {
			u.Host = ss[0]
			u.Path = path.Join(ss[1:]...)
		}
		u, err = url.Parse(u.String())
		if err != nil {
			return "", err
		}
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid URL, no hostname: %q", remoteURL)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	remoteURL = strings.TrimSuffix(u.String(), "/")
	if remoteURL == strings.TrimSuffix(mattermostSiteURL, "/") {
		return "", fmt.Errorf("%s is the Mattermost site URL. Please use the remote application's URL", remoteURL)
	}

	return remoteURL, nil
}

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		http.Error(w, "invalid (unknown?) error", http.StatusInternalServerError)
		return
	}
	switch errors.Cause(err) {
	case utils.ErrForbidden:
		http.Error(w, err.Error(), http.StatusForbidden)
	case utils.ErrUnauthorized:
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case utils.ErrNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	case utils.ErrInvalid:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func WriteJSONStatus(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

const InLimit = 10 * (1 << 20)

func ReadAndClose(in io.ReadCloser) ([]byte, error) {
	defer in.Close()
	return LimitReadAll(in, InLimit)
}

func LimitReadAll(in io.Reader, limit int64) ([]byte, error) {
	if in == nil {
		return []byte{}, nil
	}
	return io.ReadAll(&io.LimitedReader{R: in, N: limit})
}

func ProcessResponseError(w http.ResponseWriter, resp *http.Response, err error) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	if resp.StatusCode != http.StatusOK {
		bb, _ := ReadAndClose(resp.Body)
		http.Error(w,
			fmt.Sprintf("received status %v: %s", resp.Status, string(bb)),
			http.StatusBadGateway)
		return true
	}
	return false
}

func GetFromURL(url string) ([]byte, error) {
	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func CheckAuthorized(mm *pluginapi.Client, f func(_ http.ResponseWriter, _ *http.Request, sessionID, actingUserID string)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get("Mattermost-User-Id")
		if actingUserID == "" {
			WriteError(w, utils.ErrUnauthorized)
			return
		}
		sessionID := req.Header.Get("MM_SESSION_ID")
		if sessionID == "" {
			WriteError(w, errors.Wrap(utils.ErrUnauthorized, "no user session"))
			return
		}

		f(w, req, sessionID, actingUserID)
	}
}
