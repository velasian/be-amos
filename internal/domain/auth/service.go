package auth

import (
	"amos-backend/pkg/captcha"
	"amos-backend/pkg/email"
	"amos-backend/pkg/otp"
	"amos-backend/pkg/utils"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("kombinasi email/nrp dan password salah")
	ErrTokenGeneration    = errors.New("gagal membuat token login")
	ErrSessionSave        = errors.New("gagal menyimpan session")
	ErrCaptchaFailed      = errors.New("verifikasi captcha gagal")
	ErrTooManyRequests    = errors.New("terlalu banyak permintaan, silakan coba lagi nanti")
	ErrEmailSendFailed    = errors.New("gagal mengirim email OTP")
	ErrUserNotFound       = errors.New("user tidak ditemukan")
)

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
	User         *User  `json:"user"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
}

type Service interface {
	Login(identifier, password, fcmToken, deviceID, platform, userAgent, captchaToken string, isMobile bool) (*LoginResponse, error)
	RefreshToken(refreshToken, deviceID, platform, userAgent string) (*TokenResponse, error)
	Logout(fcmToken, refreshToken string) error
	ForgotPassword(emailAddr string) error
	VerifyOTP(emailAddr, code string) error
	ResetPassword(emailAddr, otpCode, newPassword string) error
	SaveFCMToken(userID uint, token, deviceInfo string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Login(identifier, password, fcmToken, deviceID, platform, userAgent, captchaToken string, isMobile bool) (*LoginResponse, error) {
	// Verifikasi Captcha jika login bukan dari aplikasi mobile
	if !isMobile {
		if err := captcha.Verify(captchaToken); err != nil {
			return nil, ErrCaptchaFailed
		}
	}

	user, err := s.repo.FindUserByEmailOrNRP(identifier)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	if fcmToken != "" {
		s.repo.SaveFCMToken(user.ID, fcmToken, platform)
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	refreshTokenStr, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, ErrTokenGeneration
	}

	rt := RefreshToken{
		UserID:     user.ID,
		Token:      refreshTokenStr,
		DeviceID:   deviceID,
		Platform:   platform,
		UserAgent:  userAgent,
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		LastUsedAt: time.Now(),
	}

	if err := s.repo.SaveRefreshToken(&rt); err != nil {
		return nil, ErrSessionSave
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         user,
	}, nil
}

func (s *service) RefreshToken(refreshToken, deviceID, platform, userAgent string) (*TokenResponse, error) {
	rt, err := s.repo.FindRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("sesi telah kadaluwarsa, silakan login kembali")
	}

	// Revoke old token immediately (rotation)
	s.repo.RevokeRefreshToken(refreshToken)

	// Fetch actual user to get the current role from DB
	user, err := s.repo.FindUserByID(rt.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	newRefreshTokenStr, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, ErrTokenGeneration
	}

	newRt := RefreshToken{
		UserID:     rt.UserID,
		Token:      newRefreshTokenStr,
		DeviceID:   deviceID,
		Platform:   platform,
		UserAgent:  userAgent,
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		LastUsedAt: time.Now(),
	}

	if err := s.repo.SaveRefreshToken(&newRt); err != nil {
		return nil, ErrSessionSave
	}

	return &TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenStr,
	}, nil
}

func (s *service) Logout(fcmToken, refreshToken string) error {
	if fcmToken != "" {
		s.repo.RemoveFCMToken(fcmToken)
	}
	if refreshToken != "" {
		s.repo.RevokeRefreshToken(refreshToken)
	}
	return nil
}

func (s *service) ForgotPassword(emailAddr string) error {
	foundUser, err := s.repo.FindUserByEmailOrNRP(emailAddr)
	if err != nil {
		// Silent error untuk mencegah user enumeration attack
		return nil
	}

	otpCode, err := otp.Generate()
	if err != nil {
		return errors.New("gagal memproses permintaan")
	}

	if err := s.repo.SaveOTP(foundUser.Email, otpCode); err != nil {
		return ErrTooManyRequests
	}

	emailData := struct {
		OTP string
	}{
		OTP: otpCode,
	}

	if err := email.Send(foundUser.Email, "Reset Password OTP", "otp.html", emailData); err != nil {
		return ErrEmailSendFailed
	}

	return nil
}

func (s *service) VerifyOTP(emailAddr, code string) error {
	return s.repo.VerifyOTP(emailAddr, code)
}

func (s *service) ResetPassword(emailAddr, otpCode, newPassword string) error {
	if err := s.repo.VerifyOTP(emailAddr, otpCode); err != nil {
		return err
	}

	foundUser, err := s.repo.FindUserByEmailOrNRP(emailAddr)
	if err != nil {
		return ErrUserNotFound
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("gagal mengenkripsi password baru")
	}
	
	foundUser.Password = hashedPassword
	if err := s.repo.UpdateUser(foundUser); err != nil {
		return errors.New("gagal mereset password")
	}

	s.repo.DeleteOTP(emailAddr)
	return nil
}

func (s *service) SaveFCMToken(userID uint, token, deviceInfo string) error {
	return s.repo.SaveFCMToken(userID, token, deviceInfo)
}
