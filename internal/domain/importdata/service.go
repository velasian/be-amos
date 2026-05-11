package importdata

import (
	"amos-backend/internal/domain/auth"
	"amos-backend/internal/domain/employee"
	"amos-backend/internal/domain/master"
	"amos-backend/pkg/utils"
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// Service defines the interface for import operations.
type Service interface {
	ProcessExcelToStaging(fileReader io.Reader) (string, int, error)
	ValidateStaging(batchID string) error
	GetStagingData(batchID string) ([]EmployeeStaging, error)
	UpdateStagingFields(id string, updates map[string]interface{}) error
	CommitStaging(batchID string) (int, int, error)
}

type service struct {
	stagingRepo  Repository
	employeeRepo employee.Repository
	masterRepo   master.Repository
	authRepo     auth.Repository
}

// NewService creates a new import service with all required dependencies.
func NewService(stagingRepo Repository, employeeRepo employee.Repository, masterRepo master.Repository, authRepo auth.Repository) Service {
	return &service{
		stagingRepo:  stagingRepo,
		employeeRepo: employeeRepo,
		masterRepo:   masterRepo,
		authRepo:     authRepo,
	}
}

// ProcessExcelToStaging reads an Excel file and inserts rows into the staging table.
func (s *service) ProcessExcelToStaging(fileReader io.Reader) (string, int, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(fileReader); err != nil {
		return "", 0, fmt.Errorf("failed to read file: %w", err)
	}

	xlsx, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", 0, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer xlsx.Close()

	sheets := xlsx.GetSheetList()
	if len(sheets) == 0 {
		return "", 0, fmt.Errorf("Excel file has no sheets")
	}

	rows, err := xlsx.GetRows(sheets[0])
	if err != nil {
		return "", 0, fmt.Errorf("failed to read rows: %w", err)
	}
	if len(rows) < 2 {
		return "", 0, fmt.Errorf("no data rows found (only header)")
	}

	batchID := uuid.New().String()
	var stagings []EmployeeStaging

	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}

		item := EmployeeStaging{
			ImportBatchID: batchID,
			// Core (0-4)
			NRP:      getString(row, 0),
			Name:     getString(row, 1),
			Email:    getString(row, 2),
			Password: getString(row, 3),
			Gender:   getString(row, 4),
			// Master Refs (5-6)
			PositionRaw: getString(row, 5),
			JobSiteRaw:  getString(row, 6),
			// Details (7-14)
			BirthPlace:      getString(row, 7),
			BirthDate:       getString(row, 8),
			Religion:        getString(row, 9),
			BloodType:       getString(row, 10),
			MaritalStatus:   getString(row, 11),
			AddressDomicile: getString(row, 12),
			PhoneNumber:     getString(row, 13),
			NPWPNumber:      getString(row, 14),
			// Contract (15-18)
			ContractTypeRaw: getString(row, 15),
			DecreeNumber:    getString(row, 16),
			ContractStart:   getString(row, 17),
			ContractEnd:     getString(row, 18),
		}

		// Auto-derive status & join date from contract
		item.EmployeeStatus = item.ContractTypeRaw
		item.JoinDate = item.ContractStart

		stagings = append(stagings, item)
	}

	if len(stagings) == 0 {
		return "", 0, fmt.Errorf("no valid data rows found")
	}

	if err := s.stagingRepo.CreateBatch(stagings); err != nil {
		return "", 0, fmt.Errorf("failed to insert staging data: %w", err)
	}

	return batchID, len(stagings), nil
}

// ValidateStaging validates all rows in a batch and marks them as VALID, UPDATE, or ERROR.
func (s *service) ValidateStaging(batchID string) error {
	items, err := s.stagingRepo.GetByBatchID(batchID)
	if err != nil {
		return err
	}

	for _, item := range items {
		status := "VALID"
		message := ""

		// Required fields check
		if item.NRP == "" || item.Name == "" {
			status = "ERROR"
			message += "NRP and Name are required. "
		}

		// Check if NRP already exists → mark as UPDATE
		if status != "ERROR" {
			existing, _ := s.employeeRepo.FindByNRP(item.NRP)
			if existing != nil && existing.ID != 0 {
				status = "UPDATE"
				message = "Employee exists, data will be updated."
			}
		}

		if err := s.stagingRepo.UpdateStatus(item.ID, status, message); err != nil {
			log.Printf("[IMPORT] Failed to update status for staging ID %d: %v", item.ID, err)
		}
	}

	return nil
}

