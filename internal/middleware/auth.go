package middleware

import (
	"amos-backend/internal/config"
	"amos-backend/internal/domain/auth"
	"amos-backend/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware memvalidasi JWT Access Token dan menyematkan User ke dalam Context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Coba ambil dari Header Authorization Bearer (Standard JWT)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. Fallback: coba ambil dari Cookie jika ada (terkadang web butuh ini untuk file download)
		if tokenString == "" {
			if cookie, err := c.Cookie("access_token"); err == nil {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Akses Token tidak ditemukan"})
			c.Abort()
			return
		}

		// Validasi kelayakan Token JWT (Cek Signature dan Expiry 15 Menit)
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi akses telah kadaluwarsa, silakan pakai refresh token"})
			c.Abort()
			return
		}

		// Ambil data user dari DB untuk memastikan user masih ada dan rolenya up-to-date
		var user auth.User
		if err := config.DB.Select("id", "nrp", "email", "role").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akun tidak ditemukan di sistem"})
			c.Abort()
			return
		}

		// Simpan data di context agar bisa dipakai oleh Handler selanjutnya
		c.Set("userID", user.ID)
		c.Set("role", user.Role)
		c.Set("user", &user)

		c.Next()
	}
}

// RoleMiddleware memastikan hanya role tertentu yang bisa mengakses rute ini
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Status user tidak jelas"})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tipe role internal server error"})
			c.Abort()
			return
		}

		// Superadmin adalah dewa, bisa menembus batasan role apapun
		if roleStr == "superadmin" {
			c.Next()
			return
		}

		for _, role := range allowedRoles {
			if role == roleStr {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk mengakses fitur ini"})
		c.Abort()
	}
}

// IoTAuthMiddleware authenticates IoT devices (ESP32) via X-API-Key header.
// It validates the key against the iot_devices table and injects the device
// into the context for downstream handlers. This is completely separate from JWT auth.
func IoTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "API key is required (X-API-Key header)",
			})
			c.Abort()
			return
		}

		// Look up device by API key
		var device struct {
			ID        uint  `gorm:"primaryKey"`
			JobSiteID *uint
			IsActive  bool
		}
		if err := config.DB.Table("iot_devices").
			Where("api_key = ?", apiKey).
			First(&device).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid API key: device not registered",
			})
			c.Abort()
			return
		}

		// Check if device is active
		if !device.IsActive {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "Device is deactivated",
			})
			c.Abort()
			return
		}

		// Inject device info into context
		c.Set("iotDeviceID", device.ID)
		if device.JobSiteID != nil {
			c.Set("iotJobSiteID", *device.JobSiteID)
		}

		c.Next()
	}
}
