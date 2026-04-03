package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/internships-backend/test-backend-engory03/internal/shared/dto"
	httphelper "github.com/internships-backend/test-backend-engory03/internal/shared/http"
	"github.com/internships-backend/test-backend-engory03/internal/users/dto"
	"github.com/internships-backend/test-backend-engory03/internal/users/mapper"
	"github.com/internships-backend/test-backend-engory03/internal/users/service"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

// CreateUser godoc
// @Summary Create user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body dto.CreateUserRequest true "New user"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} shareddto.ErrorDefaultResponse
// @Failure 500 {object} shareddto.ErrorDefaultResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest

	if !httphelper.BindJSON(c, &req) {
		return
	}

	user, err := h.service.CreateUser(req)
	if err != nil {
		httphelper.JSONError(c, http.StatusInternalServerError, err)
		return
	}

	resp := mapper.NewUserResponse(user)

	c.JSON(http.StatusCreated, resp)
}

// GetCurrentUser godoc
// @Summary Get current user
// @Description Returns the authenticated user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} shareddto.ErrorDefaultResponse
// @Failure 404 {object} shareddto.ErrorDefaultResponse
// @Router /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		httphelper.JSONError(c, http.StatusBadRequest, errors.New("user_id not found in context"))
		return
	}
	id, ok := userID.(uuid.UUID)
	if !ok {
		httphelper.JSONError(c, http.StatusBadRequest, errors.New("invalid user_id type"))
		return
	}

	user, err := h.service.GetUserById(id)
	if err != nil {
		httphelper.JSONError(c, http.StatusNotFound, err)
		return

	}

	resp := mapper.NewUserResponse(user)

	c.JSON(http.StatusOK, resp)
}

// GetUserById godoc
// @Summary Get user by ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID (uuid)"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} shareddto.ErrorDefaultResponse
// @Failure 404 {object} shareddto.ErrorDefaultResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUserById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httphelper.JSONError(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.service.GetUserById(id)
	if err != nil {
		httphelper.JSONError(c, http.StatusNotFound, err)
		return
	}

	resp := mapper.NewUserResponse(user)

	c.JSON(http.StatusOK, resp)
}

// GetAllUsers godoc
// @Summary Get all users
// @Tags Users
// @Produce json
// @Success 200 {array} dto.UserResponse
// @Router /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		httphelper.JSONError(c, http.StatusInternalServerError, err)
		return
	}
	resp := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		resp[i] = mapper.NewUserResponse(user)
	}
	c.JSON(http.StatusOK, resp)

}
