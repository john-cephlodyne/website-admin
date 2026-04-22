package helmet

import (
	"net/http"
	"strings"
)

// New creates a Helmet middleware. It accepts an optional list of URL prefixes
// (like "/assets/") that should be aggressively cached. Everything else gets
// strict no-cache headers.
func New(cachePrefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()

			// --- Granular Content Security Policy (CSP) ---
			h.Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; form-action 'self'; object-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests;")

			// --- Advanced Isolation Headers ---
			h.Set("Cross-Origin-Opener-Policy", "same-origin")
			h.Set("Cross-Origin-Embedder-Policy", "require-corp")
			h.Set("Cross-Origin-Resource-Policy", "same-origin")

			// --- HTTP Strict Transport Security (HSTS) ---
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

			// --- Basic Security Headers ---
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")

			// --- Remove Information-Leaking Headers ---
			h.Del("X-Powered-By")
			h.Del("Server")

			// --- Smart Caching Logic ---
			shouldCache := false
			for _, prefix := range cachePrefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					shouldCache = true
					break
				}
			}

			if shouldCache {
				// Apply aggressive 1-year caching for static assets
				h.Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				// Apply strict no-caching for API routes and dynamic HTML
				h.Set("Cache-Control", "no-store, no-cache, must-revalidate")
				h.Set("Pragma", "no-cache")
				h.Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}