func (s *service) GetStagingData(batchID string) ([]EmployeeStaging, error) {
	return s.stagingRepo.GetByBatchID(batchID)
}

func (s *service) UpdateStagingFields(id string, updates map[string]interface{}) error {
	return s.stagingRepo.UpdateFields(id, updates)
}

// CommitStaging processes all VALID/UPDATE rows and creates real records.
func (s *service) CommitStaging(batchID string) (int, int, error) {
	validItems, err := s.stagingRepo.GetValidByBatchID(batchID)
	if err != nil {
		return 0, 0, err
	}

	// Pre-load master data for name → ID resolution
	positions, _ := s.masterRepo.GetAllPositions()
	jobSites, _ := s.masterRepo.GetAllJobSites()
	contractTypes, _ := s.masterRepo.GetAllContractTypes()

	success := 0
	failed := 0

	for _, stg := range validItems {
		existing, _ := s.employeeRepo.FindByNRP(stg.NRP)
		isUpdate := existing != nil && existing.ID != 0

		if isUpdate {
			if err := s.updateExistingEmployee(existing, stg, positions, jobSites); err != nil {
				log.Printf("[IMPORT] Failed to update NRP %s: %v", stg.NRP, err)
				failed++
				_ = s.stagingRepo.UpdateStatus(stg.ID, "ERROR", "Update failed: "+err.Error())
				continue
			}
			success++
			_ = s.stagingRepo.DeleteStaging(stg.ID)
			continue
		}

		// CREATE new employee with user account
		if err := s.createNewEmployee(stg, positions, jobSites, contractTypes); err != nil {
			log.Printf("[IMPORT] Failed to create NRP %s: %v", stg.NRP, err)
			failed++
			_ = s.stagingRepo.UpdateStatus(stg.ID, "ERROR", "Create failed: "+err.Error())
			continue
		}

		success++
		_ = s.stagingRepo.DeleteStaging(stg.ID)
	}

	return success, failed, nil
}

// createNewEmployee creates a user account, employee, detail, and contract history.
func (s *service) createNewEmployee(stg EmployeeStaging, positions []master.Position, jobSites []master.JobSite, contractTypes []master.ContractType) error {
	nameUpper := strings.ToUpper(stg.Name)

	// Generate default password if empty
	password := stg.Password
	if password == "" {
		password = stg.NRP + "123"
	}

	// Generate email if empty
	email := strings.TrimSpace(stg.Email)
	if email == "" {
		email = strings.ToLower(stg.NRP) + "@amos.internal"
	}

	// Hash password and create user account
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	nrpPtr := &stg.NRP
	user := &auth.User{
		NRP:      nrpPtr,
		Email:    email,
		Password: hashedPassword,
		Role:     "employee",
	}

	if err := s.authRepo.CreateUser(user); err != nil {
		return fmt.Errorf("failed to create user account: %w", err)
	}

	log.Printf("[IMPORT] User created for NRP %s (UserID: %d, Email: %s)", stg.NRP, user.ID, email)

	// Build Employee
	emp := &employee.Employee{
		UserID: &user.ID,
		NRP:    stg.NRP,
		Name:   nameUpper,
		Gender: stg.Gender,
		Status: "AKTIF",
	}

	// Resolve Position → also auto-derive Department
	for _, pos := range positions {
		if strings.EqualFold(pos.Name, stg.PositionRaw) {
			emp.PositionID = &pos.ID
			if pos.DepartmentID != nil {
				emp.DepartmentID = pos.DepartmentID
			}
			break
		}
	}

	// Resolve JobSite
	for _, js := range jobSites {
		if strings.EqualFold(js.Name, stg.JobSiteRaw) {
			emp.JobSiteID = &js.ID
			break
		}
	}

	// Parse JoinDate
	if stg.JoinDate != "" {
		if t, err := time.Parse("2006-01-02", stg.JoinDate); err == nil {
			emp.JoinDate = t
		}
	}

	if err := s.employeeRepo.CreateEmployee(emp); err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}

	// Create EmployeeDetail
	detail := &employee.EmployeeDetail{
		EmployeeID:      emp.ID,
		NIK:             stg.BirthPlace, // Will be overwritten below
		BirthPlace:      stg.BirthPlace,
		Religion:        stg.Religion,
		BloodType:       stg.BloodType,
		MaritalStatus:   stg.MaritalStatus,
		AddressDomicile: stg.AddressDomicile,
		PhoneNumber:     stg.PhoneNumber,
		NPWPNumber:      stg.NPWPNumber,
	}

	if stg.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", stg.BirthDate); err == nil {
			detail.BirthDate = t
		}
	}

	if err := s.employeeRepo.SaveDetail(detail); err != nil {
		log.Printf("[IMPORT] Warning: failed to save detail for NRP %s: %v", stg.NRP, err)
	}

	// Create ContractHistory if data exists
	if stg.ContractTypeRaw != "" {
		s.createContractHistory(emp.ID, stg, contractTypes)
	}

	log.Printf("[IMPORT] ✓ Employee created: NRP %s (ID: %d)", stg.NRP, emp.ID)
	return nil
}

