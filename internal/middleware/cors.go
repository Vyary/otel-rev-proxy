package middleware

import (
	"net/http"
)

func Cors(next http.Handler) http.Handler {
	const (
		allowedMethods    = "GET, POST, PUT, DELETE, OPTIONS"
		allowedHeaders    = "Content-Type, Authorization"
		allowCredentials  = "true"
		allowOriginHeader = "*"
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		}

		// Handle preflight requests (OPTIONS method)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
