package utils

import (
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/mattermost/mattermost-server/v6/utils/fileutils"
)

// FindDir looks for the given directory in nearby ancestors relative to the current working
// directory as well as the directory of the executable, falling back to `./` if not found.
func FindDir(dir string) (string, bool) {
	commonBaseSearchPaths := []string{
		".",
		"..",
		"../..",
		"../../..",
		"../../../..",
	}
	found := fileutils.FindPath(dir, commonBaseSearchPaths, func(fileInfo os.FileInfo) bool {
		return fileInfo.IsDir()
	})
	if found == "" {
		return "./", false
	}

	return found, true
}

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
