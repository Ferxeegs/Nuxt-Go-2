package domain

type Permission struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;unique;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
}

func (Permission) TableName() string { return "permissions" }

type RolePermission struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	RoleID       uint `gorm:"not null;index" json:"role_id"`
	PermissionID uint `gorm:"not null;index" json:"permission_id"`
}

func (RolePermission) TableName() string { return "role_permissions" }
