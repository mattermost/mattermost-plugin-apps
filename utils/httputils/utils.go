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

	"github.com/mattermost/mattermost-plugin-apps/utils"
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

// WriteJSONStatus encodes and writes out an object, with a custom response
// status code.
func WriteJSONStatus(w http.ResponseWriter, statusCode int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(v)
}

// WriteJSON encodes and writes out an object, with a 200 response status code.
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	return WriteJSONStatus(w, http.StatusOK, v)
}

// HandleJSON returns an http.HandleFunc that serves a JSON-encoded data
// chunk of an object.
func HandleJSON(v interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data, err := json.Marshal(v)
		if err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}
}

// HandleJSONData returns an http.HandleFunc that serves a JSON-encoded data
// chunk.
func HandleJSONData(data []byte) http.HandlerFunc {
	return HandleData("application/json", data)
}

// HandleData returns an http.HandleFunc that serves a data chunk with a
// specified content-type.
func HandleData(ct string, data []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", ct)
		_, _ = w.Write(data)
	}
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
