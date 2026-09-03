package activitylog

type ListQuery struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Module  string `form:"module"`
	Action  string `form:"action"`
	UserID  string `form:"user_id"`
	Q       string `form:"q"`
}

type Response struct {
	ID          uint    `json:"id"`
	UserID      *string `json:"user_id,omitempty"`
	UserName    string  `json:"user_name,omitempty"`
	UserEmail   string  `json:"user_email,omitempty"`
	Action      string  `json:"action"`
	Module      string  `json:"module"`
	EntityID    *string `json:"entity_id,omitempty"`
	Description string  `json:"description"`
	IPAddress   string  `json:"ip_address"`
	CreatedAt   string  `json:"created_at"`
}

type Filter struct {
	Module string
	Action string
	UserID string
	Q      string
}
