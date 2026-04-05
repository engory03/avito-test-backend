//go:build integration

package repository_test

import (
	"os"
	"testing"

	"github.com/internships-backend/test-backend-engory03/internal/integrationtest"
)

func TestMain(m *testing.M) {
	os.Exit(integrationtest.RunWithSharedPostgres(m))
}
