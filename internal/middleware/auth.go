package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	JTIKey       contextKey = "jti"
	RequestIDKey contextKey = "request_id"
)

// Auth validates JWT access tokens from cookies.
type Auth struct {
	jwtSecret    []byte
	issuer       string
	audience     string
	isRevoked    func(ctx context.Context, jti string) (bool, error)
}

// NewAuth creates a new auth middleware.
func NewAuth(jwtSecret string, isRevoked func(ctx context.Context, jti string) (bool, error)) *Auth {
	return &Auth{
		jwtSecret: []byte(jwtSecret),
		issuer:    "finai",
		audience:  "finai-app",
		isRevoked: isRevoked,
	}
}

// RequireAuth is middleware that requires a valid JWT.
func (a *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractToken(r)
		if tokenStr == "" {
			sendUnauthorized(w, r, "Authentication required")
			return
		}

		claims, err := a.validateToken(tokenStr)
		if err != nil {
			sendUnauthorized(w, r, "Invalid or expired token")
			return
		}

		// Reject tokens without JTI unconditionally
		jti, _ := claims.GetJTI()
		if jti == "" {
			sendUnauthorized(w, r, "Invalid token: missing JTI")
			return
		}

		// Check token revocation
		if a.isRevoked != nil {
			revoked, err := a.isRevoked(r.Context(), jti)
			if err != nil || revoked {
				sendUnauthorized(w, r, "Token has been revoked")
				return
			}
		}

		// Extract user ID and add to context
		sub, err := claims.GetSubject()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, JTIKey, jti)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FinAIClaims extends standard JWT claims with a JTI.
type FinAIClaims struct {
	jwt.RegisteredClaims
}

// GetJTI returns the JWT ID claim.
func (c FinAIClaims) GetJTI() (string, error) {
	return c.ID, nil
}

func (a *Auth) validateToken(tokenStr string) (*FinAIClaims, error) {
	claims := &FinAIClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Enforce HS256 algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return a.jwtSecret, nil
	},
		jwt.WithIssuer(a.issuer),
		jwt.WithAudience(a.audience),
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func extractToken(r *http.Request) string {
	// Try cookie first (primary for HTMX)
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Fallback to Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

// sendUnauthorized handles auth failures for HTML, HTMX, and JSON requests.
func sendUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/auth/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if isHTMLRequest(r) {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	resp := map[string]any{"error": map[string]string{"code": "UNAUTHORIZED", "message": message}}
	json.NewEncoder(w).Encode(resp)
}

func isHTMLRequest(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html")
}

func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetJTI extracts the JWT ID from the request context.
func GetJTI(ctx context.Context) (string, bool) {
	jti, ok := ctx.Value(JTIKey).(string)
	return jti, ok
}
