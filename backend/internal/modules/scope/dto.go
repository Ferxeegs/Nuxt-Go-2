package scope

type CreateRequest struct {
	Code        string `json:"code" binding:"required,min=2,max=50"`
	Name        string `json:"name" binding:"required,min=2,max=150"`
	Description string `json:"description" binding:"max=255"`
	ParentID    *uint  `json:"parent_id"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateRequest struct {
	Code        string `json:"code" binding:"omitempty,min=2,max=50"`
	Name        string `json:"name" binding:"omitempty,min=2,max=150"`
	Description string `json:"description" binding:"omitempty,max=255"`
	ParentID    *uint  `json:"parent_id"`
	ClearParent bool   `json:"clear_parent"`
	IsActive    *bool  `json:"is_active"`
}

type ListQuery struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Search  string `form:"search"`
}

type Response struct {
	ID           uint   `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ParentID     *uint  `json:"parent_id,omitempty"`
	ParentCode   string `json:"parent_code,omitempty"`
	ParentName   string `json:"parent_name,omitempty"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at,omitempty"`
}
