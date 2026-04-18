package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	csrfTokenLen   = 32
)

type csrfContextKey struct{}

// NewCSRF creates a CSRF middleware with the given secure cookie flag.
// Set secureCookie to true in production (HTTPS) and false for local dev (HTTP).
func NewCSRF(secureCookie bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Ensure CSRF cookie exists
			cookie, err := r.Cookie(csrfCookieName)
			if err != nil || cookie.Value == "" {
				token := generateCSRFToken()
				http.SetCookie(w, &http.Cookie{
					Name:     csrfCookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: false, // JS needs to read this for HTMX
					Secure:   secureCookie,
					SameSite: http.SameSiteStrictMode,
					MaxAge:   86400, // 24 hours
				})
				cookie = &http.Cookie{Value: token}
			}

			// Only validate on state-changing methods
			if r.Method == http.MethodPost || r.Method == http.MethodPut ||
				r.Method == http.MethodPatch || r.Method == http.MethodDelete {

				// Only check the header — reading r.FormValue would consume the request body
				// before the handler can parse it. All forms must send the token via the
				// X-CSRF-Token header (HTMX does this via htmx:configRequest in app.js).
				headerToken := r.Header.Get(csrfHeaderName)

				if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) != 1 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"error":{"code":"FORBIDDEN","message":"Invalid CSRF token"}}`))
					return
				}
			}

			// Store token in context so GetCSRFToken works even on first visit
			// (before the cookie is sent back to the client).
			ctx := context.WithValue(r.Context(), csrfContextKey{}, cookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateCSRFToken() string {
	b := make([]byte, csrfTokenLen)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand.Read failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// GetCSRFToken extracts the CSRF token from the request context (set by CSRF middleware).
// Falls back to the request cookie if the middleware hasn't run.
func GetCSRFToken(r *http.Request) string {
	if token, ok := r.Context().Value(csrfContextKey{}).(string); ok {
		return token
	}
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
