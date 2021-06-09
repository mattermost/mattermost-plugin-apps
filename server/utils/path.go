package utils

import (
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
