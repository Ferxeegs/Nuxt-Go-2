package user

type AssignmentInput struct {
	RoleID  uint `json:"role_id" binding:"required"`
	ScopeID uint `json:"scope_id" binding:"required"`
}

type AssignmentResponse struct {
	RoleID    uint   `json:"role_id"`
	RoleName  string `json:"role_name"`
	ScopeID   uint   `json:"scope_id"`
	ScopeCode string `json:"scope_code"`
	ScopeName string `json:"scope_name"`
}

type RoleSummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CreateRequest struct {
	Name        string            `json:"name" binding:"required,min=2,max=150"`
	Email       string            `json:"email" binding:"required,email"`
	Password    string            `json:"password" binding:"required,min=6"`
	Assignments []AssignmentInput `json:"assignments" binding:"required,min=1,dive"`
}

type UpdateRequest struct {
	Name        string             `json:"name" binding:"omitempty,min=2,max=150"`
	Email       string             `json:"email" binding:"omitempty,email"`
	IsActive    *bool              `json:"is_active" binding:"omitempty"`
	Assignments *[]AssignmentInput `json:"assignments"`
}

type ListQuery struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Search  string `form:"search"`
}

type Response struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Email       string               `json:"email"`
	Roles       []RoleSummary        `json:"roles"`
	Avatar      string               `json:"avatar,omitempty"`
	IsActive    bool                 `json:"is_active"`
	Assignments []AssignmentResponse `json:"assignments"`
}
