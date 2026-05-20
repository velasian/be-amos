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

func NewService(stagingRepo Repository, employeeRepo employee.Repository, masterRepo master.Repository, authRepo auth.Repository) Service {
	return &service{
		stagingRepo:  stagingRepo,
		employeeRepo: employeeRepo,
		masterRepo:   masterRepo,
		authRepo:     authRepo,
	}
}

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
			// Core Data (0-3)
			NRP:      getString(row, 0),
			Nama:     getString(row, 1),
			Email:    getString(row, 2),
			Password: getString(row, 3),
			// Master Data (4-5)
			JabatanRaw:    getString(row, 4),
			DepartemenRaw: "", // Auto-derived from Jabatan.DepartmentID
			JobSiteRaw:    getString(row, 5),
			// Personal Data (6-14)
			TempatLahir:       getString(row, 6),
			TanggalLahir:      getString(row, 7),
			JenisKelamin:      getString(row, 8),
			Agama:             getString(row, 9),
			StatusPernikahan:  getString(row, 10),
			TanggalPernikahan: getString(row, 11),
			GolonganDarah:     getString(row, 12),
			TinggiBadan:       getString(row, 13),
			BeratBadan:        getString(row, 14),
			// Identity (15-17)
			NIK:            getString(row, 15),
			AlamatKTP:      getString(row, 16),
			AlamatDomisili: getString(row, 17),
			// Contact (18-22)
			NoHP:             getString(row, 18),
			NoHPKeluarga:     getString(row, 19),
			NamaKeluarga:     getString(row, 20),
			HubunganKeluarga: getString(row, 21),
			NamaIbuKandung:   getString(row, 22),
			// BPJS & Finance (23-28)
			NoBPJSKesehatan:       getString(row, 23),
			NoBPJSKetenagakerjaan: getString(row, 24),
			NoNPWP:                getString(row, 25),
			NamaBank:              getString(row, 26),
			NoRekening:            getString(row, 27),
			PemilikRekening:       getString(row, 28),
			// Assets (29-31)
			UkuranBaju:   getString(row, 29),
			UkuranSepatu: getString(row, 30),
			UkuranCelana: getString(row, 31),
			// Contract Data (32-35)
			TipeKontrakRaw:        getString(row, 32),
			NoSK:                  getString(row, 33),
			TanggalMulaiKontrak:   getString(row, 34),
			TanggalSelesaiKontrak: getString(row, 35),
			// Auto-derived
			StatusKaryawan:   getString(row, 32),
			TanggalBergabung: getString(row, 34),
		}

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

