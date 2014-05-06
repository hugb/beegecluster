package utils

import (
	"net/http"
	"strings"
)

func GetHostFromQueryParam(r *http.Request) string {
	if r == nil {
		return ""
	}
	if err := r.ParseForm(); err != nil && !strings.HasPrefix(err.Error(), "mime:") {
		return ""
	} else {
		return r.Form.Get("host")
	}
}
