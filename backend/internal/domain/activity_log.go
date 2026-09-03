package domain

import "time"

// ActivityLog records who did what and when, for accountability in a
// multi-admin CMS.
type ActivityLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      *string   `gorm:"type:char(36);index" json:"user_id,omitempty"`
	User        *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"user,omitempty"`
	Action      string    `gorm:"size:100;not null" json:"action"`
	Module      string    `gorm:"size:50;not null" json:"module"`
	EntityID    *string   `gorm:"size:64" json:"entity_id,omitempty"`
	Description string    `gorm:"size:500" json:"description"`
	IPAddress   string    `gorm:"size:64" json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ActivityLog) TableName() string { return "activity_logs" }
