package utils

import "net/http"

func IsUserScoped(r *http.Request) bool {
	return r.URL.Query().Get("userScoped") == "on"
}
