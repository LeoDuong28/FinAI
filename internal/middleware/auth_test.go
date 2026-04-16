package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nghiaduong/finai/internal/middleware"
)

const testJWTSecret = "test-jwt-secret-that-is-at-least-32-chars-long"

func makeToken(t *testing.T, secret string, claims middleware.FinAIClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func validClaims() middleware.FinAIClaims {
	return middleware.FinAIClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   uuid.New().String(),
			Issuer:    "finai",
			Audience:  jwt.ClaimStrings{"finai-app"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
}

func neverRevoked(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func TestAuth_ValidTokenInCookie(t *testing.T) {
	auth := middleware.NewAuth(testJWTSecret, neverRevoked)
	claims := validClaims()
	tokenStr := makeToken(t, testJWTSecret, claims)

	var gotUserID uuid.UUID
	var gotJTI string
	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = middleware.GetUserID(r.Context())
		gotJTI, _ = middleware.GetJTI(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	sub, _ := claims.GetSubject()
	assert.Equal(t, sub, gotUserID.String())
	jti, _ := claims.GetJTI()
	assert.Equal(t, jti, gotJTI)
}

func TestAuth_ValidTokenInAuthHeader(t *testing.T) {
	auth := middleware.NewAuth(testJWTSecret, neverRevoked)
	claims := validClaims()
	tokenStr := makeToken(t, testJWTSecret, claims)

	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuth_ExpiredToken(t *testing.T) {
	auth := middleware.NewAuth(testJWTSecret, neverRevoked)
	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-1 * time.Hour))
	tokenStr := makeToken(t, testJWTSecret, claims)

	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_MissingJTI(t *testing.T) {
	auth := middleware.NewAuth(testJWTSecret, neverRevoked)
	claims := validClaims()
	claims.ID = "" // no JTI
	tokenStr := makeToken(t, testJWTSecret, claims)

	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_RevokedToken(t *testing.T) {
	alwaysRevoked := func(_ context.Context, _ string) (bool, error) {
		return true, nil
	}
	auth := middleware.NewAuth(testJWTSecret, alwaysRevoked)
	claims := validClaims()
	tokenStr := makeToken(t, testJWTSecret, claims)

	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_NoToken(t *testing.T) {
	auth := middleware.NewAuth(testJWTSecret, neverRevoked)

	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_WrongSigningMethod(t *testing.T) {
	auth := middleware.NewAuth(testJWTSecret, neverRevoked)

	// Create a token with HS384 (wrong algorithm — only HS256 accepted)
	claims := validClaims()
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenStr, _ := token.SignedString([]byte(testJWTSecret))

	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
