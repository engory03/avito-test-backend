//go:build integration

package integrationtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/app"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dummy"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	AccessSecret  = "integration-access-secret-key-32b!"
	RefreshSecret = "integration-refresh-secret-key-32bytes!"
)

// OpenPostgresAndMigrate starts Postgres (testcontainers), opens GORM, runs app migrations.
func OpenPostgresAndMigrate(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("itest"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(2*time.Minute)),
	)
	if err != nil {
		t.Fatalf("postgres container: %v", err)
	}
	dsn, err := pgDSN(ctx, container)
	if err != nil {
		terminate(ctx, container)
		t.Fatalf("dsn: %v", err)
	}
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		terminate(ctx, container)
		t.Fatalf("gorm open: %v", err)
	}
	if err := app.MigrateDB(db); err != nil {
		terminate(ctx, container)
		t.Fatalf("migrate: %v", err)
	}
	cleanup := func() { terminate(context.Background(), container) }
	return db, cleanup
}

// SeedDummyUsers inserts admin/user dummy rows used by dummy JWT login.
func SeedDummyUsers(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := dummy.EnsureUsersInDB(context.Background(), db); err != nil {
		t.Fatalf("dummy users: %v", err)
	}
}

// NewJWTManager matches secrets passed to NewTestEngine.
func NewJWTManager() *utils.JWTManager {
	j := utils.NewJWTManager(AccessSecret, RefreshSecret)
	j.AccessTokenTTL = 15 * time.Minute
	j.RefreshTokenTTL = 7 * 24 * time.Hour
	return j
}

// NewTestEngine returns a full Gin engine against a migrated DB with dummy users; registers t.Cleanup for the container.
func NewTestEngine(t *testing.T) (*gin.Engine, *utils.JWTManager) {
	t.Helper()
	db, cleanup := OpenPostgresAndMigrate(t)
	t.Cleanup(cleanup)
	SeedDummyUsers(t, db)
	j := NewJWTManager()
	r := app.NewEngine(db, app.EngineConfig{
		JWTAccessSecret:   AccessSecret,
		JWTRefreshSecret:  RefreshSecret,
		AccessExpiration:  j.AccessTokenTTL,
		RefreshExpiration: j.RefreshTokenTTL,
	})
	return r, j
}

func pgDSN(ctx context.Context, c *tcpostgres.PostgresContainer) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	if host == "localhost" {
		host = "127.0.0.1"
	}
	mp, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"host=%s port=%s user=postgres password=postgres dbname=itest sslmode=disable TimeZone=UTC",
		host, mp.Port(),
	), nil
}

func terminate(ctx context.Context, c *tcpostgres.PostgresContainer) {
	if c == nil {
		return
	}
	termCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_ = c.Terminate(termCtx)
}
