package main

import (
	"amos-backend/internal/config"
	"amos-backend/internal/domain/attendance"
	"amos-backend/internal/domain/auth"
	"amos-backend/internal/domain/employee"
	"amos-backend/internal/domain/master"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type seedUserInput struct {
	Email    string
	NRP      string
	Password string
	Role     string
	Label    string
}

func main() {
	config.LoadEnv()
	config.ConnectDatabase()

	fmt.Println("Starting AMOS seed data...")

	superadmin := ensureUser(seedUserInput{
		Email:    "superadmin@amos.com",
		Password: "superadmin123",
		Role:     "superadmin",
		Label:    "Superadmin",
	})
	adminHR := ensureUser(seedUserInput{
		Email:    "admin.hr@amos.com",
		NRP:      "ADM001",
		Password: "adminhr123",
		Role:     "admin_hr",
		Label:    "Admin HR",
	})
	employeeUser := ensureUser(seedUserInput{
		Email:    "employee@amos.com",
		NRP:      "EMP001",
		Password: "employee123",
		Role:     "employee",
		Label:    "Pegawai Demo",
	})
	ryanUser := ensureUser(seedUserInput{
		Email:    "ryansyahrullah62@gmail.com",
		NRP:      "RYAN001",
		Password: "@Ryan0852",
		Role:     "superadmin",
		Label:    "Ryan",
	})
	_ = superadmin

	hrDepartment := ensureDepartment("HRGA", "HRGA")
	opsDepartment := ensureDepartment("Operations", "OPS")
	hrPosition := ensurePosition("HR Officer", hrDepartment.ID)
	employeePosition := ensurePosition("General Staff", opsDepartment.ID)
	contractType := ensureContractType("PKWT I")

	jobSite := ensureJobSite("Head Office Jakarta", -6.200000, 106.816666, 100)

	ensureEmployee(adminHR, "ADM001", "Admin HR", "F", hrDepartment.ID, hrPosition.ID, jobSite.ID)
	demoEmployee := ensureEmployee(employeeUser, "EMP001", "Demo Employee", "M", opsDepartment.ID, employeePosition.ID, jobSite.ID)
	ryanEmployee := ensureEmployee(ryanUser, "RYAN001", "Ryan Syahrullah", "M", opsDepartment.ID, employeePosition.ID, jobSite.ID)
	ensureContractHistory(demoEmployee.ID, contractType.ID)
	ensureContractHistory(ryanEmployee.ID, contractType.ID)

	ensureIoTDevice("ESP32 Main Gate", "amos_secret_device_key_001", jobSite.ID)

	fmt.Println("Seed completed.")
	fmt.Println("Default credentials:")
	fmt.Println("- Superadmin : superadmin@amos.com / superadmin123")
	fmt.Println("- Ryan       : ryansyahrullah62@gmail.com / @Ryan0852")
	fmt.Println("- Admin HR   : admin.hr@amos.com / adminhr123")
	fmt.Println("- Pegawai    : employee@amos.com / employee123")
}

func ensureUser(input seedUserInput) *auth.User {
	var existing auth.User
	err := config.DB.Where("email = ?", input.Email).First(&existing).Error
	if err == nil {
		changed := false
		if existing.Role != input.Role {
			existing.Role = input.Role
			changed = true
		}
		if input.NRP != "" && (existing.NRP == nil || *existing.NRP != input.NRP) {
			nrp := input.NRP
			existing.NRP = &nrp
			changed = true
		}
		if changed {
			if err := config.DB.Save(&existing).Error; err != nil {
				log.Fatalf("failed to update %s user: %v", input.Label, err)
			}
			fmt.Printf("Updated %s user.\n", input.Label)
		} else {
			fmt.Printf("%s user already exists, skipped.\n", input.Label)
		}
		return &existing
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query %s user: %v", input.Label, err)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash %s password: %v", input.Label, err)
	}

	var nrpPtr *string
	if input.NRP != "" {
		nrp := input.NRP
		nrpPtr = &nrp
	}

	user := auth.User{
		NRP:      nrpPtr,
		Email:    input.Email,
		Password: string(hashPassword),
		Role:     input.Role,
	}
	if err := config.DB.Create(&user).Error; err != nil {
		log.Fatalf("failed to create %s user: %v", input.Label, err)
	}

	fmt.Printf("Created %s user.\n", input.Label)
	return &user
}

func ensureDepartment(name, code string) *master.Department {
	var department master.Department
	err := config.DB.Where("name = ?", name).First(&department).Error
	if err == nil {
		if department.Code != code {
			department.Code = code
			if err := config.DB.Save(&department).Error; err != nil {
				log.Fatalf("failed to update department %s: %v", name, err)
			}
		}
		fmt.Printf("Department %s already exists, skipped.\n", name)
		return &department
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query department %s: %v", name, err)
	}

	department = master.Department{Name: name, Code: code}
	if err := config.DB.Create(&department).Error; err != nil {
		log.Fatalf("failed to create department %s: %v", name, err)
	}
	fmt.Printf("Created department %s.\n", name)
	return &department
}

func ensurePosition(name string, departmentID uint) *master.Position {
	var position master.Position
	err := config.DB.Where("name = ?", name).First(&position).Error
	if err == nil {
		if position.DepartmentID == nil || *position.DepartmentID != departmentID {
			position.DepartmentID = &departmentID
			if err := config.DB.Save(&position).Error; err != nil {
				log.Fatalf("failed to update position %s: %v", name, err)
			}
		}
		fmt.Printf("Position %s already exists, skipped.\n", name)
		return &position
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query position %s: %v", name, err)
	}

	position = master.Position{Name: name, DepartmentID: &departmentID}
	if err := config.DB.Create(&position).Error; err != nil {
		log.Fatalf("failed to create position %s: %v", name, err)
	}
	fmt.Printf("Created position %s.\n", name)
	return &position
}

