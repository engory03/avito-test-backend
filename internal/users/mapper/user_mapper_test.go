package mapper

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/users/model"
)

func TestNewUserResponse(t *testing.T) {
	now := time.Now().UTC()
	u := &model.User{
		ID: uuid.New(), Email: "e@x.y", Role: "user",
		CreatedAt: now, UpdatedAt: now,
	}
	r := NewUserResponse(u)
	if r.Email != u.Email || r.Role != "user" {
		t.Fatalf("%+v", r)
	}
}
