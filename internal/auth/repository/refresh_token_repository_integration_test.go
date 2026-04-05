//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	authmodel "github.com/internships-backend/test-backend-engory03/internal/auth/model"
	"github.com/internships-backend/test-backend-engory03/internal/auth/repository"
	"github.com/internships-backend/test-backend-engory03/internal/integrationtest"
	usersmodel "github.com/internships-backend/test-backend-engory03/internal/users/model"
	usersrepo "github.com/internships-backend/test-backend-engory03/internal/users/repository"
)

func TestRefreshTokenRepository_Integration_SaveDelete(t *testing.T) {
	db := integrationtest.OpenPostgresAndMigrate(t)
	ctx := context.Background()

	ur := usersrepo.NewUserRepository(db)
	user := &usersmodel.User{
		ID:           uuid.New(),
		Email:        "rtoken-user@example.com",
		PasswordHash: "x",
		Role:         "user",
	}
	if err := ur.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	rr := repository.NewRefreshTokenRepository(db)
	tok := &authmodel.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     "refresh-token-value-integration",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	if err := rr.Save(ctx, tok); err != nil {
		t.Fatal(err)
	}
	if err := rr.Delete(ctx, tok.Token); err != nil {
		t.Fatal(err)
	}
}
