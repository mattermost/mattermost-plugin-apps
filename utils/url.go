package utils

import (
	"net/url"

	"github.com/pkg/errors"
)

func IsValidHTTPURL(rawUrl string) error {
	u, err := url.ParseRequestURI(rawUrl)
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
