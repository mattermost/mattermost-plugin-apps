package httputils

import (
	"net/url"

	"github.com/pkg/errors"
)

// IsValidURL checks if a given URL is a valid URL with a host and a http or http scheme.
func IsValidURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.Errorf("URL schema must either be %q or %q", "http", "https")
	}

	if u.Host == "" {
		return errors.New("URL must contain a host")
	}

	return nil
}
