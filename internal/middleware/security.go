package middleware

import "net/http"

// SecurityHeaders adds common security-related headers to every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent page from being displayed in a frame (clickjacking protection)
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent browser from MIME-sniffing a response away from the declared content-type
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable the XSS filter in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Enforce HTTPS by telling the browser to only connect to the site over HTTPS for the next year.
		// Note: Only enable this if you have HTTPS set up and are ready to commit to it.
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}
