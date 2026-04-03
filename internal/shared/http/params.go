package httphelper

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ShouldInclude checks if the "include" query parameter matches the given key.
func ShouldInclude(c *gin.Context, key string) bool {
	return c.Query("include") == key
}

// ParseExpand returns a normalized list of expand keys.
// Supports: ?expand=course,branch OR ?expand=course&expand=branch
func ParseExpand(values []string) map[string]bool {
	result := make(map[string]bool)

	for _, v := range values {
		parts := strings.SplitSeq(v, ",")
		for p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result[p] = true
			}
		}
	}

	return result
}
