package utils

import (
	"net/url"
	"path"
	"strings"
)

func CleanPath(p string) (string, error) {
	if p == "" {
		return "", NewInvalidError("path must not be empty: %s", p)
	}

	cleanPath := path.Clean(p)
	if cleanPath == "." || strings.HasPrefix(cleanPath, "../") {
		return "", NewInvalidError("bad path: %q", p)
	}

	return cleanPath, nil
}

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

	cleanPath, err := CleanPath(unescaped)
	if err != nil {
		return "", err
	}

	return cleanPath, nil
}
