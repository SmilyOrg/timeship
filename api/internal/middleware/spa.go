package middleware

import (
	"net/http"
	"os"
	"path"
	"strings"
)

// SpaFs is a filesystem wrapper that serves index.html for SPA routing
type SpaFs struct {
	Root http.FileSystem
}

func (fs SpaFs) Open(name string) (http.File, error) {
	f, err := fs.Root.Open(name)
	if os.IsNotExist(err) && !strings.HasSuffix(name, ".br") && !strings.HasSuffix(name, ".gz") {
		return fs.Root.Open("index.html")
	}
	return f, err
}

// IndexHTML middleware appends index.html to directory requests
func IndexHTML() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/") || len(r.URL.Path) == 0 {
				r.URL.Path = path.Join(r.URL.Path, "index.html")
			}
			next.ServeHTTP(w, r)
		})
	}
}
