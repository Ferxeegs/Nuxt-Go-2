package auth

type LoginRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	TurnstileToken string `json:"turnstile_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ImpersonateRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

type SwitchRoleRequest struct {
	RoleID  uint `json:"role_id" binding:"required"`
	ScopeID uint `json:"scope_id" binding:"required"`
}

type TokenResponse struct {
	AccessToken            string `json:"access_token"`
	RefreshToken           string `json:"refresh_token"`
	TokenType              string `json:"token_type"`
	ExpiresAt              string `json:"expires_at"`
	RequiresRoleSelection  bool   `json:"requires_role_selection"`
	ActiveRoleID           uint   `json:"active_role_id,omitempty"`
	ActiveScopeID          uint   `json:"active_scope_id,omitempty"`
}

type RoleSummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type MeResponse struct {
	ID                    string               `json:"id"`
	Name                  string               `json:"name"`
	Email                 string               `json:"email"`
	Role                  string               `json:"role"`
	RoleID                uint                 `json:"role_id,omitempty"`
	Roles                 []RoleSummary        `json:"roles"`
	Avatar                string               `json:"avatar,omitempty"`
	IsActive              bool                 `json:"is_active"`
	CreatedAt             string               `json:"created_at"`
	LastLoginAt           string               `json:"last_login_at,omitempty"`
	ImpersonatorID        *string              `json:"impersonator_id,omitempty"`
	Permissions           []string             `json:"permissions"`
	Assignments           []AssignmentResponse `json:"assignments"`
	ActiveRoleID          uint                 `json:"active_role_id,omitempty"`
	ActiveScopeID         uint                 `json:"active_scope_id,omitempty"`
	ActiveRoleName        string               `json:"active_role_name,omitempty"`
	ActiveScopeCode       string               `json:"active_scope_code,omitempty"`
	ActiveScopeName       string               `json:"active_scope_name,omitempty"`
	RequiresRoleSelection bool                 `json:"requires_role_selection"`
	CanSwitchRole         bool                 `json:"can_switch_role"`
}

type AssignmentResponse struct {
	RoleID    uint   `json:"role_id"`
	RoleName  string `json:"role_name"`
	ScopeID   uint   `json:"scope_id"`
	ScopeCode string `json:"scope_code"`
	ScopeName string `json:"scope_name"`
}

type UpdateProfileRequest struct {
	Name   string `json:"name" binding:"omitempty,min=2,max=150"`
	Email  string `json:"email" binding:"omitempty,email"`
	Avatar string `json:"avatar" binding:"omitempty,max=255"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type DeleteProfileRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}
