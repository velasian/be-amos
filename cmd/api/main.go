package main

import (
	"amos-backend/internal/config"
	"amos-backend/internal/domain/attendance"
	"amos-backend/internal/domain/auth"
	"amos-backend/internal/domain/employee"
	"amos-backend/internal/domain/importdata"
	"amos-backend/internal/domain/master"
	"amos-backend/internal/domain/notification"
	"amos-backend/internal/domain/report"
	"amos-backend/internal/domain/system"
	"amos-backend/internal/middleware"
	"amos-backend/pkg/firebase"
	"amos-backend/pkg/storage"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Environment Variables
	config.LoadEnv()

	// 2. Connect to Database
	config.ConnectDatabase()

	// 3. Initialize External Clients (MinIO + Firebase)
	storageClient := storage.NewMinIOClient()
	fcmClient := firebase.NewClient()

	// 4. Initialize Gin Router with CORS
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// 5. Basic Health Check Endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "pong",
		})
	})

	// 6. Initialize Repositories, Services, and Handlers
	authRepo := auth.NewRepository(config.DB)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	masterRepo := master.NewRepository(config.DB)
	masterService := master.NewService(masterRepo)
	masterHandler := master.NewHandler(masterService)

	employeeRepo := employee.NewRepository(config.DB)
	employeeService := employee.NewService(employeeRepo)
	employeeHandler := employee.NewHandler(employeeService)

	fileRepo := system.NewRepository(config.DB)
	fileService := system.NewService(fileRepo, storageClient)
	fileHandler := system.NewHandler(fileService)

	importRepo := importdata.NewRepository(config.DB)
	importService := importdata.NewService(importRepo, employeeRepo, masterRepo, authRepo)
	importHandler := importdata.NewHandler(importService)

	notifRepo := notification.NewRepository(config.DB)
	notifService := notification.NewService(notifRepo, authRepo, fcmClient)
	notifHandler := notification.NewHandler(notifService)

	sseBroker := attendance.NewSSEBroker()
	attendanceRepo := attendance.NewRepository(config.DB)
	attendanceService := attendance.NewService(attendanceRepo, employeeRepo, masterRepo, notifService, storageClient, sseBroker)
	attendanceHandler := attendance.NewHandler(attendanceService, sseBroker)

	reportRepo := report.NewRepository(config.DB)
	reportService := report.NewService(reportRepo)
	reportHandler := report.NewHandler(reportService)

	// Auto-Seed Master Data (sama seperti referensi)
	// Memasukkan data JobSite, Position, dan Contract Type ke database
	// jika belum ada. Berjalan setiap kali server startup.
	seedMasterData(masterRepo)

	// 7. Setup API Routes
	apiV1 := r.Group("/api/v1")
	{
		authRoutes := apiV1.Group("/auth")
		{
			// Public Routes
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
			authRoutes.POST("/forgot-password", authHandler.ForgotPassword)
			authRoutes.POST("/verify-otp", authHandler.VerifyOTP)
			authRoutes.POST("/reset-password", authHandler.ResetPassword)

			// Protected Routes
			authRoutes.Use(middleware.AuthMiddleware())
			authRoutes.POST("/logout", authHandler.Logout)
			authRoutes.POST("/fcm-token", authHandler.SaveFCMToken)
		}

		masterRoutes := apiV1.Group("/masters")
		masterRoutes.Use(middleware.AuthMiddleware())
		{
			// Read-only (All authenticated users)
			masterRoutes.GET("/departments", masterHandler.GetDepartments)
			masterRoutes.GET("/positions", masterHandler.GetPositions)
			masterRoutes.GET("/job-sites", masterHandler.GetJobSites)
			masterRoutes.GET("/contract-types", masterHandler.GetContractTypes)

			// Write (HR Admin & Superadmin only)
			hrRoutes := masterRoutes.Group("")
			hrRoutes.Use(middleware.RoleMiddleware("admin_hr"))
			{
				hrRoutes.POST("/departments", masterHandler.CreateDepartment)
				hrRoutes.PUT("/departments/:id", masterHandler.UpdateDepartment)
				hrRoutes.DELETE("/departments/:id", masterHandler.DeleteDepartment)

				hrRoutes.POST("/positions", masterHandler.CreatePosition)
				hrRoutes.PUT("/positions/:id", masterHandler.UpdatePosition)
				hrRoutes.DELETE("/positions/:id", masterHandler.DeletePosition)

				hrRoutes.POST("/job-sites", masterHandler.CreateJobSite)
				hrRoutes.PUT("/job-sites/:id", masterHandler.UpdateJobSite)
				hrRoutes.DELETE("/job-sites/:id", masterHandler.DeleteJobSite)

				hrRoutes.POST("/contract-types", masterHandler.CreateContractType)
				hrRoutes.PUT("/contract-types/:id", masterHandler.UpdateContractType)
				hrRoutes.DELETE("/contract-types/:id", masterHandler.DeleteContractType)
			}
		}

		employeeRoutes := apiV1.Group("/employees")
		employeeRoutes.Use(middleware.AuthMiddleware())
		{
			// Me (Self Service)
			employeeRoutes.GET("/me", employeeHandler.GetMe)
			employeeRoutes.PUT("/me", employeeHandler.UpdateMe)

			// HR Admin Only
			hrRoutes := employeeRoutes.Group("")
			hrRoutes.Use(middleware.RoleMiddleware("admin_hr"))
			{
				hrRoutes.GET("", employeeHandler.GetAllEmployees)
				hrRoutes.GET("/export", employeeHandler.ExportExcel)
				hrRoutes.GET("/:id", employeeHandler.GetEmployeeByID)
				hrRoutes.POST("", employeeHandler.CreateEmployee)
				hrRoutes.PUT("/:id", employeeHandler.UpdateEmployee)
				hrRoutes.DELETE("/:id", employeeHandler.DeleteEmployee)
			}
		}

		// File Management Routes
		fileRoutes := apiV1.Group("/files")
		fileRoutes.Use(middleware.AuthMiddleware())
		{
			// All authenticated users can view and download files
			fileRoutes.GET("", fileHandler.GetFilesByEntity)
			fileRoutes.GET("/:id/download", fileHandler.GetFileDownloadURL)

			// Upload & Delete (HR Admin only)
			fileHRRoutes := fileRoutes.Group("")
			fileHRRoutes.Use(middleware.RoleMiddleware("admin_hr"))
			{
				fileHRRoutes.POST("/upload", fileHandler.UploadFile)
				fileHRRoutes.DELETE("/:id", fileHandler.DeleteFile)
			}
		}

		// Employee Import Routes (HR Admin only)
		importRoutes := apiV1.Group("/import")
		importRoutes.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin_hr"))
		{
			importRoutes.GET("/template", importHandler.DownloadTemplate)
			importRoutes.POST("/parse", importHandler.ParseExcel)
			importRoutes.GET("/staging/:batchId", importHandler.GetStagingData)
			importRoutes.PATCH("/staging/:id", importHandler.UpdateStagingField)
			importRoutes.POST("/commit", importHandler.SubmitImport)
		}

		// Notification Routes
		notifRoutes := apiV1.Group("/notifications")
		notifRoutes.Use(middleware.AuthMiddleware())
		{
			notifRoutes.GET("", notifHandler.GetInbox)
			notifRoutes.GET("/unread-count", notifHandler.GetUnreadCount)
			notifRoutes.PATCH("/:id/read", notifHandler.MarkAsRead)
			notifRoutes.PATCH("/read-all", notifHandler.MarkAllAsRead)

			// Admin test endpoint
			notifRoutes.POST("/test", middleware.RoleMiddleware("admin_hr"), notifHandler.SendTestNotification)
		}

		// IoT Routes (API Key auth — for ESP32 devices)
		iotRoutes := apiV1.Group("/iot")
		iotRoutes.Use(middleware.IoTAuthMiddleware())
		{
			iotRoutes.POST("/scan", attendanceHandler.ScanNFC)
			iotRoutes.POST("/assign", attendanceHandler.ReportNFCUID)
		}

		// IoT Device Management + NFC Registration (JWT + Admin only)
		iotAdminRoutes := apiV1.Group("/iot")
		iotAdminRoutes.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin_hr"))
		{
			iotAdminRoutes.POST("/devices", attendanceHandler.RegisterDevice)
			iotAdminRoutes.GET("/devices", attendanceHandler.GetAllDevices)
			iotAdminRoutes.GET("/listen", attendanceHandler.ListenNFC)
			iotAdminRoutes.POST("/assign-employee", attendanceHandler.AssignNFC)
		}

		// Attendance Routes (JWT — for mobile app)
		attendanceRoutes := apiV1.Group("/attendances")
		attendanceRoutes.Use(middleware.AuthMiddleware())
		{
			attendanceRoutes.GET("/session", attendanceHandler.GetActiveSession)
			attendanceRoutes.GET("/me", attendanceHandler.GetMyAttendances)
			attendanceRoutes.POST("/verify", attendanceHandler.VerifyAttendance)

			attendanceHRRoutes := attendanceRoutes.Group("")
			attendanceHRRoutes.Use(middleware.RoleMiddleware("admin_hr"))
			{
				attendanceHRRoutes.GET("", attendanceHandler.GetAllAttendances)
			}
		}

		// Report Routes (HR Admin only)
		reportRoutes := apiV1.Group("/reports")
		reportRoutes.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin_hr"))
		{
			reportRoutes.GET("/stats", reportHandler.GetStats)
			reportRoutes.GET("/attendance/export", reportHandler.ExportAttendance)
			reportRoutes.GET("/export", reportHandler.ExportAttendance)
		}
	}

	// 8. Start Server
	port := config.GetEnv("PORT", "8080")
	log.Printf("Server AMOS berjalan di port %s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}

// seedMasterData memasukkan data master default ke database jika belum ada.
// Fungsi ini idempotent — aman dijalankan berulang kali tanpa menyebabkan duplikasi.
// Data yang di-seed sama persis dengan yang ada di referensi (amos-hcgs).
func seedMasterData(repo master.Repository) {
	log.Println("[SEED] Memulai seeding data master...")

	// ========================================
	// 1. Seed JobSites (Lokasi Kerja)
	// ========================================
	// Daftar lokasi kerja standar yang digunakan di lapangan
	sites := []string{"BRE RANTAU", "AGMR BLOK 2", "AGMR BLOK 3"}

	// Ambil semua job site yang sudah ada di database
	existingSites, _ := repo.GetAllJobSites()

	// Buat map untuk pengecekan cepat apakah site sudah ada (O(1) lookup)
	existingSiteMap := make(map[string]bool)
	for _, s := range existingSites {
		existingSiteMap[s.Name] = true // Kunci = nama, nilai = true (ada)
	}

	// Loop setiap site, hanya insert jika belum ada di map
	for _, s := range sites {
		if !existingSiteMap[s] {
			if err := repo.CreateJobSite(&master.JobSite{Name: s}); err != nil {
				log.Printf("[SEED] Error membuat JobSite %s: %v", s, err)
			} else {
				log.Printf("[SEED] JobSite dibuat: %s", s)
			}
		}
	}

	// ========================================
	// 2. Seed Positions (Jabatan)
	// ========================================
	// Daftar jabatan standar di perusahaan mining/kontraktor
	positions := []string{
		"DRIVER", "DRIVER WT", "OPERATOR", "OPERATOR MASTER", "GENERAL MANAGER",
		"PJO", "SAFETY OFFICER", "GL", "MPIS", "OFFICER",
		"OFFICER HO", "OFFICER PLANER", "SECURITY", "WAKAR",
		"CHIEF MEKANIK", "MEKANIK", "MEKANIK JR", "HELPER",
		"J PART", "DRIVER SARANA LV", "KASIR", "DIREKTUR",
	}

	// Ambil semua position yang sudah ada di database
	existingPos, err := repo.GetAllPositions()
	if err != nil {
		log.Printf("[SEED] Error mengambil positions: %v", err)
	}
	log.Printf("[SEED] Ditemukan %d positions yang sudah ada", len(existingPos))

	// Buat map untuk pengecekan cepat
	existingPosMap := make(map[string]bool)
	for _, p := range existingPos {
		existingPosMap[p.Name] = true
	}

	// Insert position baru yang belum ada
	createdCount := 0
	for _, p := range positions {
		if !existingPosMap[p] {
			if err := repo.CreatePosition(&master.Position{Name: p}); err != nil {
				log.Printf("[SEED] Error membuat Position %s: %v", p, err)
			} else {
				createdCount++
			}
		}
	}
	log.Printf("[SEED] %d positions baru dibuat", createdCount)

	// ========================================
	// 3. Seed Contract Types (Tipe Kontrak)
	// ========================================
	// Daftar tipe kontrak: Kontrak I sampai IX, Permanen, dan Tidak Aktif
	contractTypes := []string{
		"Kontrak I", "Kontrak II", "Kontrak III", "Kontrak IV", "Kontrak V",
		"Kontrak VI", "Kontrak VII", "Kontrak VIII", "Kontrak IX",
		"Permanen", "Tidak Aktif",
	}

	// Ambil semua contract type yang sudah ada
	existingCT, _ := repo.GetAllContractTypes()

	// Buat map untuk pengecekan cepat
	existingCTMap := make(map[string]bool)
	for _, ct := range existingCT {
		existingCTMap[ct.Name] = true
	}

	// Insert contract type baru yang belum ada
	ctCreatedCount := 0
	for _, ct := range contractTypes {
		if !existingCTMap[ct] {
			if err := repo.CreateContractType(&master.ContractType{Name: ct}); err != nil {
				log.Printf("[SEED] Error membuat ContractType %s: %v", ct, err)
			} else {
				ctCreatedCount++
			}
		}
	}
	log.Printf("[SEED] %d contract types baru dibuat", ctCreatedCount)

	log.Println("[SEED] Seeding selesai!")
}
