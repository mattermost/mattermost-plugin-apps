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

func CleanURLPath(got string) (unescaped string, err error) {
	if got == "" {
		return "", NewInvalidError("empty path")
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
	cleanPath, err := CleanPath(unescaped)
	if err != nil {
		return "", err
	}

	return cleanPath, nil
}

func CleanStaticPath(got string) (unescaped string, err error) {
	cleanPath, err := CleanURLPath(got)
	if err != nil {
		return "", err
	}
	if cleanPath[0] == '/' {
		return "", NewInvalidError("asset names may not start with a '/'")
	}
	return cleanPath, nil
}

func CleanURL(got string) (string, error) {
	u, err := url.Parse(got)
	if err != nil {
		return "", err
	}
	u.Path, err = CleanURLPath(u.Path)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
