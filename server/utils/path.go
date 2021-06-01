package utils

import (
	"net/url"
	"path"
	"strings"
)

func CleanStaticPath(got string) (unescaped string, err error) {
	if got == "" {
		return "", NewInvalidError("asset name is not specified")
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
		return "", NewInvalidError("asset names may not start with a '/'")
	}

	assetName := path.Clean(unescaped)
	if assetName == "." || strings.HasPrefix(assetName, "../") {
		return "", NewInvalidError("bad path: %s", got)
	}
	return assetName, nil
}