func ensureContractType(name string) *master.ContractType {
	var contractType master.ContractType
	err := config.DB.Where("name = ?", name).First(&contractType).Error
	if err == nil {
		fmt.Printf("Contract type %s already exists, skipped.\n", name)
		return &contractType
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query contract type %s: %v", name, err)
	}

	contractType = master.ContractType{Name: name}
	if err := config.DB.Create(&contractType).Error; err != nil {
		log.Fatalf("failed to create contract type %s: %v", name, err)
	}
	fmt.Printf("Created contract type %s.\n", name)
	return &contractType
}

func ensureJobSite(name string, latitude, longitude float64, radiusMeters int) *master.JobSite {
	var jobSite master.JobSite
	err := config.DB.Where("name = ?", name).First(&jobSite).Error
	if err == nil {
		jobSite.Latitude = latitude
		jobSite.Longitude = longitude
		jobSite.RadiusMeters = radiusMeters
		if err := config.DB.Save(&jobSite).Error; err != nil {
			log.Fatalf("failed to update job site %s: %v", name, err)
		}
		fmt.Printf("Job site %s already exists, updated.\n", name)
		return &jobSite
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query job site %s: %v", name, err)
	}

	jobSite = master.JobSite{
		Name:         name,
		Latitude:     latitude,
		Longitude:    longitude,
		RadiusMeters: radiusMeters,
	}
	if err := config.DB.Create(&jobSite).Error; err != nil {
		log.Fatalf("failed to create job site %s: %v", name, err)
	}
	fmt.Printf("Created job site %s.\n", name)
	return &jobSite
}

func ensureEmployee(user *auth.User, nrp, name, gender string, departmentID, positionID, jobSiteID uint) *employee.Employee {
	var emp employee.Employee
	err := config.DB.Where("nrp = ?", nrp).First(&emp).Error
	if err == nil {
		emp.UserID = &user.ID
		emp.Name = name
		emp.Gender = gender
		emp.DepartmentID = &departmentID
		emp.PositionID = &positionID
		emp.JobSiteID = &jobSiteID
		emp.Status = "AKTIF"
		if err := config.DB.Save(&emp).Error; err != nil {
			log.Fatalf("failed to update employee %s: %v", nrp, err)
		}
		ensureEmployeeDetail(emp.ID)
		fmt.Printf("Employee %s already exists, updated.\n", nrp)
		return &emp
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query employee %s: %v", nrp, err)
	}

	emp = employee.Employee{
		UserID:       &user.ID,
		NRP:          nrp,
		Name:         name,
		Gender:       gender,
		DepartmentID: &departmentID,
		PositionID:   &positionID,
		JobSiteID:    &jobSiteID,
		Status:       "AKTIF",
		JoinDate:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := config.DB.Create(&emp).Error; err != nil {
		log.Fatalf("failed to create employee %s: %v", nrp, err)
	}
	ensureEmployeeDetail(emp.ID)
	fmt.Printf("Created employee %s.\n", nrp)
	return &emp
}

func ensureEmployeeDetail(employeeID uint) {
	var detail employee.EmployeeDetail
	err := config.DB.Where("employee_id = ?", employeeID).First(&detail).Error
	if err == nil {
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query employee detail %d: %v", employeeID, err)
	}

	detail = employee.EmployeeDetail{EmployeeID: employeeID}
	if err := config.DB.Create(&detail).Error; err != nil {
		log.Fatalf("failed to create employee detail %d: %v", employeeID, err)
	}
}

func ensureContractHistory(employeeID, contractTypeID uint) {
	var contract employee.ContractHistory
	err := config.DB.Where("employee_id = ? AND contract_type_id = ?", employeeID, contractTypeID).First(&contract).Error
	if err == nil {
		fmt.Printf("Contract history for employee %d already exists, skipped.\n", employeeID)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query contract history: %v", err)
	}

	contract = employee.ContractHistory{
		EmployeeID:     employeeID,
		ContractTypeID: &contractTypeID,
		DecreeNumber:   "SK/AMOS/DEMO/2026",
		StartDate:      time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		EndDate:        time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC),
	}
	if err := config.DB.Create(&contract).Error; err != nil {
		log.Fatalf("failed to create contract history: %v", err)
	}
	fmt.Printf("Created contract history for employee %d.\n", employeeID)
}

func ensureIoTDevice(name, apiKey string, jobSiteID uint) *attendance.IoTDevice {
	var device attendance.IoTDevice
	err := config.DB.Where("api_key = ?", apiKey).First(&device).Error
	if err == nil {
		device.Name = name
		device.JobSiteID = &jobSiteID
		device.IsActive = true
		if err := config.DB.Save(&device).Error; err != nil {
			log.Fatalf("failed to update IoT device %s: %v", name, err)
		}
		fmt.Printf("IoT device %s already exists, updated.\n", name)
		return &device
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to query IoT device %s: %v", name, err)
	}

	device = attendance.IoTDevice{
		JobSiteID: &jobSiteID,
		Name:      name,
		APIKey:    apiKey,
		IsActive:  true,
	}
	if err := config.DB.Create(&device).Error; err != nil {
		log.Fatalf("failed to create IoT device %s: %v", name, err)
	}
	fmt.Printf("Created IoT device %s.\n", name)
	return &device
}
