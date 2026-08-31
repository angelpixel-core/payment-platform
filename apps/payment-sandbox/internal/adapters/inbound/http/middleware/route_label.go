package middleware

import (
	"net/http"
	"strings"
)

func routeLabel(r *http.Request) string {
	if r == nil {
		return ""
	}
	if pattern := strings.TrimSpace(r.Pattern); pattern != "" {
		return pattern
	}
	if r.URL != nil {
		return r.URL.Path
	}
	return ""
}
