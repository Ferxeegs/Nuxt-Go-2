package domain

import "time"

// WebSetting is a flexible key-value store for site configuration
// (site_name, logo, contact_email, ...) so new fields don't require a
// migration.
type WebSetting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"size:100;unique;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Type      string    `gorm:"size:20;default:string" json:"type"` // string/text/json/image
	UpdatedAt time.Time `json:"updated_at"`
}

func (WebSetting) TableName() string { return "web_settings" }
