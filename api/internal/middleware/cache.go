package middleware

import (
	"net/http"
	"regexp"
)

var staticCacheRegex = regexp.MustCompile(`.+\.\w`)

// CacheControl middleware sets cache headers for static assets
func CacheControl() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if staticCacheRegex.MatchString(r.URL.Path) {
				w.Header().Set("Cache-Control", "max-age=31536000")
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
