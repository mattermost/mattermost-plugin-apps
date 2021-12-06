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

	http.Error(w, err.Error(), ErrorToStatus(err))
}

func ErrorToStatus(err error) int {
	switch errors.Cause(err) {
	case utils.ErrForbidden:
		return http.StatusForbidden
	case utils.ErrUnauthorized:
		return http.StatusUnauthorized
	case utils.ErrNotFound:
		return http.StatusNotFound
	case utils.ErrInvalid:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
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

// DoHandleJSONData returns an http.HandleFunc that serves a JSON-encoded data
// chunk.
func DoHandleJSONData(data []byte) http.HandlerFunc {
	return DoHandleData("application/json", data)
}

// DoHandleData returns an http.HandleFunc that serves a data chunk with a
// specified content-type.
func DoHandleData(ct string, data []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", ct)
		_, _ = w.Write(data)
	}
}

// DoHandleJSON returns an http.HandleFunc that serves a data chunk with a
// specified content-type.
func DoHandleJSON(v interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		_ = WriteJSON(w, v)
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
