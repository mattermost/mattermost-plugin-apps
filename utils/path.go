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

func cleanURLPath(got string) (unescaped string, err error) {
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

func CleanStaticURL(got string) (unescaped string, err error) {
	u, err := url.Parse(got)
	if err != nil {
		return "", err
	}
	u.Path, err = cleanURLPath(u.Path)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" && u.Path[0] == '/' {
		u.Path = "." + u.Path
	}
	return u.String(), nil
}

func CleanURL(got string) (string, error) {
	u, err := url.Parse(got)
	if err != nil {
		return "", err
	}
	u.Path, err = cleanURLPath(u.Path)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
