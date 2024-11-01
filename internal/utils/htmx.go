package utils

import "net/http"

func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func HTMXRedirect(w http.ResponseWriter, path string) {
	w.Header().Add("Hx-Redirect", path)
}
