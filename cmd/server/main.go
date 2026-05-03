package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"website-admin/internal/api"
	"website-admin/internal/ev"
	"website-admin/internal/helmet"
	"website-admin/internal/jot"
)

// Restored to your exact original version, no inline styles!
const custom404HTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>404 - Page Not Found</title>
    <link rel="stylesheet" href="/server.css">
</head>
<body>
    <div class="container">
        <h1 class="error-code">404</h1>
        <p class="message">The path you requested could not be found.</p>
        <a href="https://cephlodyne.com" class="btn">Return to Cephlodyne</a>
    </div>
</body>
</html>`

func cleanQueryMiddleware(allowedKeys ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, k := range allowedKeys {
		allowed[k] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.RawQuery == "" {
				next.ServeHTTP(w, r)
				return
			}

			q := r.URL.Query()
			needsRedirect := false

			for key := range q {
				if !allowed[key] {
					q.Del(key)
					needsRedirect = true
				}
			}

			if needsRedirect {
				redirectPath := r.URL.Path
				cleanQuery := q.Encode()
				if cleanQuery != "" {
					redirectPath += "?" + cleanQuery
				}
				http.Redirect(w, r, redirectPath, http.StatusMovedPermanently)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		next.ServeHTTP(w, r)
	})
}

// FIXED: Now properly handles client-side routing
func spaHandler(staticPath string, indexPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticPath, filepath.Clean(r.URL.Path))

		info, err := os.Stat(path)

		if os.IsNotExist(err) {
			// If the path has no extension (e.g., "/dashboard", "/users"),
			// it is a Svelte route. Serve index.html and let Svelte handle it!
			if filepath.Ext(path) == "" {
				http.ServeFile(w, r, filepath.Join(staticPath, indexPath))
				return
			}

			// If it HAS an extension (e.g., "/logo.png", "/server.css"),
			// it's a missing file. Serve the custom 404 HTML.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(custom404HTML))
			return
		} else if err != nil {
			if jot.Log(err, "failed to stat file") {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if info.IsDir() {
			http.ServeFile(w, r, filepath.Join(staticPath, indexPath))
			return
		}

		http.ServeFile(w, r, path)
	}
}

func main() {
	port, err := ev.Get("PORT")
	jot.Fatal(err, "port not set")

	mux := http.NewServeMux()

	// Register the Connect-RPC API Routes FIRST
	api.Register(mux)

	staticDir := "frontend/dist"
	spa := spaHandler(staticDir, "index.html")
	mux.Handle("/", spa)

	queryCleaner := cleanQueryMiddleware()
	secureApp := helmet.New("/assets/")
	app := secureApp(queryCleaner(mux))

	h2s := &http2.Server{}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: h2c.NewHandler(app, h2s),
	}

	jot.Info("Starting h2c server on port: " + port)

	err = server.ListenAndServe()
	jot.Fatal(err, "could not start server")
}
