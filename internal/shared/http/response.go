package httphelper

import (
	"net/http"

	"github.com/gin-gonic/gin"
	shareddto "github.com/internships-backend/test-backend-engory03/internal/shared/dto"
)

// JSONError sends a standard error response.
func JSONError(c *gin.Context, status int, err error) {
	c.JSON(status, shareddto.ErrorDefaultResponse{Error: err.Error()})
}

// JSONAPIError sends an error body matching api.yaml ErrorResponse / InternalErrorResponse.
func JSONAPIError(c *gin.Context, status int, code, message string) {
	c.JSON(status, shareddto.APIErrorResponse{
		Error: shareddto.APIErrorDetail{Code: code, Message: message},
	})
}

// JSONAccessToken sends only the access token (no refresh token).
func JSONAccessToken(c *gin.Context, accessToken string) {
	c.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
	})
}
