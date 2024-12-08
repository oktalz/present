package handlers

import (
	"net/http"

	configuration "github.com/oktalz/present/config"
)

func Homepage(config configuration.Config) http.Handler {
	if config.AspectRatio.DisableAspectRatio {
		return NoLayout(config)
	}
	return IFrame(config)
}
