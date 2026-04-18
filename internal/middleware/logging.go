package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// Logger is a structured request logging middleware with PII sanitization.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		reqID := chimw.GetReqID(r.Context())

		log.Info().
			Str("request_id", reqID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.Status()).
			Int("bytes", ww.BytesWritten()).
			Dur("duration", duration).
			Str("ip", sanitizeIP(r.RemoteAddr)).
			Msg("request")
	})
}

// sanitizeIP masks the last octet/segment of an IP for privacy.
// Handles both IPv4 (1.2.3.***) and IPv6 ([2001:db8:***]).
func sanitizeIP(addr string) string {
	// Strip port using net.SplitHostPort (handles both IPv4 and IPv6)
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr // no port present
	}

	// IPv4
	parts := strings.Split(host, ".")
	if len(parts) == 4 {
		parts[3] = "***"
		return strings.Join(parts, ".")
	}

	// IPv6: mask last segment
	segments := strings.Split(host, ":")
	if len(segments) > 1 {
		segments[len(segments)-1] = "***"
		return strings.Join(segments, ":")
	}

	return host
}
