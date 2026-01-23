package security

import (
	"net/http"

	"github.com/vimaurya/lumen/internal/config"
)

func PasswordProtection(cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || user != "admin" || pass != "very" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Lumen Admin"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}
