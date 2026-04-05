//go:build integration

package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dummy"
	"github.com/internships-backend/test-backend-engory03/internal/integrationtest"
	roomdto "github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
)

func TestRoomHandler_Integration_Create_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := integrationtest.NewTestEngine(t)

	body := []byte(`{"name":"R1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", w.Code)
	}
}

func TestRoomHandler_Integration_Create_ForbiddenForUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, j := integrationtest.NewTestEngine(t)

	token, err := j.GenerateAccessToken(dummy.DummyUserID, "user")
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"name":"R1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d %s", w.Code, w.Body.String())
	}
}

func TestRoomHandler_Integration_CreateAndList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, j := integrationtest.NewTestEngine(t)

	adminTok, err := j.GenerateAccessToken(dummy.DummyAdminID, "admin")
	if err != nil {
		t.Fatal(err)
	}
	createBody := []byte(`{"name":"Integration Room X"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/create", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminTok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status %d %s", w.Code, w.Body.String())
	}
	var created roomdto.CreateRoomResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Room == nil || created.Room.Name != "Integration Room X" {
		t.Fatalf("room %+v", created.Room)
	}

	userTok, err := j.GenerateAccessToken(dummy.DummyUserID, "user")
	if err != nil {
		t.Fatal(err)
	}
	listReq := httptest.NewRequest(http.MethodGet, "/api/rooms/list", http.NoBody)
	listReq.Header.Set("Authorization", "Bearer "+userTok)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, listReq)
	if w2.Code != http.StatusOK {
		t.Fatalf("list status %d %s", w2.Code, w2.Body.String())
	}
	var list roomdto.ListRoomsResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, room := range list.Rooms {
		if room != nil && room.Name == "Integration Room X" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("room not listed: %+v", list.Rooms)
	}
}
