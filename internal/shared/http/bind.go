package httphelper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BindJSON is a convenience function to bind a JSON payload and
// automatically return a 400 error if binding fails.
// It returns true if binding succeeded, false otherwise.
func BindJSON[T any](c *gin.Context, req *T) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return false
	}
	return true
}
