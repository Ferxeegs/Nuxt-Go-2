package user

import (
	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/middleware"
	"go-cp-boilerplate/internal/pkg/response"
	"go-cp-boilerplate/internal/pkg/validator"
)

type Handler struct {
	service         Service
	bcryptCost      int
	defaultPassword string
}

func NewHandler(service Service, bcryptCost int, defaultPassword string) *Handler {
	return &Handler{service: service, bcryptCost: bcryptCost, defaultPassword: defaultPassword}
}

// List godoc
// @Summary   List users
// @Tags      admin-users
// @Security  BearerAuth
// @Produce   json
// @Param     page query int false "page"
// @Param     per_page query int false "per_page"
// @Param     search query string false "search"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/users [get]
func (h *Handler) List(c *gin.Context) {
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, 400, "Invalid query parameters", nil)
		return
	}

	items, meta, err := h.service.List(q)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.SuccessWithMeta(c, items, "OK", response.Meta{
		Page: meta.Page, PerPage: meta.PerPage, TotalData: meta.TotalData, TotalPages: meta.TotalPages,
	})
}

// Get godoc
// @Summary   Get user by ID
// @Tags      admin-users
// @Security  BearerAuth
// @Produce   json
// @Param     id path string true "User UUID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/users/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := ParseUserID(c.Param("id"))
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	u, err := h.service.Get(id)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, u, "OK")
}

// Create godoc
// @Summary   Create user
// @Tags      admin-users
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     body body CreateRequest true "User payload"
// @Success   201 {object} response.Meta
// @Router    /api/v1/admin/users [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}
	actorRole, _ := c.Get(middleware.ContextRole)
	roleName, _ := actorRole.(string)
	u, err := h.service.Create(req, h.bcryptCost, roleName)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Created(c, u, "User created successfully")
}

// Update godoc
// @Summary   Update user
// @Tags      admin-users
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id path string true "User UUID"
// @Param     body body UpdateRequest true "User payload"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/users/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := ParseUserID(c.Param("id"))
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}
	actorRole, _ := c.Get(middleware.ContextRole)
	roleName, _ := actorRole.(string)
	u, err := h.service.Update(id, req, roleName)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, u, "User updated successfully")
}

// ResetPassword godoc
// @Summary   Reset user password to DEFAULT_PASSWORD
// @Tags      admin-users
// @Security  BearerAuth
// @Produce   json
// @Param     id path string true "User UUID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/users/{id}/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	id, err := ParseUserID(c.Param("id"))
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	if err := h.service.ResetPassword(id, h.bcryptCost, h.defaultPassword); err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, nil, "Password reset successfully")
}

// Delete godoc
// @Summary   Delete user
// @Tags      admin-users
// @Security  BearerAuth
// @Produce   json
// @Param     id path string true "User UUID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := ParseUserID(c.Param("id"))
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	if err := h.service.Delete(id); err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, nil, "User deleted successfully")
}
