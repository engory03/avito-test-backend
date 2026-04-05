//go:build integration

package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dto"
	"github.com/internships-backend/test-backend-engory03/internal/integrationtest"
)

func TestAuthHandler_Integration_DummyLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := integrationtest.NewTestEngine(t)

	body := []byte(`{"role":"admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp dto.TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Token == "" {
		t.Fatal("empty token")
	}
}

func TestAuthHandler_Integration_DummyLogin_InvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := integrationtest.NewTestEngine(t)

	body := []byte(`{"role":"superuser"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d %s", w.Code, w.Body.String())
	}
}
