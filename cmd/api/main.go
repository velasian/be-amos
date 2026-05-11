package main

import (
	"amos-backend/internal/config"
	"amos-backend/internal/domain/attendance"
	"amos-backend/internal/domain/auth"
	"amos-backend/internal/domain/employee"
	"amos-backend/internal/domain/importdata"
	"amos-backend/internal/domain/master"
	"amos-backend/internal/domain/notification"
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

	// 3. Initialize MinIO Storage Client
	storageClient := storage.NewMinIOClient()

	// 4. Initialize Firebase Client
	fcmClient := firebase.NewClient()

	// 4. Initialize Gin Router
	r := gin.Default()

	// 5. Basic Health Check Endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "pong",
		})
	})

	// 5. Initialize Repositories, Services, and Handlers
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

	// 6. Setup API Routes
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
			attendanceRoutes.POST("/verify", attendanceHandler.VerifyAttendance)
		}
	}

	// 7. Start Server
	port := config.GetEnv("PORT", "8080")
	log.Printf("Server AMOS berjalan di port %s", port)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
