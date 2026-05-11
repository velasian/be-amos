package payslip

import (
	"time"
)

type Payslip struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EmployeeID  uint      `gorm:"not null" json:"employee_id"`
	Month       int       `gorm:"not null" json:"month"`
	Year        int       `gorm:"not null" json:"year"`
	FileURL     string    `gorm:"type:text;not null" json:"file_url"`
	IsPublished bool      `gorm:"default:false" json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