// updateExistingEmployee updates core and detail fields for an existing employee.
func (s *service) updateExistingEmployee(emp *employee.Employee, stg EmployeeStaging, positions []master.Position, jobSites []master.JobSite) error {
	emp.Name = strings.ToUpper(stg.Name)
	if stg.Gender != "" {
		emp.Gender = stg.Gender
	}

	// Resolve Position
	for _, pos := range positions {
		if strings.EqualFold(pos.Name, stg.PositionRaw) {
			emp.PositionID = &pos.ID
			if pos.DepartmentID != nil {
				emp.DepartmentID = pos.DepartmentID
			}
			break
		}
	}

	// Resolve JobSite
	for _, js := range jobSites {
		if strings.EqualFold(js.Name, stg.JobSiteRaw) {
			emp.JobSiteID = &js.ID
			break
		}
	}

	if err := s.employeeRepo.UpdateEmployee(emp); err != nil {
		return err
	}

	// Update detail
	detail := emp.Detail
	if detail == nil {
		detail = &employee.EmployeeDetail{EmployeeID: emp.ID}
	}

	detail.BirthPlace = stg.BirthPlace
	detail.Religion = stg.Religion
	detail.BloodType = stg.BloodType
	detail.MaritalStatus = stg.MaritalStatus
	detail.AddressDomicile = stg.AddressDomicile
	detail.PhoneNumber = stg.PhoneNumber
	detail.NPWPNumber = stg.NPWPNumber

	if stg.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", stg.BirthDate); err == nil {
			detail.BirthDate = t
		}
	}

	return s.employeeRepo.SaveDetail(detail)
}

// createContractHistory resolves contract type by name and creates a history record.
func (s *service) createContractHistory(employeeID uint, stg EmployeeStaging, contractTypes []master.ContractType) {
	var contractTypeID uint
	for _, ct := range contractTypes {
		if strings.EqualFold(ct.Name, stg.ContractTypeRaw) {
			contractTypeID = ct.ID
			break
		}
	}

	if contractTypeID == 0 {
		log.Printf("[IMPORT] Warning: Contract type '%s' not found in master data for NRP %s", stg.ContractTypeRaw, stg.NRP)
		return
	}

	contract := &employee.ContractHistory{
		EmployeeID:     employeeID,
		ContractTypeID: &contractTypeID,
		DecreeNumber:   stg.DecreeNumber,
	}

	if stg.ContractStart != "" {
		if t, err := time.Parse("2006-01-02", stg.ContractStart); err == nil {
			contract.StartDate = t
		}
	}
	if stg.ContractEnd != "" {
		if t, err := time.Parse("2006-01-02", stg.ContractEnd); err == nil {
			contract.EndDate = t
		}
	}

	if err := s.employeeRepo.SaveContractHistory(contract); err != nil {
		log.Printf("[IMPORT] Warning: failed to save contract for NRP %s: %v", stg.NRP, err)
	}
}

// getString safely gets a string value from a row slice.
func getString(row []string, index int) string {
	if index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}

// parseInt converts a string to int, returning 0 on failure.
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	val, _ := strconv.Atoi(s)
	return val
}
