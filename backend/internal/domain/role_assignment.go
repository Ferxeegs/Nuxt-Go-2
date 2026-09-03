package domain

import "time"

// RoleAssignment grants a role to a user within a specific organizational scope.
// Global access is represented by assigning a role on the seeded "global" scope.
type RoleAssignment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"type:char(36);not null;index;uniqueIndex:ux_user_role_scope" json:"user_id"`
	RoleID    uint      `gorm:"not null;index;uniqueIndex:ux_user_role_scope" json:"role_id"`
	ScopeID   uint      `gorm:"not null;index;uniqueIndex:ux_user_role_scope" json:"scope_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User  *User  `json:"user,omitempty"`
	Role  *Role  `json:"role,omitempty"`
	Scope *Scope `json:"scope,omitempty"`
}

func (RoleAssignment) TableName() string { return "role_assignments" }
