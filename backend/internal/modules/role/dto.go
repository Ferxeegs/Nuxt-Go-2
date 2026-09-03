package role

type CreateRequest struct {
	Name          string `json:"name" binding:"required,min=2,max=50"`
	Description   string `json:"description" binding:"max=255"`
	PermissionIDs []uint `json:"permission_ids"`
}

type UpdateRequest struct {
	Name          string  `json:"name" binding:"omitempty,min=2,max=50"`
	Description   string  `json:"description" binding:"omitempty,max=255"`
	PermissionIDs *[]uint `json:"permission_ids"`
}

type Response struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// PermissionResponse is the full permission catalogue entry (with its real
// numeric id), used by admin UIs to build a role's permission picklist —
// role.Response itself only carries permission names, not ids.
type PermissionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
