//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/integrationtest"
	"github.com/internships-backend/test-backend-engory03/internal/users/model"
	"github.com/internships-backend/test-backend-engory03/internal/users/repository"
)

func TestUserRepository_Integration_CRUD(t *testing.T) {
	db, cleanup := integrationtest.OpenPostgresAndMigrate(t)
	t.Cleanup(cleanup)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	u := &model.User{
		ID:           uuid.New(),
		Email:        "integration-user@example.com",
		PasswordHash: "hashed",
		Role:         "user",
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil || got == nil || got.Email != u.Email {
		t.Fatalf("GetByID: %+v %v", got, err)
	}

	byEmail, err := repo.GetByEmail(ctx, u.Email)
	if err != nil || byEmail == nil || byEmail.ID != u.ID {
		t.Fatalf("GetByEmail: %+v %v", byEmail, err)
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, x := range all {
		if x.ID == u.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created user not in GetAll")
	}
}
