package domain

import "time"

// MediaAsset records every uploaded file so orphaned files can be found and
// cleaned up later, and uploads never "pile up" without a reference.
type MediaAsset struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	FileName      string    `gorm:"size:255;not null" json:"file_name"`
	Path          string    `gorm:"size:500;not null" json:"path"`
	MimeType      string    `gorm:"size:100" json:"mime_type"`
	Size          int64     `json:"size"`
	UploadedBy    *string   `gorm:"type:char(36);index" json:"uploaded_by,omitempty"`
	RelatedModule string    `gorm:"size:50" json:"related_module"`
	RelatedID     *uint     `json:"related_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func (MediaAsset) TableName() string { return "media_assets" }
