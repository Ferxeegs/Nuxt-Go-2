package domain

import "time"

// Scope is an organizational boundary (branch, unit, tenant) used to
// constrain role assignments in scoped RBAC.
type Scope struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Code        string     `gorm:"size:50;unique;not null" json:"code"`
	Name        string     `gorm:"size:150;not null" json:"name"`
	Description string     `gorm:"size:255" json:"description"`
	ParentID    *uint      `gorm:"index" json:"parent_id,omitempty"`
	Parent      *Scope     `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Scope) TableName() string { return "scopes" }

const ScopeCodeGlobal = "global"
