package httphelper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseExpand(t *testing.T) {
	m := ParseExpand([]string{"a,b", " c ", ""})
	if !m["a"] || !m["b"] || !m["c"] {
		t.Fatalf("%v", m)
	}
}

func TestJSONError_JSONAPIError_JSONAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	JSONError(c, 400, errSample{})
	if w.Code != 400 {
		t.Fatal(w.Code)
	}

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	JSONAPIError(c2, 403, "C", "msg")
	if w2.Code != 403 {
		t.Fatal(w2.Code)
	}

	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	JSONAccessToken(c3, "tok")
	if w3.Code != 200 {
		t.Fatal(w3.Code)
	}
}

type errSample struct{}

func (errSample) Error() string { return "x" }

func TestBindJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	var v struct{ A int }
	if BindJSON(c, &v) {
		t.Fatal("expected bind fail on empty body")
	}
}

func TestShouldInclude(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?include=slots", nil)
	if !ShouldInclude(c, "slots") || ShouldInclude(c, "other") {
		t.Fatal()
	}
}

func TestSetRefreshTokenCookie_AndClear(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	SetRefreshTokenCookie(c, "abc")
	ClearRefreshTokenCookie(c)
}

func TestGetRefreshTokenFromCookie_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "tokval"})
	c.Request = req
	got, err := GetRefreshTokenFromCookie(c)
	if err != nil || got != "tokval" {
		t.Fatalf("%v %v", got, err)
	}
}

func TestGetRefreshTokenFromCookie_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	_, err := GetRefreshTokenFromCookie(c)
	if err == nil {
		t.Fatal("expected missing cookie error")
	}
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: ""})
	c2.Request = req
	_, err = GetRefreshTokenFromCookie(c2)
	if err == nil {
		t.Fatal("expected empty cookie error")
	}
}
