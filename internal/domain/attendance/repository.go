package attendance

import (
	"gorm.io/gorm"
)

// Repository defines the interface for attendance persistence.
type Repository interface {
	// IoT Devices
	CreateDevice(d *IoTDevice) error
	FindDeviceByID(id uint) (*IoTDevice, error)
	FindDeviceByAPIKey(apiKey string) (*IoTDevice, error)
	GetAllDevices() ([]IoTDevice, error)
	UpdateDevice(d *IoTDevice) error

	// Attendance Sessions
	CreateSession(s *AttendanceSession) error
	FindSessionByID(id uint) (*AttendanceSession, error)
	FindPendingSessionByEmployee(employeeID uint) (*AttendanceSession, error)
	UpdateSession(s *AttendanceSession) error
	ExpireStaleSessions() error

	// Attendances
	CreateAttendance(a *Attendance) error
	GetTodayAttendanceByEmployee(employeeID uint) ([]Attendance, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new attendance repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// --- IoT Devices ---

func (r *repository) CreateDevice(d *IoTDevice) error {
	return r.db.Create(d).Error
}

func (r *repository) FindDeviceByID(id uint) (*IoTDevice, error) {
	var d IoTDevice
	err := r.db.First(&d, id).Error
	return &d, err
}

func (r *repository) FindDeviceByAPIKey(apiKey string) (*IoTDevice, error) {
	var d IoTDevice
	err := r.db.Where("api_key = ?", apiKey).First(&d).Error
	return &d, err
}

func (r *repository) GetAllDevices() ([]IoTDevice, error) {
	var devices []IoTDevice
	err := r.db.Find(&devices).Error
	return devices, err
}

func (r *repository) UpdateDevice(d *IoTDevice) error {
	return r.db.Save(d).Error
}

// --- Attendance Sessions ---

func (r *repository) CreateSession(s *AttendanceSession) error {
	return r.db.Create(s).Error
}

func (r *repository) FindSessionByID(id uint) (*AttendanceSession, error) {
	var s AttendanceSession
	err := r.db.First(&s, id).Error
	return &s, err
}

func (r *repository) FindPendingSessionByEmployee(employeeID uint) (*AttendanceSession, error) {
	var s AttendanceSession
	err := r.db.Where("employee_id = ? AND status = ? AND expires_at > NOW()", employeeID, "pending").
		Order("created_at DESC").First(&s).Error
	return &s, err
}

func (r *repository) UpdateSession(s *AttendanceSession) error {
	return r.db.Save(s).Error
}

func (r *repository) ExpireStaleSessions() error {
	return r.db.Model(&AttendanceSession{}).
		Where("status = ? AND expires_at < NOW()", "pending").
		Update("status", "expired").Error
}

// --- Attendances ---

func (r *repository) CreateAttendance(a *Attendance) error {
	return r.db.Create(a).Error
}

func (r *repository) GetTodayAttendanceByEmployee(employeeID uint) ([]Attendance, error) {
	var list []Attendance
	err := r.db.Where("employee_id = ? AND DATE(recorded_at) = CURRENT_DATE", employeeID).
		Order("recorded_at ASC").Find(&list).Error
	return list, err
}
