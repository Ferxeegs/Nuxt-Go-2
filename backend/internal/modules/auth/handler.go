package auth

import (
	"time"

	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/middleware"
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

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password, returns an access and refresh token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Credentials"
// @Success      200 {object} response.Meta
// @Failure      401 {object} response.Meta
// @Router       /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}

	tokens, err := h.service.Login(c.Request.Context(), req.Email, req.Password, req.TurnstileToken, c.ClientIP())
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, tokens, "Login successful")
}

// Refresh godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest true "Refresh token"
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}

	tokens, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, tokens, "Token refreshed")
}

// Logout godoc
// @Summary      Logout
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	jti, _ := c.Get(middleware.ContextTokenJTI)
	expiry, _ := c.Get(middleware.ContextTokenExpiry)

	uid, ok := userID.(string)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	jtiStr, _ := jti.(string)
	expiresAt, ok := expiry.(time.Time)
	if !ok {
		expiresAt = time.Now()
	}

	if err := h.service.Logout(uid, jtiStr, expiresAt); err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, nil, "Logout successful")
}

// Me godoc
// @Summary      Current authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	uid, impersonatorID, activeRoleID, activeScopeID, ok := sessionContextFull(c)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	me, err := h.service.Me(uid, impersonatorID, activeRoleID, activeScopeID)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, me, "OK")
}

func sessionContext(c *gin.Context) (userID string, impersonatorID *string, ok bool) {
	uid, impersonatorID, _, _, ok := sessionContextFull(c)
	return uid, impersonatorID, ok
}

func sessionContextFull(c *gin.Context) (userID string, impersonatorID *string, activeRoleID, activeScopeID uint, ok bool) {
	rawUserID, exists := c.Get(middleware.ContextUserID)
	if !exists {
		return "", nil, 0, 0, false
	}
	uid, ok := rawUserID.(string)
	if !ok {
		return "", nil, 0, 0, false
	}

	if raw, exists := c.Get(middleware.ContextImpersonatorID); exists {
		if id, ok := raw.(string); ok {
			impersonatorID = &id
		}
	}
	if raw, exists := c.Get(middleware.ContextActiveRoleID); exists {
		switch v := raw.(type) {
		case uint:
			activeRoleID = v
		case int:
			activeRoleID = uint(v)
		case float64:
			activeRoleID = uint(v)
		}
	}
	if raw, exists := c.Get(middleware.ContextActiveScopeID); exists {
		switch v := raw.(type) {
		case uint:
			activeScopeID = v
		case int:
			activeScopeID = uint(v)
		case float64:
			activeScopeID = uint(v)
		}
	}
	return uid, impersonatorID, activeRoleID, activeScopeID, true
}

// UpdateProfile godoc
// @Summary      Update current user profile
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body UpdateProfileRequest true "Profile fields"
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/me [patch]
func (h *Handler) UpdateProfile(c *gin.Context) {
	uid, impersonatorID, activeRoleID, activeScopeID, ok := sessionContextFull(c)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}

	me, err := h.service.UpdateProfile(uid, impersonatorID, activeRoleID, activeScopeID, req)
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, me, "Profile updated")
}

// SwitchRole godoc
// @Summary      Switch active role assignment
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body SwitchRoleRequest true "Role and scope"
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/switch-role [post]
func (h *Handler) SwitchRole(c *gin.Context) {
	uid, impersonatorID, ok := sessionContext(c)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	var req SwitchRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}

	jti, _ := c.Get(middleware.ContextTokenJTI)
	jtiStr, _ := jti.(string)
	expiry, _ := c.Get(middleware.ContextTokenExpiry)
	expiresAt, ok := expiry.(time.Time)
	if !ok {
		expiresAt = time.Now()
	}

	tokens, err := h.service.SwitchRole(uid, impersonatorID, req.RoleID, req.ScopeID, jtiStr, expiresAt, c.ClientIP())
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, tokens, "Active role updated")
}

// ChangePassword godoc
// @Summary      Change current user password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body ChangePasswordRequest true "Password change"
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/me/password [patch]
func (h *Handler) ChangePassword(c *gin.Context) {
	uid, impersonatorID, ok := sessionContext(c)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}

	if err := h.service.ChangePassword(uid, impersonatorID, req); err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, nil, "Password updated")
}

// DeleteProfile godoc
// @Summary      Deactivate current user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body DeleteProfileRequest true "Confirm password"
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/me [delete]
func (h *Handler) DeleteProfile(c *gin.Context) {
	uid, impersonatorID, ok := sessionContext(c)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	var req DeleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}

	if err := h.service.DeleteProfile(uid, impersonatorID, req); err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, nil, "Account deactivated")
}

// Impersonate godoc
// @Summary      Start impersonating another user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body ImpersonateRequest true "Target user"
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/impersonate [post]
func (h *Handler) Impersonate(c *gin.Context) {
	if _, exists := c.Get(middleware.ContextImpersonatorID); exists {
		response.FromAppError(c, apperror.BadRequest("Already impersonating another user; stop first"))
		return
	}

	var req ImpersonateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Validation failed", validator.TranslateErrors(err))
		return
	}

	actorID, ok := c.Get(middleware.ContextUserID)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}
	uid, ok := actorID.(string)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	jti, _ := c.Get(middleware.ContextTokenJTI)
	jtiStr, _ := jti.(string)
	expiry, _ := c.Get(middleware.ContextTokenExpiry)
	expiresAt, ok := expiry.(time.Time)
	if !ok {
		expiresAt = time.Now()
	}

	tokens, err := h.service.Impersonate(uid, req.UserID, jtiStr, expiresAt, c.ClientIP())
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, tokens, "Impersonation started")
}

// StopImpersonate godoc
// @Summary      Stop impersonation and restore the original admin session
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Meta
// @Router       /api/v1/auth/stop-impersonate [post]
func (h *Handler) StopImpersonate(c *gin.Context) {
	rawActor, exists := c.Get(middleware.ContextImpersonatorID)
	if !exists {
		response.FromAppError(c, apperror.BadRequest("Not currently impersonating"))
		return
	}
	impersonatorID, ok := rawActor.(string)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	targetID, ok := c.Get(middleware.ContextUserID)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}
	uid, ok := targetID.(string)
	if !ok {
		response.FromAppError(c, apperror.Unauthorized("Invalid session"))
		return
	}

	jti, _ := c.Get(middleware.ContextTokenJTI)
	jtiStr, _ := jti.(string)
	expiry, _ := c.Get(middleware.ContextTokenExpiry)
	expiresAt, ok := expiry.(time.Time)
	if !ok {
		expiresAt = time.Now()
	}

	tokens, err := h.service.StopImpersonate(uid, impersonatorID, jtiStr, expiresAt, c.ClientIP())
	if err != nil {
		response.FromAppError(c, err)
		return
	}
	response.Success(c, tokens, "Impersonation stopped")
}
