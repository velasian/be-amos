package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

type LoginRequest struct {
	Identifier   string `json:"identifier" binding:"required"`
	Password     string `json:"password" binding:"required"`
	CaptchaToken string `json:"captcha_token"`
	FCMToken     string `json:"fcm_token"`
}

func (h *Handler) setAuthCookies(c *gin.Context, refreshToken string) {
	isSecure := os.Getenv("APP_ENV") != "development"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", refreshToken, 3600*24*30, "/", "", isSecure, true)
}

func (h *Handler) clearAuthCookies(c *gin.Context) {
	isSecure := os.Getenv("APP_ENV") != "development"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", "", -1, "/", "", isSecure, true)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid"})
		return
	}

	deviceID := c.GetHeader("X-Device-Id")
	platform := c.GetHeader("X-Platform")
	userAgent := c.GetHeader("User-Agent")
	isMobile := c.GetHeader("X-App-Source") == "mobile"

	resp, err := h.service.Login(req.Identifier, req.Password, req.FCMToken, deviceID, platform, userAgent, req.CaptchaToken, isMobile)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.setAuthCookies(c, resp.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Login berhasil",
		"access_token": resp.AccessToken,
		"user": gin.H{
			"id":    resp.User.ID,
			"nrp":   resp.User.NRP,
			"email": resp.User.Email,
			"role":  resp.User.Role,
		},
	})
}

func (h *Handler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session tidak valid atau kadaluwarsa"})
		return
	}

	deviceID := c.GetHeader("X-Device-Id")
	platform := c.GetHeader("X-Platform")
	userAgent := c.GetHeader("User-Agent")

	resp, err := h.service.RefreshToken(refreshToken, deviceID, platform, userAgent)
	if err != nil {
		h.clearAuthCookies(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.setAuthCookies(c, resp.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"access_token": resp.AccessToken})
}

func (h *Handler) Logout(c *gin.Context) {
	fcmToken := c.GetHeader("X-FCM-Token")
	refreshToken, _ := c.Cookie("refresh_token")
	
	h.service.Logout(fcmToken, refreshToken)
	h.clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logout berhasil"})
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ForgotPassword(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Jika email terdaftar, OTP telah dikirimkan."})
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

func (h *Handler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.VerifyOTP(req.Email, req.OTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP valid"})
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ResetPassword(req.Email, req.OTP, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}
