package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	httphelper "github.com/internships-backend/test-backend-engory03/internal/shared/http"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			httphelper.JSONAPIError(c, http.StatusForbidden, "FORBIDDEN", "missing role in context")
			c.Abort()
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			httphelper.JSONAPIError(c, http.StatusForbidden, "FORBIDDEN", "invalid role type")
			c.Abort()
			return
		}

		if !allowed[role] {
			httphelper.JSONAPIError(c, http.StatusForbidden, "FORBIDDEN", "forbidden: insufficient role")
			c.Abort()
			return
		}

		c.Next()
	}
}
