package master

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Code      string         `gorm:"type:varchar(20)" json:"code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Position struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	DepartmentID *uint          `json:"department_id"`
	Department   *Department    `gorm:"foreignKey:DepartmentID;constraint:OnDelete:SET NULL" json:"department,omitempty"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type JobSite struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Latitude     float64        `gorm:"type:double precision" json:"latitude"`
	Longitude    float64        `gorm:"type:double precision" json:"longitude"`
	RadiusMeters int            `gorm:"default:100" json:"radius_meters"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type ContractType struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
