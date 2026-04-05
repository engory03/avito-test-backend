//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dummy"
	"github.com/internships-backend/test-backend-engory03/internal/integrationtest"
	userdto "github.com/internships-backend/test-backend-engory03/internal/users/dto"
)

func TestUserHandler_Integration_Me_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := integrationtest.NewTestEngine(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", w.Code)
	}
}

func TestUserHandler_Integration_Me_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, j := integrationtest.NewTestEngine(t)

	token, err := j.GenerateAccessToken(dummy.DummyUserID, "user")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var u userdto.UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		t.Fatal(err)
	}
	if u.Email != "user@mail.ru" {
		t.Fatalf("email %q", u.Email)
	}
}
