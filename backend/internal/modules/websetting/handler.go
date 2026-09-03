package websetting

import (
	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/pkg/response"
	"go-cp-boilerplate/internal/pkg/validator"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ListPublic godoc
// @Summary   Get public web settings
// @Tags      public-settings
// @Produce   json
// @Success   200 {object} response.Meta
// @Router    /api/v1/public/settings [get]
func (h *Handler) ListPublic(c *gin.Context) {
	items, err := h.service.All()
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, items, "OK")
}

// ListAdmin godoc
// @Summary   Get web settings (admin)
// @Tags      admin-settings
// @Security  BearerAuth
// @Produce   json
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/settings [get]
func (h *Handler) ListAdmin(c *gin.Context) {
	h.ListPublic(c)
}

// Update godoc
// @Summary   Update web settings
// @Tags      admin-settings
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     body body UpsertRequest true "Settings payload"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/settings [put]
func (h *Handler) Update(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}
	items, err := h.service.Upsert(req)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, items, "Settings updated successfully")
}
