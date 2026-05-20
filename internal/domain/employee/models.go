package employee

import (
	"amos-backend/internal/domain/auth"
	"amos-backend/internal/domain/master"
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

	// Relasi Master & Auth (Lazy Loaded)
	User            *auth.User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Position        *master.Position   `gorm:"foreignKey:PositionID" json:"position,omitempty"`
	Department      *master.Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	JobSite         *master.JobSite    `gorm:"foreignKey:JobSiteID" json:"job_site,omitempty"`
	ContractHistory []ContractHistory  `gorm:"foreignKey:EmployeeID" json:"contract_history,omitempty"`
}

type EmployeeDetail struct {
	EmployeeID         uint      `gorm:"primaryKey" json:"employee_id"`
	NIK                string    `gorm:"type:varchar(16)" json:"nik"`
	BirthPlace         string    `gorm:"type:varchar(100)" json:"birth_place"`
	BirthDate          time.Time `gorm:"type:date" json:"birth_date"`
	Religion           string    `gorm:"type:varchar(50)" json:"religion"`
	BloodType          string    `gorm:"type:varchar(5)" json:"blood_type"`
	MaritalStatus      string    `gorm:"type:varchar(50)" json:"marital_status"`
	MarriageDate       *time.Time `gorm:"type:date" json:"marriage_date"` // Pointer karena bisa null jika belum menikah
	
	// Fisik & Seragam
	Height             int       `json:"height"`
	Weight             int       `json:"weight"`
	ShirtSize          string    `gorm:"type:varchar(10)" json:"shirt_size"`
	ShoeSize           string    `gorm:"type:varchar(10)" json:"shoe_size"`
	PantsSize          string    `gorm:"type:varchar(10)" json:"pants_size"`
	
	// Kontak & Alamat
	AddressKTP         string    `gorm:"type:text" json:"address_ktp"`
	AddressDomicile    string    `gorm:"type:text" json:"address_domicile"`
	PhoneNumber        string    `gorm:"type:varchar(20)" json:"phone_number"`
	PersonalEmail      string    `gorm:"type:varchar(255)" json:"personal_email"`
	
	// Identitas Negara & Finansial
	NPWPNumber         string    `gorm:"type:varchar(30)" json:"npwp_number"`
	BPJSKesehatan      string    `gorm:"type:varchar(30)" json:"bpjs_kesehatan"`
	BPJSTenagaKerja    string    `gorm:"type:varchar(30)" json:"bpjs_tenaga_kerja"`
	BankName           string    `gorm:"type:varchar(50)" json:"bank_name"`
	BankAccount        string    `gorm:"type:varchar(30)" json:"bank_account"`
	BankAccountName    string    `gorm:"type:varchar(100)" json:"bank_account_name"`
	
	// Keluarga & Darurat
	MotherName         string    `gorm:"type:varchar(100)" json:"mother_name"`
	EmergencyName      string    `gorm:"type:varchar(100)" json:"emergency_name"`
	EmergencyRelation  string    `gorm:"type:varchar(50)" json:"emergency_relation"`
	EmergencyPhone     string    `gorm:"type:varchar(20)" json:"emergency_phone"`
	
	UpdatedAt          time.Time `json:"updated_at"`
}

type ContractHistory struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	EmployeeID     uint      `gorm:"not null" json:"employee_id"`
	ContractTypeID *uint     `json:"contract_type_id"`
	DecreeNumber   string    `gorm:"type:varchar(100)" json:"decree_number"`
	StartDate      time.Time `gorm:"type:date" json:"start_date"`
	EndDate        time.Time             `gorm:"type:date" json:"end_date"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`

	ContractType *master.ContractType `gorm:"foreignKey:ContractTypeID" json:"contract_type,omitempty"`
}
