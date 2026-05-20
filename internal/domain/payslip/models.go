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

type PayslipStaging struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ImportBatchID string    `gorm:"index;size:50" json:"import_batch_id"`
	Status        string    `gorm:"index;size:20;default:'PENDING'" json:"status"` // PENDING, VALID, ERROR, DUPLICATE
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`
	
	// Source Info
	Filename string `gorm:"size:255" json:"filename"`
	NRPRaw   string `gorm:"size:100" json:"nrp_raw"`
	
	// Resolved Data
	EmployeeID   uint   `json:"employee_id"`
	EmployeeNRP  string `json:"employee_nrp"`
	EmployeeName string `json:"employee_name"`
	
	Month int `json:"month"`
	Year  int `json:"year"`
	
	// Temp File Storage
	FilePath    string    `gorm:"size:500" json:"file_path"`
	ContentType string    `gorm:"size:100" json:"content_type"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
