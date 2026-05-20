package employee

import (
	"bytes"
	"errors"
	"time"

	"github.com/xuri/excelize/v2"
)

type Service interface {
	CreateEmployee(input CreateEmployeeInput) (*Employee, error)
	GetEmployeeByID(id uint) (*Employee, error)
	GetEmployeeByUserID(userID uint) (*Employee, error)
	UpdateEmployee(id uint, input UpdateEmployeeInput) (*Employee, error)
	DeleteEmployee(id uint) error
	GetAllEmployeesPaginated(page, limit int, search string) (*PaginatedResult, error)
	UpdateBiodata(pegawaiID uint, input UpdateBiodataInput) (*Employee, error)
	ExportToExcel() (*bytes.Buffer, error)
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

func (s *service) ExportToExcel() (*bytes.Buffer, error) {
	list, err := s.repo.GetAllWithDetails()
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheetName := "Data Pegawai"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{
		"NRP*", "Full Name*", "Email", "Password", "Gender (M/F)",
		"Position", "Job Site",
		"Birth Place", "Birth Date (YYYY-MM-DD)", "Religion", "Blood Type",
		"Marital Status", "Domicile Address", "Phone Number", "NPWP Number",
		"Contract Type*", "Decree Number", "Contract Start* (YYYY-MM-DD)", "Contract End (YYYY-MM-DD)",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"1E3A5F"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	f.SetRowHeight(sheetName, 1, 30)

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	for i, p := range list {
		row := i + 2

		formatDateVal := func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02")
		}

		var noSK, tipeKontrak, tglMulai, tglSelesai string
		if len(p.ContractHistory) > 0 {
			latest := p.ContractHistory[0]
			noSK = latest.DecreeNumber
			if latest.ContractType != nil {
				tipeKontrak = latest.ContractType.Name
			}
			tglMulai = formatDateVal(latest.StartDate)
			tglSelesai = formatDateVal(latest.EndDate)
		}

		var d EmployeeDetail
		if p.Detail != nil {
			d = *p.Detail
		}

		posName := ""
		if p.Position != nil {
			posName = p.Position.Name
		}
		jsName := ""
		if p.JobSite != nil {
			jsName = p.JobSite.Name
		}
		
		email := ""
		if p.User != nil {
			email = p.User.Email
		}

		data := []interface{}{
			p.NRP,
			p.Name,
			email,
			"", // Password
			p.Gender,
			posName,
			jsName,
			d.BirthPlace,
			formatDateVal(d.BirthDate),
			d.Religion,
			d.BloodType,
			d.MaritalStatus,
			d.AddressDomicile,
			d.PhoneNumber,
			d.NPWPNumber,
			tipeKontrak,
			noSK,
			tglMulai,
			tglSelesai,
		}

		for j, val := range data {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheetName, cell, val)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
	}

	colWidths := map[string]float64{
		"A": 12, "B": 25, "C": 25, "D": 12, "E": 10,
		"F": 20, "G": 18,
		"H": 15, "I": 20, "J": 12, "K": 10, "L": 15, "M": 35, "N": 15, "O": 20,
		"P": 15, "Q": 20, "R": 22, "S": 22,
	}

	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	buffer := new(bytes.Buffer)
	if err := f.Write(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}
