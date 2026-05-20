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
	GetAttendanceReport(startDate, endDate string, employeeID *uint) ([]Attendance, error)
	GetAttendanceList(filter AttendanceListFilter) ([]AttendanceListItem, int64, error)
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

func (r *repository) GetAttendanceReport(startDate, endDate string, employeeID *uint) ([]Attendance, error) {
	var list []Attendance
	query := r.db.Where("DATE(recorded_at) >= ? AND DATE(recorded_at) <= ?", startDate, endDate)
	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}
	err := query.Order("recorded_at ASC").Find(&list).Error
	return list, err
}

func (r *repository) GetAttendanceList(filter AttendanceListFilter) ([]AttendanceListItem, int64, error) {
	var list []AttendanceListItem
	var total int64

	base := r.db.Table("attendances a").
		Joins("LEFT JOIN employees e ON e.id = a.employee_id").
		Joins("LEFT JOIN departments d ON d.id = e.department_id").
		Joins("LEFT JOIN positions p ON p.id = e.position_id").
		Joins("LEFT JOIN job_sites j ON j.id = e.job_site_id")

	base = applyAttendanceListFilter(base, filter)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	if offset < 0 {
		offset = 0
	}

	err := base.
		Select(`a.id, a.session_id, a.employee_id, a.type, a.selfie_url,
			a.latitude, a.longitude, a.is_within_geofence, a.geofence_distance,
			a.is_late, a.late_minutes, a.recorded_at, a.created_at,
			COALESCE(e.nrp, '') AS employee_nrp,
			COALESCE(e.name, '') AS employee_name,
			COALESCE(d.name, '') AS department_name,
			COALESCE(p.name, '') AS position_name,
			COALESCE(j.name, '') AS job_site_name`).
		Order("a.recorded_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&list).Error

	return list, total, err
}

func applyAttendanceListFilter(query *gorm.DB, filter AttendanceListFilter) *gorm.DB {
	if filter.StartDate != "" {
		query = query.Where("DATE(a.recorded_at) >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("DATE(a.recorded_at) <= ?", filter.EndDate)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("e.name ILIKE ? OR e.nrp ILIKE ?", searchPattern, searchPattern)
	}
	if filter.Type != "" {
		query = query.Where("a.type = ?", filter.Type)
	}
	if filter.EmployeeID != nil {
		query = query.Where("a.employee_id = ?", *filter.EmployeeID)
	}
	if filter.DepartmentID != nil {
		query = query.Where("e.department_id = ?", *filter.DepartmentID)
	}
	if filter.JobSiteID != nil {
		query = query.Where("e.job_site_id = ?", *filter.JobSiteID)
	}
	return query
}
