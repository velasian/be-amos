package main

import (
	"amos-backend/internal/config"
	"amos-backend/internal/domain/auth"
	"amos-backend/internal/domain/employee"
	"amos-backend/internal/domain/master"
	"amos-backend/internal/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Environment Variables
	config.LoadEnv()

	// 2. Connect to Database
	config.ConnectDatabase()

	// 3. Initialize Gin Router
	r := gin.Default()

	// 4. Basic Health Check Endpoint
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
	}

	// 7. Start Server
	port := config.GetEnv("PORT", "8080")
	log.Printf("Server AMOS berjalan di port %s", port)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
