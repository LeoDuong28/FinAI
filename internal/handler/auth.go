package handler

import (
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/a-h/templ"
	"github.com/rs/zerolog/log"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/pages"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginPage renders the login page.
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Login(csrf, "")).ServeHTTP(w, r)
}

// RegisterPage renders the registration page.
func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Register(csrf, "")).ServeHTTP(w, r)
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderAuthError(w, r, "register", "Invalid form data")
		return
	}

	input := service.RegisterInput{
		Email:     strings.ToLower(strings.TrimSpace(r.FormValue("email"))),
		Password:  r.FormValue("password"),
		FirstName: strings.TrimSpace(r.FormValue("first_name")),
		LastName:  strings.TrimSpace(r.FormValue("last_name")),
	}

	// Validate password confirmation
	if r.FormValue("password") != r.FormValue("password_confirm") {
		renderAuthError(w, r, "register", "Passwords do not match")
		return
	}

	// Basic validation
	if input.Email == "" || input.FirstName == "" || input.LastName == "" {
		renderAuthError(w, r, "register", "All fields are required")
		return
	}

	// Email format validation
	if len(input.Email) > 254 {
		renderAuthError(w, r, "register", "Email address is too long")
		return
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		renderAuthError(w, r, "register", "Invalid email address")
		return
	}

	// Name length validation
	if utf8.RuneCountInString(input.FirstName) > 100 || utf8.RuneCountInString(input.LastName) > 100 {
		renderAuthError(w, r, "register", "Name must be 100 characters or fewer")
		return
	}

	_, err := h.authService.Register(r.Context(), input)
	if err != nil {
		if domainErr, ok := apperr.IsDomainError(err); ok {
			renderAuthError(w, r, "register", domainErr.Message)
			return
		}
		log.Error().Err(err).Msg("registration failed")
		renderAuthError(w, r, "register", "An unexpected error occurred")
		return
	}

	// Auto-login after registration
	tokens, err := h.authService.Login(r.Context(), service.LoginInput{
		Email:     input.Email,
		Password:  input.Password,
		UserAgent: r.UserAgent(),
		IPAddress: clientIP(r),
	})
	if err != nil {
		// Registration succeeded but auto-login failed — redirect to login
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	setAuthCookies(w, tokens)

	// HTMX redirect
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/app/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/app/dashboard", http.StatusSeeOther)
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderAuthError(w, r, "login", "Invalid form data")
		return
	}

	input := service.LoginInput{
		Email:     strings.ToLower(strings.TrimSpace(r.FormValue("email"))),
		Password:  r.FormValue("password"),
		UserAgent: r.UserAgent(),
		IPAddress: clientIP(r),
	}

	tokens, err := h.authService.Login(r.Context(), input)
	if err != nil {
		if domainErr, ok := apperr.IsDomainError(err); ok {
			renderAuthError(w, r, "login", domainErr.Message)
			return
		}
		log.Error().Err(err).Msg("login failed")
		renderAuthError(w, r, "login", "An unexpected error occurred")
		return
	}

	setAuthCookies(w, tokens)

	// HTMX redirect
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/app/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/app/dashboard", http.StatusSeeOther)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.authService.Logout(r.Context()); err != nil {
		log.Error().Err(err).Msg("logout failed")
	}
	clearAuthCookies(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// secureCookies is set at init based on environment.
// In production (HTTPS), cookies are Secure; in dev (HTTP), they are not.
// Uses atomic.Bool to prevent data races from concurrent HTTP handlers.
var secureCookies atomic.Bool

// SetSecureCookies configures whether auth cookies use the Secure flag.
func SetSecureCookies(secure bool) {
	secureCookies.Store(secure)
}

func setAuthCookies(w http.ResponseWriter, tokens *service.TokenPair) {
	secure := secureCookies.Load()
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  tokens.ExpiresAt,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/auth/refresh",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	secure := secureCookies.Load()
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth/refresh",
		HttpOnly: true,
		Secure:   secure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// clientIP extracts the IP address from r.RemoteAddr, stripping the port.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func renderAuthError(w http.ResponseWriter, r *http.Request, page, message string) {
	w.WriteHeader(http.StatusUnprocessableEntity)

	// For HTMX requests, return just the error fragment (form targets #auth-error).
	// For full page loads, return the complete page with error.
	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(pages.AuthError(message)).ServeHTTP(w, r)
		return
	}

	csrf := middleware.GetCSRFToken(r)
	var component templ.Component
	switch page {
	case "login":
		component = pages.Login(csrf, message)
	case "register":
		component = pages.Register(csrf, message)
	default:
		component = pages.Login(csrf, message)
	}
	templ.Handler(component).ServeHTTP(w, r)
}
