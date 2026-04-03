package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dto"
	"github.com/internships-backend/test-backend-engory03/internal/auth/service"
	shareddto "github.com/internships-backend/test-backend-engory03/internal/shared/dto"
	httphelper "github.com/internships-backend/test-backend-engory03/internal/shared/http"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

// DummyLogin godoc
// @Summary      Получить тестовый JWT по роли (ОБЯЗАТЕЛЬНО). Доступен без авторизации.
// @Description  Выдаёт access JWT: фиксированный `user_id` (и `sub`) для роли admin и отдельный для user; в токене также `role`. Только значения admin и user.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.DummyLoginRequest true "Поле role: admin или user"
// @Success      200 {object} dto.TokenResponse "JWT для Authorization Bearer"
// @Failure      400 {object} shareddto.APIErrorResponse "Недопустимое значение роли"
// @Failure      500 {object} shareddto.APIErrorResponse "Внутренняя ошибка сервера"
// @Router       /dummyLogin [post]
func (h *AuthHandler) DummyLogin(c *gin.Context) {
	var req dto.DummyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	token, err := h.service.DummyLogin(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidDummyRole) {
			httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{Token: token})
}

// Register godoc
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param register body dto.RegisterRequest true "User registration data"
// @Success 200 {object} dto.AuthTokens
// @Failure 400 {object} shareddto.ErrorDefaultResponse
// @Failure 409 {object} shareddto.ErrorDefaultResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !httphelper.BindJSON(c, &req) {
		return
	}

	tokens, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "user with this username already exists" {
			httphelper.JSONError(c, http.StatusConflict, err)
		} else {
			httphelper.JSONError(c, http.StatusBadRequest, err)
		}
		return
	}

	httphelper.SetRefreshTokenCookie(c, tokens.RefreshToken)
	httphelper.JSONAccessToken(c, tokens.AccessToken)
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "User credentials"
// @Success 200 {object} dto.AuthTokens
// @Failure 400 {object} shareddto.ErrorDefaultResponse
// @Failure 401 {object} shareddto.ErrorDefaultResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !httphelper.BindJSON(c, &req) {
		return
	}

	tokens, err := h.service.Login(req)
	if err != nil {
		httphelper.JSONError(c, http.StatusUnauthorized, err)
		return
	}

	httphelper.SetRefreshTokenCookie(c, tokens.RefreshToken)
	httphelper.JSONAccessToken(c, tokens.AccessToken)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.AuthTokens
// @Failure 400 {object} shareddto.ErrorDefaultResponse
// @Failure 401 {object} shareddto.ErrorDefaultResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	// Extract token from httpOnly cookie
	refreshToken, err := httphelper.GetRefreshTokenFromCookie(c)
	if err != nil {
		httphelper.JSONError(c, http.StatusBadRequest, errors.New("refresh token cookie missing"))
		return
	}

	// Attempt to refresh tokens
	tokens, err := h.service.Refresh(refreshToken)
	if err != nil {
		httphelper.JSONError(c, http.StatusUnauthorized, err)
		return
	}

	// Rotate refresh token cookie
	httphelper.SetRefreshTokenCookie(c, tokens.RefreshToken)

	// Return access token in JSON body
	httphelper.JSONAccessToken(c, tokens.AccessToken)
}

// Logout godoc
// @Summary Logout user (invalidate refresh token)
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} shareddto.MessageDefaultResponse
// @Failure 400 {object} shareddto.ErrorDefaultResponse
// @Failure 500 {object} shareddto.ErrorDefaultResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Read refresh token from HTTP-only cookie
	refreshToken, err := httphelper.GetRefreshTokenFromCookie(c)
	if err != nil {
		// cookie missing or invalid
		httphelper.JSONError(c, http.StatusBadRequest, errors.New("refresh token cookie not found"))
		return
	}

	// Try to remove refresh token from store (service should handle missing token gracefully)
	if err := h.service.Logout(refreshToken); err != nil {
		// return 500 only on internal error; if token was already gone, service can return nil
		httphelper.JSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Clear cookie on client
	httphelper.ClearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, shareddto.MessageDefaultResponse{Message: "ok"})
}
