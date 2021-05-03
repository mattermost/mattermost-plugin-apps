package gateway

import (
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) static(w http.ResponseWriter, req *http.Request, _, _ string) {
	appID := appIDVar(req)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	vars := mux.Vars(req)
	if len(vars) == 0 {
		httputils.WriteError(w, utils.NewInvalidError("invalid URL format"))
		return
	}
	assetName, err := cleanStaticPath(vars["name"])
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	// TODO verify that request is from the correct app

	body, status, err := g.proxy.GetAsset(appID, assetName)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	copyHeader(w.Header(), req.Header)
	w.WriteHeader(status)
	if _, err := io.Copy(w, body); err != nil {
		httputils.WriteError(w, err)
		return
	}
	if err := body.Close(); err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func copyHeader(dst, src http.Header) {
	headerKey := "Content-Type"
	dst.Add(headerKey, src.Get(headerKey))
}

func cleanStaticPath(got string) (unescaped string, err error) {
	if got == "" {
		return "", utils.NewInvalidError("asset name is not specified")
	}
	for escaped := got; ; escaped = unescaped {
		unescaped, err = url.PathUnescape(escaped)
		if err != nil {
			return "", err
		}
		if unescaped == escaped {
			break
		}
	}
	if unescaped[0] == '/' {
		return "", utils.NewInvalidError("asset names may not start with a '/'")
	}

	assetName := path.Clean(unescaped)
	if assetName == "." || strings.HasPrefix(assetName, "../") {
		return "", utils.NewInvalidError("bad path: %s", got)
	}
	return assetName, nil
}
