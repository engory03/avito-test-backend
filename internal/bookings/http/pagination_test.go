package http

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseBookingListPagination_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/bookings/list", nil)
	page, ps, ok := parseBookingListPagination(c)
	if !ok || page != 1 || ps != 20 {
		t.Fatalf("%d %d %v", page, ps, ok)
	}
}

func TestParseBookingListPagination_Custom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/bookings/list?page=2&pageSize=50", nil)
	page, ps, ok := parseBookingListPagination(c)
	if !ok || page != 2 || ps != 50 {
		t.Fatalf("%d %d", page, ps)
	}
}

func TestParseBookingListPagination_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, q := range []string{"?page=0", "?page=abc", "?pageSize=0", "?pageSize=101", "?pageSize=xx"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/list"+q, nil)
		_, _, ok := parseBookingListPagination(c)
		if ok {
			t.Fatalf("expected invalid for %s", q)
		}
	}
}
