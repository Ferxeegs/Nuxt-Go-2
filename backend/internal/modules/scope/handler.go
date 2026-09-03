package scope

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
// @Summary   List scopes
// @Tags      admin-scopes
// @Security  BearerAuth
// @Produce   json
// @Param     page query int false "Page"
// @Param     per_page query int false "Per page"
// @Param     search query string false "Search"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/scopes [get]
func (h *Handler) List(c *gin.Context) {
	var q ListQuery
	_ = c.ShouldBindQuery(&q)
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
// @Summary   Get scope by ID
// @Tags      admin-scopes
// @Security  BearerAuth
// @Produce   json
// @Param     id path int true "Scope ID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/scopes/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	item, err := h.service.Get(id)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, item, "OK")
}

// Create godoc
// @Summary   Create scope
// @Tags      admin-scopes
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     body body CreateRequest true "Scope payload"
// @Success   201 {object} response.Meta
// @Router    /api/v1/admin/scopes [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Created(c, item, "Scope created successfully")
}

// Update godoc
// @Summary   Update scope
// @Tags      admin-scopes
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id path int true "Scope ID"
// @Param     body body UpdateRequest true "Scope payload"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/scopes/{id} [put]
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
	item, err := h.service.Update(id, req)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, item, "Scope updated successfully")
}

// Delete godoc
// @Summary   Delete scope
// @Tags      admin-scopes
// @Security  BearerAuth
// @Produce   json
// @Param     id path int true "Scope ID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/scopes/{id} [delete]
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
	response.Success(c, nil, "Scope deleted successfully")
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, apperror.BadRequest("Invalid ID")
	}
	return uint(id), nil
}
