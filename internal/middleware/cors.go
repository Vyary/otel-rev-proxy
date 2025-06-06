package middleware

import (
	"net/http"
	"strings"
)

func Cors(next http.Handler) http.Handler {
	const (
		allowedMethods   = "GET, POST, OPTIONS"
		allowedHeaders   = "Content-Type, Authorization"
		allowCredentials = "true"
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if strings.TrimSpace(origin) != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
