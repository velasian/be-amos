package employee

import (
	"errors"
	"time"
)

type Service interface {
	CreateEmployee(input CreateEmployeeInput) (*Employee, error)
	GetEmployeeByID(id uint) (*Employee, error)
	GetEmployeeByUserID(userID uint) (*Employee, error)
	UpdateEmployee(id uint, input UpdateEmployeeInput) (*Employee, error)
	DeleteEmployee(id uint) error
	GetAllEmployeesPaginated(page, limit int, search string) (*PaginatedResult, error)
	UpdateBiodata(pegawaiID uint, input UpdateBiodataInput) (*Employee, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Helper
func uintPtrOrNil(id uint) *uint {
	if id == 0 {
		return nil
	}
	return &id
}
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type CreateEmployeeInput struct {
	UserID       uint   `json:"user_id"`
	NRP          string `json:"nrp" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Gender       string `json:"gender"`
	PositionID   uint   `json:"position_id"`
	DepartmentID uint   `json:"department_id"`
	JobSiteID    uint   `json:"job_site_id"`
	Status       string `json:"status"`
	JoinDate     string `json:"join_date"` // YYYY-MM-DD
	NfcUID       string `json:"nfc_uid"`   // NEW: For TA
}

type UpdateEmployeeInput struct {
	Name         string `json:"name"`
	NRP          string `json:"nrp"`
	Gender       string `json:"gender"`
	PositionID   uint   `json:"position_id"`
	DepartmentID uint   `json:"department_id"`
	JobSiteID    uint   `json:"job_site_id"`
	Status       string `json:"status"`
	NfcUID       string `json:"nfc_uid"`

	// Detail Fields
	NIK             string `json:"nik"`
	BirthPlace      string `json:"birth_place"`
	BirthDate       string `json:"birth_date"` // YYYY-MM-DD
	Religion        string `json:"religion"`
	BloodType       string `json:"blood_type"`
	MaritalStatus   string `json:"marital_status"`
	AddressDomicile string `json:"address_domicile"`
	PhoneNumber     string `json:"phone_number"`
	NPWPNumber      string `json:"npwp_number"`
}

func (s *service) CreateEmployee(input CreateEmployeeInput) (*Employee, error) {
	existing, _ := s.repo.FindByNRP(input.NRP)
	if existing != nil && existing.ID != 0 {
		return nil, errors.New("NRP already exists")
	}

	if input.NfcUID != "" {
		existingNFC, _ := s.repo.FindByNfcUID(input.NfcUID)
		if existingNFC != nil && existingNFC.ID != 0 {
			return nil, errors.New("NFC UID already assigned to another employee")
		}
	}

	e := &Employee{
		UserID:       uintPtrOrNil(input.UserID),
		NRP:          input.NRP,
		Name:         input.Name,
		Gender:       input.Gender,
		PositionID:   uintPtrOrNil(input.PositionID),
		DepartmentID: uintPtrOrNil(input.DepartmentID),
		JobSiteID:    uintPtrOrNil(input.JobSiteID),
		Status:       input.Status,
		NfcUID:       stringPtrOrNil(input.NfcUID),
	}

	if e.Status == "" {
		e.Status = "AKTIF"
	}

	if input.JoinDate != "" {
		t, err := time.Parse("2006-01-02", input.JoinDate)
		if err == nil {
			e.JoinDate = t
		}
	}

	if err := s.repo.CreateEmployee(e); err != nil {
		return nil, err
	}

	detail := &EmployeeDetail{EmployeeID: e.ID}
	_ = s.repo.SaveDetail(detail)
	e.Detail = detail

	return e, nil
}

func (s *service) GetEmployeeByID(id uint) (*Employee, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if e.Detail == nil {
		detail := &EmployeeDetail{EmployeeID: e.ID}
		_ = s.repo.SaveDetail(detail)
		e.Detail = detail
	}
	return e, nil
}

func (s *service) GetEmployeeByUserID(userID uint) (*Employee, error) {
	e, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if e.Detail == nil {
		detail := &EmployeeDetail{EmployeeID: e.ID}
		_ = s.repo.SaveDetail(detail)
		e.Detail = detail
	}
	return e, nil
}

type PaginatedResult struct {
	Data       []Employee `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
}

func (s *service) GetAllEmployeesPaginated(page, limit int, search string) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	list, total, err := s.repo.GetAllEmployeesPaginated(page, limit, search)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &PaginatedResult{
		Data:       list,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *service) UpdateEmployee(id uint, input UpdateEmployeeInput) (*Employee, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		e.Name = input.Name
	}
	if input.NRP != "" && e.NRP != input.NRP {
		existing, _ := s.repo.FindByNRP(input.NRP)
		if existing != nil && existing.ID != 0 {
			return nil, errors.New("NRP already exists")
		}
		e.NRP = input.NRP
	}
	if input.NfcUID != "" && (e.NfcUID == nil || *e.NfcUID != input.NfcUID) {
		existingNFC, _ := s.repo.FindByNfcUID(input.NfcUID)
		if existingNFC != nil && existingNFC.ID != 0 {
			return nil, errors.New("NFC UID already assigned to another employee")
		}
		e.NfcUID = &input.NfcUID
	}

	if input.PositionID > 0 {
		e.PositionID = &input.PositionID
	}
	if input.DepartmentID > 0 {
		e.DepartmentID = &input.DepartmentID
	}
	if input.JobSiteID > 0 {
		e.JobSiteID = &input.JobSiteID
	}
	if input.Status != "" {
		e.Status = input.Status
	}
	if input.Gender != "" {
		e.Gender = input.Gender
	}

	if err := s.repo.UpdateEmployee(e); err != nil {
		return nil, err
	}

	if e.Detail == nil {
		e.Detail = &EmployeeDetail{EmployeeID: e.ID}
	}
	e.Detail.NIK = input.NIK
	e.Detail.BirthPlace = input.BirthPlace
	if input.BirthDate != "" {
		t, err := time.Parse("2006-01-02", input.BirthDate)
		if err == nil {
			e.Detail.BirthDate = t
		}
	}
	e.Detail.Religion = input.Religion
	e.Detail.BloodType = input.BloodType
	e.Detail.MaritalStatus = input.MaritalStatus
	e.Detail.AddressDomicile = input.AddressDomicile
	e.Detail.PhoneNumber = input.PhoneNumber
	e.Detail.NPWPNumber = input.NPWPNumber

	if err := s.repo.SaveDetail(e.Detail); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

func (s *service) DeleteEmployee(id uint) error {
	return s.repo.DeleteEmployee(id)
}

// UpdateBiodataInput for Mobile app (self service)
type UpdateBiodataInput struct {
	NIK             string `json:"nik"`
	BirthPlace      string `json:"birth_place"`
	BirthDate       string `json:"birth_date"`
	Religion        string `json:"religion"`
	BloodType       string `json:"blood_type"`
	MaritalStatus   string `json:"marital_status"`
	AddressDomicile string `json:"address_domicile"`
	PhoneNumber     string `json:"phone_number"`
	NPWPNumber      string `json:"npwp_number"`
}

func (s *service) UpdateBiodata(employeeID uint, input UpdateBiodataInput) (*Employee, error) {
	e, err := s.repo.FindByID(employeeID)
	if err != nil {
		return nil, err
	}

	if e.Detail == nil {
		e.Detail = &EmployeeDetail{EmployeeID: e.ID}
	}
	e.Detail.NIK = input.NIK
	e.Detail.BirthPlace = input.BirthPlace
	if input.BirthDate != "" {
		t, err := time.Parse("2006-01-02", input.BirthDate)
		if err == nil {
			e.Detail.BirthDate = t
		}
	}
	e.Detail.Religion = input.Religion
	e.Detail.BloodType = input.BloodType
	e.Detail.MaritalStatus = input.MaritalStatus
	e.Detail.AddressDomicile = input.AddressDomicile
	e.Detail.PhoneNumber = input.PhoneNumber
	e.Detail.NPWPNumber = input.NPWPNumber

	if err := s.repo.SaveDetail(e.Detail); err != nil {
		return nil, err
	}

	return s.repo.FindByID(employeeID)
}
