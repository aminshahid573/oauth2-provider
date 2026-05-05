package middleware

import "net/http"

// SecurityHeaders adds common security-related headers to every response.
// In production, HSTS is enabled per RFC 6797 to prevent SSL-stripping attacks.
// Referrer-Policy and a restrictive Content-Security-Policy are always set
// to protect an authorization server from clickjacking, XSS, and data leakage.
func SecurityHeaders(appEnv string) func(http.Handler) http.Handler {
	isProduction := appEnv == "production"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent page from being displayed in a frame (clickjacking protection)
			w.Header().Set("X-Frame-Options", "DENY")

			// Prevent browser from MIME-sniffing a response away from the declared content-type
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Enable the XSS filter in older browsers
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Limit referrer information to prevent leaking OAuth parameters (codes, tokens)
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Restrictive CSP suitable for an authorization server that serves its own HTML pages.
			// Allows inline styles for template rendering but blocks all external scripts and frames.
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'; form-action 'self'; base-uri 'self'")

			// HSTS: enforce HTTPS for 2 years with subdomain inclusion (RFC 6797).
			// Only enabled in production to avoid locking out local dev environments.
			if isProduction {
				w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}
