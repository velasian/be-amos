package auth

import "time"

// User merepresentasikan tabel 'users'
type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	NRP       *string    `gorm:"type:varchar(50);uniqueIndex" json:"nrp"` // *string (Pointer) agar mengizinkan NULL untuk Superadmin
	Email     string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string     `gorm:"type:varchar(255);not null" json:"-"`     // json:"-" agar password tidak ikut terkirim di API
	Role      string     `gorm:"type:varchar(50);default:'employee'" json:"role"`
	FCMTokens []FCMToken `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"fcm_tokens,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// FCMToken merepresentasikan tabel 'fcm_tokens'
type FCMToken struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	Token      string    `gorm:"type:text;uniqueIndex;not null" json:"token"`
	DeviceInfo string    `gorm:"type:varchar(255)" json:"device_info"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RefreshToken merepresentasikan sesi login (Rotasi Token untuk Production Ready)
type RefreshToken struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	Token      string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	DeviceID   string    `gorm:"type:varchar(100)" json:"device_id"`
	Platform   string    `gorm:"type:varchar(50)" json:"platform"`
	UserAgent  string    `gorm:"type:varchar(255)" json:"user_agent"`
	Revoked    bool      `gorm:"default:false" json:"revoked"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// PasswordReset merepresentasikan OTP untuk lupa password
type PasswordReset struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(255);not null" json:"email"`
	Token     string    `gorm:"type:varchar(10);not null" json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
