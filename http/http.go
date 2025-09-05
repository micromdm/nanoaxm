package http

import "net/http"

// DelHeaderMiddleware deletes header from the HTTP request headers before calling h.
func DelHeaderMiddleware(h http.Handler, header string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(header)
		h.ServeHTTP(w, r)
	}
}
