package system

import (
	"time"
)

type File struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EntityType string    `gorm:"type:varchar(50)" json:"entity_type"` // PEGAWAI, KONTRAK, CUTI
	EntityID   uint      `json:"entity_id"`
	FilePath   string    `gorm:"type:text;not null" json:"file_path"`
	FileName   string    `gorm:"type:varchar(255)" json:"file_name"`
	FileType   string    `gorm:"type:varchar(50)" json:"file_type"`
	Size       int       `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
}
