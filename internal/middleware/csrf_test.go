package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestCSRF_GETPassesWithoutToken(t *testing.T) {
	csrf := middleware.NewCSRF(false)
	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRF_GETSetsCookie(t *testing.T) {
	csrf := middleware.NewCSRF(false)
	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			found = true
			assert.NotEmpty(t, c.Value)
		}
	}
	assert.True(t, found, "csrf_token cookie should be set")
}

func TestCSRF_POSTWithValidToken(t *testing.T) {
	csrf := middleware.NewCSRF(false)
	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := "test-csrf-token-value"
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRF_POSTWithoutHeader(t *testing.T) {
	csrf := middleware.NewCSRF(false)
	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	token := "test-csrf-token-value"
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	// No X-CSRF-Token header
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRF_POSTWithMismatchedToken(t *testing.T) {
	csrf := middleware.NewCSRF(false)
	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "cookie-token"})
	req.Header.Set("X-CSRF-Token", "different-header-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRF_GetCSRFToken(t *testing.T) {
	csrf := middleware.NewCSRF(false)
	var extractedToken string

	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedToken = middleware.GetCSRFToken(r)
		w.WriteHeader(http.StatusOK)
	}))

	token := "my-csrf-token"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, token, extractedToken)
}
