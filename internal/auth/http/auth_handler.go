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
// @Summary      Регистрация нового пользователя
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        register body dto.RegisterRequest true "Данные для регистрации пользователя"
// @Success      200 {object} dto.TokenResponse "JWT для Authorization Bearer"
// @Failure      400 {object} shareddto.ErrorDefaultResponse "Неверный запрос"
// @Failure      409 {object} shareddto.ErrorDefaultResponse "Пользователь с таким email уже существует"
// @Router       /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	tokens, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "user with this email already exists" {
			httphelper.JSONAPIError(c, http.StatusConflict, "EMAIL_EXISTS", err.Error())
		} else {
			httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		}
		return
	}

	httphelper.SetRefreshTokenCookie(c, tokens.RefreshToken)
	c.JSON(http.StatusOK, dto.TokenResponse{Token: tokens.AccessToken})
}

// Login godoc
// @Summary Вход пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Учетные данные пользователя"
// @Success 200 {object} dto.TokenResponse "JWT для Authorization Bearer"
// @Failure 400 {object} shareddto.ErrorDefaultResponse "Неверный запрос (ошибка валидации)"
// @Failure 401 {object} shareddto.ErrorDefaultResponse "Неверная почта или пароль"
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	tokens, err := h.service.Login(req)
	if err != nil {
		httphelper.JSONAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	httphelper.SetRefreshTokenCookie(c, tokens.RefreshToken)
	c.JSON(http.StatusOK, dto.TokenResponse{Token: tokens.AccessToken})
}

// Refresh godoc
// @Summary Обновить access token
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.TokenResponse "JWT для Authorization Bearer"
// @Failure 400 {object} shareddto.ErrorDefaultResponse "Неверный запрос (refresh token отсутствует или некорректен)"
// @Failure 401 {object} shareddto.ErrorDefaultResponse "Неавторизован (refresh token невалиден или истёк)"
// @Router /refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	// Извлекаем токен из HTTP-only cookie
	refreshToken, err := httphelper.GetRefreshTokenFromCookie(c)
	if err != nil {
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "refresh token cookie missing")
		return
	}

	// Пытаемся обновить токены
	tokens, err := h.service.Refresh(refreshToken)
	if err != nil {
		httphelper.JSONAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	// Обновляем refresh token cookie
	httphelper.SetRefreshTokenCookie(c, tokens.RefreshToken)

	// Возвращаем access token в JSON body
	c.JSON(http.StatusOK, dto.TokenResponse{Token: tokens.AccessToken})
}

// Logout godoc
// @Summary Выход пользователя (удаление refresh token)
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} shareddto.MessageDefaultResponse "Токены успешно удалены"
// @Failure 400 {object} shareddto.ErrorDefaultResponse "Неверный запрос (refresh token отсутствует или некорректен)"
// @Failure 500 {object} shareddto.ErrorDefaultResponse "Внутренняя ошибка сервера"
// @Router /logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Чтение refresh token из HTTP-only cookie
	refreshToken, err := httphelper.GetRefreshTokenFromCookie(c)
	if err != nil {
		// cookie отсутствуют или некорректны
		httphelper.JSONAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "refresh token cookie not found")
		return
	}

	// Пытаемся удалить refresh token из хранилища (сервис должен корректно обрабатывать отсутствие токена)
	if err := h.service.Logout(refreshToken); err != nil {
		// возвращать 500 только при внутренней ошибке; если токен уже был удалён, сервис может вернуть nil
		httphelper.JSONAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httphelper.ClearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, shareddto.MessageDefaultResponse{Message: "ok"})
}
