package role

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/response"
	"go-cp-boilerplate/internal/pkg/validator"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List godoc
// @Summary   List roles
// @Tags      admin-roles
// @Security  BearerAuth
// @Produce   json
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/roles [get]
func (h *Handler) List(c *gin.Context) {
	roles, err := h.service.List()
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, roles, "OK")
}

// Get godoc
// @Summary   Get role by ID
// @Tags      admin-roles
// @Security  BearerAuth
// @Produce   json
// @Param     id path int true "Role ID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/roles/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	r, err := h.service.Get(id)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, r, "OK")
}

// Create godoc
// @Summary   Create role
// @Tags      admin-roles
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     body body CreateRequest true "Role payload"
// @Success   201 {object} response.Meta
// @Router    /api/v1/admin/roles [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}
	r, err := h.service.Create(req)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Created(c, r, "Role created successfully")
}

// Update godoc
// @Summary   Update role
// @Tags      admin-roles
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id path int true "Role ID"
// @Param     body body UpdateRequest true "Role payload"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/roles/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}
	r, err := h.service.Update(id, req)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, r, "Role updated successfully")
}

// Delete godoc
// @Summary   Delete role
// @Tags      admin-roles
// @Security  BearerAuth
// @Produce   json
// @Param     id path int true "Role ID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/roles/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	if err := h.service.Delete(id); err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, nil, "Role deleted successfully")
}

// ListPermissions godoc
// @Summary   List the full permission catalogue
// @Tags      admin-roles
// @Security  BearerAuth
// @Produce   json
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/roles/permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.service.ListPermissions()
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, perms, "OK")
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, apperror.BadRequest("Invalid ID")
	}
	return uint(id), nil
}
