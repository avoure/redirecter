package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func AuthenticateMW(next http.Handler) http.Handler {

	var SECRET_TOKEN string = os.Getenv("SECRET_TOKEN")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(token), "bearer") {
			token = strings.TrimSpace(token[len("bearer"):])
		}

		if token != SECRET_TOKEN {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}
