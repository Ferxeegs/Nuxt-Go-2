package activitylog

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List godoc
// @Summary   List activity logs
// @Tags      admin-activity-logs
// @Security  BearerAuth
// @Produce   json
// @Param     page query int false "page"
// @Param     per_page query int false "per_page"
// @Param     module query string false "module filter"
// @Param     action query string false "action filter"
// @Param     user_id query string false "user id filter"
// @Param     q query string false "search"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/activity-logs [get]
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
// @Summary   Get activity log detail
// @Tags      admin-activity-logs
// @Security  BearerAuth
// @Produce   json
// @Param     id path int true "Activity Log ID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/activity-logs/{id} [get]
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

// Delete godoc
// @Summary   Delete an activity log
// @Tags      admin-activity-logs
// @Security  BearerAuth
// @Produce   json
// @Param     id path int true "Activity Log ID"
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/activity-logs/{id} [delete]
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
	response.Success(c, nil, "Activity log deleted successfully")
}

// ListFilters godoc
// @Summary   List distinct modules and actions for filters
// @Tags      admin-activity-logs
// @Security  BearerAuth
// @Produce   json
// @Success   200 {object} response.Meta
// @Router    /api/v1/admin/activity-logs/filters [get]
func (h *Handler) ListFilters(c *gin.Context) {
	modules, err := h.service.ListModules()
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	actions, err := h.service.ListActions()
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, gin.H{"modules": modules, "actions": actions}, "OK")
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, apperror.BadRequest("Invalid ID")
	}
	return uint(id), nil
}
