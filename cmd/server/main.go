package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"website-admin/internal/ev"
	"website-admin/internal/helmet"
	"website-admin/internal/jot"
)

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

// cleanQueryMiddleware creates a middleware that strips any query parameters
// not explicitly provided in the allowedKeys list, issuing a 301 redirect if necessary.
func cleanQueryMiddleware(allowedKeys ...string) func(http.Handler) http.Handler {
	// Create a map of allowed keys for O(1) lookups
	allowed := make(map[string]bool)
	for _, k := range allowedKeys {
		allowed[k] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Fast path: if there are no query parameters, move on immediately
			if r.URL.RawQuery == "" {
				next.ServeHTTP(w, r)
				return
			}

			q := r.URL.Query()
			needsRedirect := false

			// Loop through all provided parameters
			for key := range q {
				// If the key isn't in our inclusion list, delete it
				if !allowed[key] {
					q.Del(key)
					needsRedirect = true
				}
			}

			// If we removed anything, issue a redirect to the cleaned up URL
			if needsRedirect {
				redirectPath := r.URL.Path
				cleanQuery := q.Encode()

				// Reattach the query string only if there are still valid params left
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

func spaHandler(staticPath string, indexPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticPath, r.URL.Path)

		info, err := os.Stat(path)

		// If the file does not exist, serve the custom 404 HTML page.
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(custom404HTML))
			return
		} else if jot.Log(err, "failed to stat file") { // <-- Automatically logs if err exists!
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If the path is a directory (like the root path "/"),
		// serve the index.html fallback.
		if info.IsDir() {
			http.ServeFile(w, r, filepath.Join(staticPath, indexPath))
			return
		}

		// The file exists and is not a directory. Serve it directly.
		http.ServeFile(w, r, path)
	}
}

func main() {
	port, err := ev.Get("PORT")
	jot.Fatal(err, "port not set")

	staticDir := "frontend/dist"
	spa := spaHandler(staticDir, "index.html")
	mux := http.NewServeMux()
	mux.Handle("/", spa)

	// Create the query cleaner middleware with your inclusion list.
	// Currently empty, so it strips ALL query parameters.
	queryCleaner := cleanQueryMiddleware()

	// Wrap the mux: Cache -> Query Cleaner -> Helmet
	secureApp := helmet.New("/assets/")
	app := secureApp(queryCleaner(mux))

	// --- h2c Implementation ---
	h2s := &http2.Server{}
	server := &http.Server{
		Addr: ":" + port,
		// Wrap the fully constructed middleware chain 'app' in h2c
		Handler: h2c.NewHandler(app, h2s),
	}

	jot.Info("Starting h2c server on port: " + port)

	err = server.ListenAndServe()
	jot.Fatal(err, "could not start server")
}
