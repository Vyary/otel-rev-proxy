package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

func Cors(next http.Handler) http.Handler {
	AllowedOrigins := []string{"*"}

	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	if originsEnv != "" {
		if err := json.Unmarshal([]byte(originsEnv), &AllowedOrigins); err != nil {
			slog.Error("Failed to Unmarshal ALLOWED_ORIGINS env")
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if isOriginAllowed(origin, AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}
