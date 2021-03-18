// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package httputils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"
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

func WriteError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
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

func WriteInternalServerError(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusInternalServerError, err)
}

func WriteBadRequestError(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusBadRequest, err)
}

func WriteNotFoundError(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusNotFound, err)
}

func WriteUnauthorizedError(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusUnauthorized, err)
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
	return ioutil.ReadAll(&io.LimitedReader{R: in, N: limit})
}

func ProcessResponseError(w http.ResponseWriter, resp *http.Response, err error) bool {
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return true
	}
	if resp.StatusCode != http.StatusOK {
		bb, _ := ReadAndClose(resp.Body)
		WriteError(w, http.StatusBadGateway,
			errors.Errorf("received status %v: %s", resp.Status, string(bb)))
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
	return ioutil.ReadAll(resp.Body)
}
