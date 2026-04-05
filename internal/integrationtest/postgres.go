//go:build integration

package integrationtest

import (
	"context"
	"fmt"
	"os"
	"sync"
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

var (
	sharedMu        sync.Mutex
	sharedDB        *gorm.DB
	sharedContainer *tcpostgres.PostgresContainer
)

// RunWithSharedPostgres starts a single Postgres container (testcontainers) for the whole
// test binary, runs migrations, then m.Run(), then tears down. Call from TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(integrationtest.RunWithSharedPostgres(m)) }
func RunWithSharedPostgres(m *testing.M) int {
	gin.SetMode(gin.TestMode)
	if err := startSharedPostgres(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "integrationtest: postgres: %v\n", err)
		return 1
	}
	code := m.Run()
	shutdownSharedPostgres()
	return code
}

// SharedDB returns the DB opened by RunWithSharedPostgres. Each test package must use TestMain.
func SharedDB(t *testing.T) *gorm.DB {
	t.Helper()
	sharedMu.Lock()
	defer sharedMu.Unlock()
	if sharedDB == nil {
		t.Fatal("integrationtest: SharedDB requires TestMain calling integrationtest.RunWithSharedPostgres")
	}
	return sharedDB
}

// OpenPostgresAndMigrate returns the shared migrated DB (same instance for all tests in the binary).
func OpenPostgresAndMigrate(t *testing.T) *gorm.DB {
	t.Helper()
	return SharedDB(t)
}

func startSharedPostgres() error {
	sharedMu.Lock()
	defer sharedMu.Unlock()
	if sharedDB != nil {
		return nil
	}
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
		return fmt.Errorf("postgres container: %w", err)
	}
	dsn, err := pgDSN(ctx, container)
	if err != nil {
		terminate(ctx, container)
		return fmt.Errorf("dsn: %w", err)
	}
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		terminate(ctx, container)
		return fmt.Errorf("gorm open: %w", err)
	}
	if err := app.MigrateDB(db); err != nil {
		terminate(ctx, container)
		return fmt.Errorf("migrate: %w", err)
	}
	sharedContainer = container
	sharedDB = db
	return nil
}

func shutdownSharedPostgres() {
	sharedMu.Lock()
	defer sharedMu.Unlock()
	if sharedDB != nil {
		if sqlDB, err := sharedDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
		sharedDB = nil
	}
	if sharedContainer != nil {
		terminate(context.Background(), sharedContainer)
		sharedContainer = nil
	}
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

// NewTestEngine returns a full Gin engine against the shared migrated DB with dummy users.
func NewTestEngine(t *testing.T) (*gin.Engine, *utils.JWTManager) {
	t.Helper()
	db := SharedDB(t)
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
