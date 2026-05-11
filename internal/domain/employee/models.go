package employee

import (
	"time"
)

type Employee struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       *uint     `gorm:"uniqueIndex" json:"user_id"`
	PositionID   *uint     `json:"position_id"`
	DepartmentID *uint     `json:"department_id"`
	JobSiteID    *uint     `json:"job_site_id"`
	NRP          string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"nrp"`
	NfcUID       *string   `gorm:"type:varchar(100);uniqueIndex" json:"nfc_uid"`
	Name         string    `gorm:"type:varchar(255)" json:"name"`
	Gender       string    `gorm:"type:char(1)" json:"gender"`
	Status       string    `gorm:"type:varchar(50)" json:"status"` // AKTIF, RESIGN, TERMINATED
	JoinDate     time.Time `gorm:"type:date" json:"join_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relasi Detail Pegawai (Lazy Loaded)
	Detail *EmployeeDetail `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE" json:"detail,omitempty"`
}

type EmployeeDetail struct {
	EmployeeID      uint      `gorm:"primaryKey" json:"employee_id"` // Menggunakan ID Pegawai sebagai Primary Key
	NIK             string    `gorm:"type:varchar(16)" json:"nik"`
	BirthPlace      string    `gorm:"type:varchar(100)" json:"birth_place"`
	BirthDate       time.Time `gorm:"type:date" json:"birth_date"`
	Religion        string    `gorm:"type:varchar(50)" json:"religion"`
	BloodType       string    `gorm:"type:varchar(5)" json:"blood_type"`
	MaritalStatus   string    `gorm:"type:varchar(50)" json:"marital_status"`
	AddressDomicile string    `gorm:"type:text" json:"address_domicile"`
	PhoneNumber     string    `gorm:"type:varchar(20)" json:"phone_number"`
	NPWPNumber      string    `gorm:"type:varchar(30)" json:"npwp_number"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ContractHistory struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	EmployeeID     uint      `gorm:"not null" json:"employee_id"`
	ContractTypeID *uint     `json:"contract_type_id"`
	DecreeNumber   string    `gorm:"type:varchar(100)" json:"decree_number"`
	StartDate      time.Time `gorm:"type:date" json:"start_date"`
	EndDate        time.Time `gorm:"type:date" json:"end_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
