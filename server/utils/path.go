package utils

import (
	"path"
	"strings"
)

func CleanPath(p string) (string, error) {
	if p == "" {
		return "", NewInvalidError("invalid path: %q", p)
	}

	cleanPath := path.Clean(p)
	if p == "." || strings.HasPrefix(cleanPath, "../") {
		return "", NewInvalidError("bad path: %q", p)
	}

	return cleanPath, nil
}