func (s *service) ValidateStaging(batchID string) error {
	items, err := s.stagingRepo.GetByBatchID(batchID)
	if err != nil {
		return err
	}

	for _, item := range items {
		status := "VALID"
		message := ""

		if item.NRP == "" || item.Nama == "" {
			status = "ERROR"
			message += "NRP and Name are required. "
		}

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

func (s *service) CommitStaging(batchID string) (int, int, error) {
	validItems, err := s.stagingRepo.GetValidByBatchID(batchID)
	if err != nil {
		return 0, 0, err
	}

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

func (s *service) createNewEmployee(stg EmployeeStaging, positions []master.Position, jobSites []master.JobSite, contractTypes []master.ContractType) error {
	nameUpper := strings.ToUpper(stg.Nama)

	password := stg.Password
	if password == "" {
		password = stg.NRP + "123"
	}

	email := strings.TrimSpace(stg.Email)
	if email == "" {
		email = strings.ToLower(stg.NRP) + "@amos.internal"
	}

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

	emp := &employee.Employee{
		UserID: &user.ID,
		NRP:    stg.NRP,
		Name:   nameUpper,
		Gender: stg.JenisKelamin,
		Status: "AKTIF",
	}

	for _, pos := range positions {
		if strings.EqualFold(pos.Name, stg.JabatanRaw) {
			emp.PositionID = &pos.ID
			if pos.DepartmentID != nil {
				emp.DepartmentID = pos.DepartmentID
			}
			break
		}
	}

	for _, js := range jobSites {
		if strings.EqualFold(js.Name, stg.JobSiteRaw) {
			emp.JobSiteID = &js.ID
			break
		}
	}

	if stg.TanggalBergabung != "" {
		if t, err := time.Parse("2006-01-02", stg.TanggalBergabung); err == nil {
			emp.JoinDate = t
		}
	}

	if err := s.employeeRepo.CreateEmployee(emp); err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}

	detail := &employee.EmployeeDetail{
		EmployeeID:        emp.ID,
		NIK:               stg.NIK,
		BirthPlace:        stg.TempatLahir,
		Religion:          stg.Agama,
		BloodType:         stg.GolonganDarah,
		MaritalStatus:     stg.StatusPernikahan,
		AddressKTP:        stg.AlamatKTP,
		AddressDomicile:   stg.AlamatDomisili,
		PhoneNumber:       stg.NoHP,
		PersonalEmail:     stg.Email,
		NPWPNumber:        stg.NoNPWP,
		BPJSKesehatan:     stg.NoBPJSKesehatan,
		BPJSTenagaKerja:   stg.NoBPJSKetenagakerjaan,
		BankName:          stg.NamaBank,
		BankAccount:       stg.NoRekening,
		BankAccountName:   stg.PemilikRekening,
		ShirtSize:         stg.UkuranBaju,
		ShoeSize:          stg.UkuranSepatu,
		PantsSize:         stg.UkuranCelana,
		MotherName:        stg.NamaIbuKandung,
		EmergencyName:     stg.NamaKeluarga,
		EmergencyRelation: stg.HubunganKeluarga,
		EmergencyPhone:    stg.NoHPKeluarga,
	}

	detail.Height = parseInt(stg.TinggiBadan)
	detail.Weight = parseInt(stg.BeratBadan)

	if stg.TanggalLahir != "" {
		if t, err := time.Parse("2006-01-02", stg.TanggalLahir); err == nil {
			detail.BirthDate = t
		}
	}
	if stg.TanggalPernikahan != "" {
		if t, err := time.Parse("2006-01-02", stg.TanggalPernikahan); err == nil {
			detail.MarriageDate = &t
		}
	}

	if err := s.employeeRepo.SaveDetail(detail); err != nil {
		log.Printf("[IMPORT] Warning: failed to save detail for NRP %s: %v", stg.NRP, err)
	}

	if stg.TipeKontrakRaw != "" {
		s.createContractHistory(emp.ID, stg, contractTypes)
	}

	return nil
}

func (s *service) updateExistingEmployee(emp *employee.Employee, stg EmployeeStaging, positions []master.Position, jobSites []master.JobSite) error {
	emp.Name = strings.ToUpper(stg.Nama)
	if stg.JenisKelamin != "" {
		emp.Gender = stg.JenisKelamin
	}

	for _, pos := range positions {
		if strings.EqualFold(pos.Name, stg.JabatanRaw) {
			emp.PositionID = &pos.ID
			if pos.DepartmentID != nil {
				emp.DepartmentID = pos.DepartmentID
			}
			break
		}
	}

	for _, js := range jobSites {
		if strings.EqualFold(js.Name, stg.JobSiteRaw) {
			emp.JobSiteID = &js.ID
			break
		}
	}

	if stg.TanggalBergabung != "" {
		if t, err := time.Parse("2006-01-02", stg.TanggalBergabung); err == nil {
			emp.JoinDate = t
		}
	}

	if err := s.employeeRepo.UpdateEmployee(emp); err != nil {
		return err
	}

	detail := emp.Detail
	if detail == nil {
		detail = &employee.EmployeeDetail{EmployeeID: emp.ID}
	}

	detail.NIK = stg.NIK
	detail.BirthPlace = stg.TempatLahir
	detail.Religion = stg.Agama
	detail.BloodType = stg.GolonganDarah
	detail.MaritalStatus = stg.StatusPernikahan
	detail.AddressKTP = stg.AlamatKTP
	detail.AddressDomicile = stg.AlamatDomisili
	detail.PhoneNumber = stg.NoHP
	detail.PersonalEmail = stg.Email
	detail.NPWPNumber = stg.NoNPWP
	detail.BPJSKesehatan = stg.NoBPJSKesehatan
	detail.BPJSTenagaKerja = stg.NoBPJSKetenagakerjaan
	detail.BankName = stg.NamaBank
	detail.BankAccount = stg.NoRekening
	detail.BankAccountName = stg.PemilikRekening
	detail.ShirtSize = stg.UkuranBaju
	detail.ShoeSize = stg.UkuranSepatu
	detail.PantsSize = stg.UkuranCelana
	detail.MotherName = stg.NamaIbuKandung
	detail.EmergencyName = stg.NamaKeluarga
	detail.EmergencyRelation = stg.HubunganKeluarga
	detail.EmergencyPhone = stg.NoHPKeluarga
	detail.Height = parseInt(stg.TinggiBadan)
	detail.Weight = parseInt(stg.BeratBadan)

	if stg.TanggalLahir != "" {
		if t, err := time.Parse("2006-01-02", stg.TanggalLahir); err == nil {
			detail.BirthDate = t
		}
	}
	if stg.TanggalPernikahan != "" {
		if t, err := time.Parse("2006-01-02", stg.TanggalPernikahan); err == nil {
			detail.MarriageDate = &t
		}
	}

	return s.employeeRepo.SaveDetail(detail)
}

func (s *service) createContractHistory(employeeID uint, stg EmployeeStaging, contractTypes []master.ContractType) {
	var contractTypeID uint
	for _, ct := range contractTypes {
		if strings.EqualFold(ct.Name, stg.TipeKontrakRaw) {
			contractTypeID = ct.ID
			break
		}
	}

	if contractTypeID == 0 {
		return
	}

	contract := &employee.ContractHistory{
		EmployeeID:     employeeID,
		ContractTypeID: &contractTypeID,
		DecreeNumber:   stg.NoSK,
	}

	if stg.TanggalMulaiKontrak != "" {
		if t, err := time.Parse("2006-01-02", stg.TanggalMulaiKontrak); err == nil {
			contract.StartDate = t
		}
	}
	if stg.TanggalSelesaiKontrak != "" {
		if t, err := time.Parse("2006-01-02", stg.TanggalSelesaiKontrak); err == nil {
			contract.EndDate = t
		}
	}

	_ = s.employeeRepo.SaveContractHistory(contract)
}

func getString(row []string, index int) string {
	if index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	val, _ := strconv.Atoi(s)
	return val
}
