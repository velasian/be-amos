package report

import (
	"gorm.io/gorm"
)

// Repository defines the interface for report-specific data queries.
// Uses JOIN queries and aggregations to fetch enriched data in minimal calls.
type Repository interface {
	// Excel Export
	GetAttendanceReport(startDate, endDate string, employeeID *uint) ([]AttendanceReportRow, error)

	// Dashboard Stats
	GetTotalActiveEmployees() (int64, error)
	GetDailyStats(startDate, endDate string) ([]DailyStatRow, error)
	GetDepartmentStats() ([]DepartmentStatRow, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new report repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// GetAttendanceReport fetches attendance records enriched with employee,
// department, position, and job site data using LEFT JOINs.
func (r *repository) GetAttendanceReport(startDate, endDate string, employeeID *uint) ([]AttendanceReportRow, error) {
	var rows []AttendanceReportRow

	query := r.db.Table("attendances a").
		Select(`a.id, a.employee_id, a.type, a.selfie_url,
				a.latitude, a.longitude, a.is_within_geofence, a.geofence_distance,
				a.is_late, a.late_minutes, a.recorded_at,
				COALESCE(e.nrp, '') as employee_nrp,
				COALESCE(e.name, '') as employee_name,
				COALESCE(d.name, '') as department_name,
				COALESCE(p.name, '') as position_name,
				COALESCE(j.name, '') as job_site_name`).
		Joins("LEFT JOIN employees e ON e.id = a.employee_id").
		Joins("LEFT JOIN departments d ON d.id = e.department_id").
		Joins("LEFT JOIN positions p ON p.id = e.position_id").
		Joins("LEFT JOIN job_sites j ON j.id = e.job_site_id").
		Where("DATE(a.recorded_at) >= ? AND DATE(a.recorded_at) <= ?", startDate, endDate)

	if employeeID != nil {
		query = query.Where("a.employee_id = ?", *employeeID)
	}

	err := query.Order("a.recorded_at ASC").Find(&rows).Error
	return rows, err
}

// GetTotalActiveEmployees counts all employees with status AKTIF.
func (r *repository) GetTotalActiveEmployees() (int64, error) {
	var count int64
	err := r.db.Table("employees").Where("status = ?", "AKTIF").Count(&count).Error
	return count, err
}

// GetDailyStats aggregates attendance data per day within a date range.
// Returns distinct present count and late count per day.
func (r *repository) GetDailyStats(startDate, endDate string) ([]DailyStatRow, error) {
	var rows []DailyStatRow

	err := r.db.Table("attendances").
		Select(`DATE(recorded_at) as stat_date,
				COUNT(DISTINCT employee_id) as present,
				COUNT(DISTINCT CASE WHEN is_late = true AND type = 'clock_in' THEN employee_id END) as late`).
		Where("DATE(recorded_at) >= ? AND DATE(recorded_at) <= ?", startDate, endDate).
		Group("DATE(recorded_at)").
		Order("stat_date ASC").
		Find(&rows).Error

	return rows, err
}

// GetDepartmentStats returns today's attendance breakdown per department.
// LEFT JOINs ensure departments with zero attendance still appear.
func (r *repository) GetDepartmentStats() ([]DepartmentStatRow, error) {
	var rows []DepartmentStatRow

	err := r.db.Table("employees e").
		Select(`COALESCE(d.name, 'Tanpa Departemen') as department,
				COUNT(DISTINCT e.id) as total,
				COUNT(DISTINCT CASE WHEN a.id IS NOT NULL THEN e.id END) as present,
				COUNT(DISTINCT CASE WHEN a.is_late = true AND a.type = 'clock_in' THEN e.id END) as late`).
		Joins("LEFT JOIN departments d ON d.id = e.department_id").
		Joins("LEFT JOIN attendances a ON a.employee_id = e.id AND DATE(a.recorded_at) = CURRENT_DATE").
		Where("e.status = ?", "AKTIF").
		Group("d.name").
		Order("department ASC").
		Find(&rows).Error

	return rows, err
}
