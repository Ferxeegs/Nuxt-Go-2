package domain

import "time"

// RefreshToken stores a hashed refresh token per user so tokens can be
// rotated (old one invalidated on every refresh) and revoked on logout,
// preventing replay attacks.
type RefreshToken struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         string    `gorm:"type:char(36);not null;index" json:"user_id"`
	TokenHash      string    `gorm:"size:255;not null;index" json:"-"`
	ImpersonatorID *string   `gorm:"type:char(36);index" json:"impersonator_id,omitempty"`
	ActiveRoleID   *uint     `gorm:"index" json:"active_role_id,omitempty"`
	ActiveScopeID  *uint     `gorm:"index" json:"active_scope_id,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	Revoked        bool      `gorm:"default:false" json:"revoked"`
	CreatedAt      time.Time `json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
