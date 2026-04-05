package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	shareddto "github.com/internships-backend/test-backend-engory03/internal/shared/dto"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
)

const testAccessSecret = "unit-test-access-secret-key-32bytes!"
const testRefreshSecret = "unit-test-refresh-secret-key-32bytes!"

func testJWTManager() *utils.JWTManager {
	return utils.NewJWTManager(testAccessSecret, testRefreshSecret)
}

func parseAPIError(t *testing.T, body []byte) shareddto.APIErrorResponse {
	t.Helper()
	var out shareddto.APIErrorResponse
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	j := testJWTManager()
	var next bool
	r := gin.New()
	r.Use(AuthMiddleware(j))
	r.GET("/x", func(c *gin.Context) { next = true; c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", http.NoBody))
	if next || w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d next=%v", w.Code, next)
	}
	errBody := parseAPIError(t, w.Body.Bytes())
	if errBody.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("code %q", errBody.Error.Code)
	}
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	j := testJWTManager()
	r := gin.New()
	r.Use(AuthMiddleware(j))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
	req.Header.Set("Authorization", "Basic x")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	j := testJWTManager()
	r := gin.New()
	r.Use(AuthMiddleware(j))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
	req.Header.Set("Authorization", "Bearer not.a.jwt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
}

func TestAuthMiddleware_WrongSigningSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uid := uuid.New()
	other := utils.NewJWTManager("different-access-secret-key-32bytes!!", testRefreshSecret)
	tok, err := other.GenerateAccessToken(uid, "user")
	if err != nil {
		t.Fatal(err)
	}
	j := testJWTManager()
	r := gin.New()
	r.Use(AuthMiddleware(j))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
}

func TestAuthMiddleware_InvalidUserIDClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)
	j := testJWTManager()
	claims := &utils.AccessTokenClaims{
		Role:   "user",
		UserID: "not-a-valid-uuid",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := token.SignedString([]byte(testAccessSecret))
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.Use(AuthMiddleware(j))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
	msg := parseAPIError(t, w.Body.Bytes()).Error.Message
	if msg != "invalid user id" {
		t.Fatalf("message %q", msg)
	}
}

func TestAuthMiddleware_OK_SetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	j := testJWTManager()
	uid := uuid.New()
	tok, err := j.GenerateAccessToken(uid, "admin")
	if err != nil {
		t.Fatal(err)
	}

	var gotUID uuid.UUID
	var gotRole string
	r := gin.New()
	r.Use(AuthMiddleware(j))
	r.GET("/x", func(c *gin.Context) {
		v, _ := c.Get("user_id")
		gotUID = v.(uuid.UUID)
		rval, _ := c.Get("role")
		gotRole = rval.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	if gotUID != uid || gotRole != "admin" {
		t.Fatalf("uid=%v role=%q", gotUID, gotRole)
	}
}

func TestRoleMiddleware_MissingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var next bool
	r := gin.New()
	r.Use(RoleMiddleware("admin"))
	r.GET("/x", func(c *gin.Context) { next = true; c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", http.NoBody))
	if next || w.Code != http.StatusForbidden {
		t.Fatalf("code=%d next=%v", w.Code, next)
	}
	errBody := parseAPIError(t, w.Body.Bytes())
	if errBody.Error.Code != "FORBIDDEN" || errBody.Error.Message != "missing role in context" {
		t.Fatalf("%+v", errBody.Error)
	}
}

func TestRoleMiddleware_InvalidRoleType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", 42)
		c.Next()
	})
	r.Use(RoleMiddleware("admin"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", http.NoBody))
	if w.Code != http.StatusForbidden {
		t.Fatal(w.Code)
	}
}

func TestRoleMiddleware_InsufficientRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	})
	r.Use(RoleMiddleware("admin"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", http.NoBody))
	if w.Code != http.StatusForbidden {
		t.Fatal(w.Code)
	}
	msg := parseAPIError(t, w.Body.Bytes()).Error.Message
	if msg != "forbidden: insufficient role" {
		t.Fatalf("message %q", msg)
	}
}

func TestRoleMiddleware_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var next bool
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	r.Use(RoleMiddleware("admin", "super"))
	r.GET("/x", func(c *gin.Context) {
		next = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", http.NoBody))
	if !next || w.Code != http.StatusOK {
		t.Fatalf("code=%d next=%v", w.Code, next)
	}
}
