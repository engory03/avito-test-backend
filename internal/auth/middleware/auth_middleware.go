package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	httphelper "github.com/internships-backend/test-backend-engory03/internal/shared/http"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
)

func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			httphelper.JSONAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			httphelper.JSONAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header")
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := jwtManager.VerifyAccessToken(token)
		if err != nil {
			httphelper.JSONAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			httphelper.JSONAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid subject")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
