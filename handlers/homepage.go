package handlers

import (
	"net/http"

	configuration "github.com/oktalz/present/config"
)

func Homepage(iframeHandler http.Handler, config configuration.Config) http.Handler {
	if config.AspectRatio.Disable {
		return NoLayout(config)
	}
	return iframeHandler
}
