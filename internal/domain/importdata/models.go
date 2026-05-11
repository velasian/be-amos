package importdata

import (
	"time"
)

// EmployeeStaging acts as a temporary buffer for imported Excel data.
// All fields are strings to safely hold raw values before validation and parsing.
type EmployeeStaging struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	ImportBatchID string `gorm:"index;type:varchar(50)" json:"import_batch_id"`
	Status        string `gorm:"index;type:varchar(20);default:'PENDING'" json:"status"` // PENDING, VALID, UPDATE, ERROR
	ErrorMessage  string `gorm:"type:text" json:"error_message"`

	// Core Employee Data
	NRP      string `gorm:"type:varchar(50)" json:"nrp"`
	Name     string `gorm:"type:varchar(255)" json:"name"`
	Email    string `gorm:"type:varchar(100)" json:"email"`
	Password string `gorm:"type:varchar(100)" json:"-"`
	Gender   string `gorm:"type:varchar(10)" json:"gender"`

	// Master Data References (raw strings for lookup)
	PositionRaw string `gorm:"type:varchar(100)" json:"position_raw"`
	JobSiteRaw  string `gorm:"type:varchar(100)" json:"job_site_raw"`

	// Employee Details
	BirthPlace      string `gorm:"type:varchar(100)" json:"birth_place"`
	BirthDate       string `gorm:"type:varchar(20)" json:"birth_date"`
	Religion        string `gorm:"type:varchar(50)" json:"religion"`
	BloodType       string `gorm:"type:varchar(5)" json:"blood_type"`
	MaritalStatus   string `gorm:"type:varchar(50)" json:"marital_status"`
	AddressDomicile string `gorm:"type:text" json:"address_domicile"`
	PhoneNumber     string `gorm:"type:varchar(20)" json:"phone_number"`
	NPWPNumber      string `gorm:"type:varchar(30)" json:"npwp_number"`

	// Contract Data
	ContractTypeRaw  string `gorm:"type:varchar(50)" json:"contract_type_raw"`
	DecreeNumber     string `gorm:"type:varchar(100)" json:"decree_number"`
	ContractStart    string `gorm:"type:varchar(20)" json:"contract_start"`
	ContractEnd      string `gorm:"type:varchar(20)" json:"contract_end"`
	EmployeeStatus   string `gorm:"type:varchar(50)" json:"employee_status"` // Auto-derived from contract
	JoinDate         string `gorm:"type:varchar(20)" json:"join_date"`       // Auto-derived from contract start

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
