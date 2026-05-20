package main

import (
	"amos-backend/internal/config"
	"amos-backend/internal/domain/announcement"
	"amos-backend/internal/domain/attendance"
	"amos-backend/internal/domain/auth"
	"amos-backend/internal/domain/employee"
	"amos-backend/internal/domain/importdata"
	"amos-backend/internal/domain/leave"
	"amos-backend/internal/domain/master"
	"amos-backend/internal/domain/mcu"
	"amos-backend/internal/domain/notification"
	"amos-backend/internal/domain/payslip"
	"amos-backend/internal/domain/system"
	"fmt"
	"log"
)

func main() {
	config.LoadEnv()
	config.ConnectDatabase()

	fmt.Println("Memulai AutoMigrate tabel-tabel AMOS...")

	err := config.DB.AutoMigrate(
		// Auth
		&auth.User{},
		&auth.FCMToken{},
		&auth.RefreshToken{},
		&auth.PasswordReset{},
		
		// Master Data
		&master.Department{},
		&master.Position{},
		&master.JobSite{},
		&master.ContractType{},

		// Employee
		&employee.Employee{},
		&employee.EmployeeDetail{},
		&employee.ContractHistory{},

		// Attendance (TA Core)
		&attendance.IoTDevice{},
		&attendance.AttendanceSession{},
		&attendance.Attendance{},

		// HR Modules
		&announcement.Announcement{},
		&leave.LeaveRequest{},
		&mcu.MCUSchedule{},
		&payslip.Payslip{},
		&payslip.PayslipStaging{},

		// System
		&notification.Notification{},
		&system.File{},

		// Import Staging
		&importdata.EmployeeStaging{},
	)

	if err != nil {
		log.Fatalf("Gagal melakukan AutoMigrate: %v", err)
	}

	fmt.Println("✅ AutoMigrate berhasil! Semua tabel telah terbuat di database PostgreSQL.")
}
