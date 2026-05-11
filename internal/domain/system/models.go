package system

import (
	"time"

	"gorm.io/gorm"
)

// File represents a polymorphic file record in the database.
// It tracks metadata for files stored in MinIO object storage.
type File struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	EntityType string         `gorm:"type:varchar(50);index" json:"entity_type"` // employee, contract, leave
	EntityID   uint           `gorm:"index" json:"entity_id"`
	Category   string         `gorm:"type:varchar(50);index" json:"category"` // profile_photo, document, ktp, sk_scan
	FilePath   string         `gorm:"type:text;not null" json:"file_path"`    // MinIO object name
	FileName   string         `gorm:"type:varchar(255)" json:"file_name"`     // Original filename
	FileType   string         `gorm:"type:varchar(20)" json:"file_type"`      // Extension (.jpg, .pdf)
	MimeType   string         `gorm:"type:varchar(100)" json:"mime_type"`     // MIME type
	Size       int64          `json:"size"`                                   // File size in bytes
	PublicURL  string         `gorm:"type:text" json:"public_url,omitempty"`  // Cached public URL
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
