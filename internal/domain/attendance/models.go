package attendance

import (
	"time"
)

// IoTDevice mewakili alat pemindai (ESP32) di lokasi tertentu
type IoTDevice struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	JobSiteID *uint     `json:"job_site_id"` // Berelasi ke master.JobSite
	Name      string    `gorm:"type:varchar(100)" json:"name"`
	APIKey    string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"` // Disembunyikan di JSON agar aman
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName overrides GORM's default table naming (prevents "io_t_devices")
func (IoTDevice) TableName() string { return "iot_devices" }

// AttendanceSession mewakili sesi sementara saat NFC baru saja di-tap
type AttendanceSession struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EmployeeID  uint      `gorm:"not null" json:"employee_id"` // Berelasi ke employee.Employee
	IoTDeviceID *uint     `json:"iot_device_id"`               // Berelasi ke IoTDevice
	Status      string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, verified, expired
	ScannedAt   time.Time `json:"scanned_at"`
	ExpiresAt   time.Time `json:"expires_at"` // Waktu kadaluwarsa sesi (misal: 5 menit)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Attendance mewakili catatan akhir kehadiran yang sudah divalidasi oleh Mobile App
type Attendance struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	SessionID        uint      `gorm:"uniqueIndex;not null" json:"session_id"` // 1 session hanya punya 1 attendance
	EmployeeID       uint      `gorm:"not null" json:"employee_id"`
	Type             string    `gorm:"type:varchar(50)" json:"type"` // clock_in, clock_out
	SelfieURL        string    `gorm:"type:text" json:"selfie_url"`
	Latitude         float64   `gorm:"type:double precision" json:"latitude"`
	Longitude        float64   `gorm:"type:double precision" json:"longitude"`
	IsWithinGeofence bool      `json:"is_within_geofence"`
	GeofenceDistance float64   `gorm:"type:double precision" json:"geofence_distance"`
	IsLate           bool      `json:"is_late"`
	LateMinutes      int       `gorm:"default:0" json:"late_minutes"`
	RecordedAt       time.Time `json:"recorded_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
