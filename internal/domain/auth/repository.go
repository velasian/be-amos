package auth

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindUserByEmailOrNRP(identifier string) (*User, error)
	FindUserByID(id uint) (*User, error)
	SaveFCMToken(userID uint, token, deviceInfo string) error
	RemoveFCMToken(token string) error
	CreateUser(u *User) error
	UpdateUser(u *User) error
	SaveRefreshToken(rt *RefreshToken) error
	FindRefreshToken(token string) (*RefreshToken, error)
	RevokeRefreshToken(token string) error
	SaveOTP(email, otp string) error
	VerifyOTP(email, otp string) error
	DeleteOTP(email string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindUserByEmailOrNRP(identifier string) (*User, error) {
	var u User
	err := r.db.Where("email = ? OR nrp = ?", identifier, identifier).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) FindUserByID(id uint) (*User, error) {
	var u User
	err := r.db.Select("id", "nrp", "email", "role").First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) SaveFCMToken(userID uint, token, deviceInfo string) error {
	var existing FCMToken
	err := r.db.Where("token = ?", token).First(&existing).Error
	if err == nil {
		existing.UserID = userID
		existing.DeviceInfo = deviceInfo
		return r.db.Save(&existing).Error
	}

	newToken := FCMToken{
		UserID:     userID,
		Token:      token,
		DeviceInfo: deviceInfo,
	}
	return r.db.Create(&newToken).Error
}

func (r *repository) RemoveFCMToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&FCMToken{}).Error
}

func (r *repository) CreateUser(u *User) error {
	return r.db.Create(u).Error
}

func (r *repository) UpdateUser(u *User) error {
	return r.db.Save(u).Error
}

func (r *repository) SaveRefreshToken(rt *RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *repository) FindRefreshToken(token string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.db.Where("token = ? AND revoked = ? AND expires_at > ?", token, false, time.Now()).First(&rt).Error
	if err != nil {
		return nil, err
	}

	r.db.Model(&rt).Update("last_used_at", time.Now())
	return &rt, nil
}

func (r *repository) RevokeRefreshToken(token string) error {
	return r.db.Model(&RefreshToken{}).Where("token = ?", token).Update("revoked", true).Error
}

func (r *repository) SaveOTP(email string, otp string) error {
	var existingOTP PasswordReset
	err := r.db.Where("email = ?", email).First(&existingOTP).Error
	if err == nil {
		// Cooldown 90 detik
		if time.Since(existingOTP.CreatedAt) < 90*time.Second {
			return errors.New("silakan tunggu 90 detik sebelum meminta OTP baru")
		}
		existingOTP.Token = otp
		existingOTP.ExpiresAt = time.Now().Add(10 * time.Minute)
		existingOTP.CreatedAt = time.Now()
		return r.db.Save(&existingOTP).Error
	}

	newOTP := PasswordReset{
		Email:     email,
		Token:     otp,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	}
	return r.db.Create(&newOTP).Error
}

func (r *repository) VerifyOTP(email string, otp string) error {
	var record PasswordReset
	err := r.db.Where("email = ? AND token = ?", email, otp).First(&record).Error
	if err != nil {
		return errors.New("OTP tidak valid")
	}

	if time.Now().After(record.ExpiresAt) {
		return errors.New("OTP sudah kadaluwarsa")
	}
	return nil
}

func (r *repository) DeleteOTP(email string) error {
	return r.db.Where("email = ?", email).Delete(&PasswordReset{}).Error
}


