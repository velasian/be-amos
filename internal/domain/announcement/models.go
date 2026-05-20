package announcement

import (
	"time"

	"gorm.io/gorm"
)

type Announcement struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"not null" json:"title"`
	Content   string         `gorm:"type:text" json:"content"`
	FileURL   string         `json:"file_url"`
	FileType  string         `json:"file_type"` // "image" or "pdf"
	ExpiresAt *time.Time     `json:"expires_at"`
	CreatedBy string         `json:"created_by"` // Username or Email
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
