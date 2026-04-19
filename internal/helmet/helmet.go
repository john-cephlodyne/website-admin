// Package helmet provides a simple middleware to set secure HTTP headers by default.
// It's designed to be a secure starting point for any Go web application.
package helmet

import (
	"net/http"
)

// Helmet is a middleware that applies a comprehensive and restrictive security header baseline.
// It follows a secure-by-default approach, and individual headers can be overridden
// in subsequent handlers if necessary.
func Helmet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		// --- Granular Content Security Policy (CSP) ---
		// Prevents XSS, data injection, and clickjacking by restricting content sources.
		h.Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; form-action 'self'; object-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests;")

		// --- Advanced Isolation Headers (Enables Cross-Origin Isolation) ---
		// Mitigates side-channel attacks like Spectre. Can break third-party embeds.
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Embedder-Policy", "require-corp")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")

		// --- HTTP Strict Transport Security (HSTS) ---
		// Enforces HTTPS connections. Test thoroughly before enabling 'preload'.
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// --- Basic Security Headers ---
		h.Set("X-Content-Type-Options", "nosniff")                                                  // Prevents MIME-sniffing.
		h.Set("X-Frame-Options", "DENY")                                                            // Legacy clickjacking protection.
		h.Set("Referrer-Policy", "no-referrer")                                                     // Enhances user privacy.
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()") // Disables invasive browser features.

		// --- Caching Headers ---
		// Prevents caching of potentially sensitive information.
		h.Set("Cache-Control", "no-store, no-cache, must-revalidate")
		h.Set("Pragma", "no-cache") // For legacy HTTP/1.0
		h.Set("Expires", "0")       // For legacy proxies

		// --- Remove Information-Leaking Headers ---
		h.Del("X-Powered-By")
		h.Del("Server")

		next.ServeHTTP(w, r)
	})
}
