package media

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/middleware"
	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Upload godoc
// @Summary   Upload a media file
// @Tags      admin-media
// @Security  BearerAuth
// @Accept    multipart/form-data
// @Produce   json
// @Param     file formData file true "Image file"
// @Param     module formData string false "Related module (e.g. article, user_avatar)"
// @Success   201 {object} response.Meta
// @Router    /api/v1/admin/media/upload [post]
func (h *Handler) Upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.FromAppError(c, apperror.BadRequest("No file uploaded"))
		return
	}

	module := c.PostForm("module")
	if module == "" {
		module = "general"
	}

	userID, _ := c.Get(middleware.ContextUserID)
	uploadedBy, _ := userID.(string)

	asset, err := h.service.Upload(fh, module, uploadedBy)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Created(c, asset, "File uploaded successfully")
}

// Delete godoc
// @Summary   Delete a media asset
// @Tags      admin-media
// @Security  BearerAuth
// @Produce   json
// @Param     id path int true "Media Asset ID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/media/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FromAppError(c, apperror.BadRequest("Invalid ID"))
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, nil, "Media asset deleted successfully")
}

// List godoc
// @Summary   List media assets (for the media library picker)
// @Tags      admin-media
// @Security  BearerAuth
// @Produce   json
// @Param     page query int false "page"
// @Param     per_page query int false "per_page"
// @Param     module query string false "filter by related_module"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/media [get]
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("per_page"))
	module := c.Query("module")

	items, meta, err := h.service.List(page, perPage, module)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.SuccessWithMeta(c, items, "OK", response.Meta{
		Page: meta.Page, PerPage: meta.PerPage, TotalData: meta.TotalData, TotalPages: meta.TotalPages,
	})
}
