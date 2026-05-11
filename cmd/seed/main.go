package main

import (
	"amos-backend/internal/config"
	"amos-backend/internal/domain/attendance"
	"amos-backend/internal/domain/auth"
	"amos-backend/internal/domain/master"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	config.LoadEnv()
	config.ConnectDatabase()

	fmt.Println("Memulai Seeding Data Awal...")

	// 1. Seed Superadmin
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte("superadmin123"), bcrypt.DefaultCost)
	superadmin := auth.User{
		Email:    "superadmin@amos.com",
		Password: string(hashPassword),
		Role:     "superadmin",
	}

	var existingUser auth.User
	if err := config.DB.Where("email = ?", superadmin.Email).First(&existingUser).Error; err != nil {
		config.DB.Create(&superadmin)
		fmt.Println("✅ Data Superadmin berhasil ditambahkan!")
	} else {
		fmt.Println("⚠️ Data Superadmin sudah ada, dilewati.")
	}

	// 2. Seed Job Site Default
	jobSite := master.JobSite{
		Name:         "Head Office Jakarta",
		Latitude:     -6.200000,
		Longitude:    106.816666,
		RadiusMeters: 100,
	}

	var existingJobSite master.JobSite
	if err := config.DB.Where("name = ?", jobSite.Name).First(&existingJobSite).Error; err != nil {
		config.DB.Create(&jobSite)
		fmt.Println("✅ Data Job Site 'Head Office Jakarta' berhasil ditambahkan!")
		existingJobSite = jobSite
	} else {
		fmt.Println("⚠️ Data Job Site 'Head Office Jakarta' sudah ada, dilewati.")
	}

	// 3. Seed IoT Device Default
	iotDevice := attendance.IoTDevice{
		JobSiteID: &existingJobSite.ID,
		Name:      "ESP32 Main Gate",
		APIKey:    "amos_secret_device_key_001",
		IsActive:  true,
	}

	var existingDevice attendance.IoTDevice
	if err := config.DB.Where("api_key = ?", iotDevice.APIKey).First(&existingDevice).Error; err != nil {
		config.DB.Create(&iotDevice)
		fmt.Println("✅ Data IoT Device 'ESP32 Main Gate' berhasil ditambahkan!")
	} else {
		fmt.Println("⚠️ Data IoT Device 'ESP32 Main Gate' sudah ada, dilewati.")
	}

	fmt.Println("🎉 Seeding Selesai!")
}
