package e2e

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/app"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dummy"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// testAPIBaseURL is set in TestMain (in-process server + Postgres container, or E2E_BASE_URL).
var testAPIBaseURL string

func TestMain(m *testing.M) {
	if u := strings.TrimSpace(os.Getenv("E2E_BASE_URL")); u != "" {
		testAPIBaseURL = strings.TrimRight(u, "/")
		os.Exit(m.Run())
	}

	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("e2e"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(2*time.Minute)),
	)
	if err != nil {
		log.Fatalf("e2e postgres container: %v", err)
	}

	dsn, err := postgresLibPQDSN(ctx, container)
	if err != nil {
		terminateContainer(container)
		log.Fatalf("e2e postgres dsn: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		terminateContainer(container)
		log.Fatalf("e2e gorm open: %v", err)
	}

	if migErr := app.MigrateDB(db); migErr != nil {
		terminateContainer(container)
		log.Fatalf("e2e migrate: %v", migErr)
	}
	if seedErr := dummy.EnsureUsersInDB(db); seedErr != nil {
		terminateContainer(container)
		log.Fatalf("e2e dummy users: %v", seedErr)
	}

	engine := app.NewEngine(db, app.EngineConfig{
		JWTAccessSecret:   "e2e-jwt-access-secret-key-32bytes!",
		JWTRefreshSecret:  "e2e-jwt-refresh-secret-key-32bytes",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		terminateContainer(container)
		log.Fatalf("e2e listen: %v", err)
	}

	srv := &http.Server{
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("e2e http server: %v", err)
		}
	}()
	testAPIBaseURL = "http://" + ln.Addr().String()

	code := m.Run()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	_ = srv.Shutdown(shutdownCtx)
	cancel()

	terminateContainer(container)

	os.Exit(code)
}

// postgresLibPQDSN uses libpq key=value form and prefers 127.0.0.1 over localhost
// so pgx does not hit IPv6 (::1) on Windows Docker Desktop (often returns EOF).
func postgresLibPQDSN(ctx context.Context, c *tcpostgres.PostgresContainer) (string, error) {
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
		"host=%s port=%s user=postgres password=postgres dbname=e2e sslmode=disable TimeZone=UTC",
		host, mp.Port(),
	), nil
}

func terminateContainer(c *tcpostgres.PostgresContainer) {
	if c == nil {
		return
	}
	termCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.Terminate(termCtx); err != nil {
		log.Printf("e2e postgres terminate: %v", err)
	}
}
