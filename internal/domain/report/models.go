package report

import "time"

// AttendanceReportRow represents a single row in the attendance Excel report,
// enriched with employee and organizational data via JOIN queries.
type AttendanceReportRow struct {
	// Attendance Data
	ID               uint      `json:"id"`
	EmployeeID       uint      `json:"employee_id"`
	Type             string    `json:"type"`
	SelfieURL        string    `json:"selfie_url"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	IsWithinGeofence bool      `json:"is_within_geofence"`
	GeofenceDistance  float64   `json:"geofence_distance"`
	IsLate           bool      `json:"is_late"`
	LateMinutes      int       `json:"late_minutes"`
	RecordedAt       time.Time `json:"recorded_at"`

	// Employee Data (from JOIN)
	EmployeeNRP    string `json:"employee_nrp"`
	EmployeeName   string `json:"employee_name"`
	DepartmentName string `json:"department_name"`
	PositionName   string `json:"position_name"`
	JobSiteName    string `json:"job_site_name"`
}

// --- Dashboard Stats ---

// DashboardStats is the complete response for the /reports/stats API.
type DashboardStats struct {
	Today           TodayStats       `json:"today"`
	Period          PeriodStats      `json:"period"`
	DepartmentStats []DepartmentStat `json:"department_stats"`
}

// TodayStats holds today's attendance snapshot.
type TodayStats struct {
	Date           string  `json:"date"`
	TotalEmployees int64   `json:"total_employees"`
	Present        int64   `json:"present"`
	Absent         int64   `json:"absent"`
	Late           int64   `json:"late"`
	OnTime         int64   `json:"on_time"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// PeriodStats holds attendance trend data over a date range.
type PeriodStats struct {
	StartDate  string      `json:"start_date"`
	EndDate    string      `json:"end_date"`
	DailyStats []DailyStat `json:"daily_stats"`
}

// DailyStat holds aggregated attendance data for a single day.
type DailyStat struct {
	Date    string `json:"date"`
	Present int64  `json:"present"`
	Late    int64  `json:"late"`
	OnTime  int64  `json:"on_time"`
}

// DepartmentStat holds attendance breakdown per department.
type DepartmentStat struct {
	Department string `json:"department"`
	Total      int64  `json:"total"`
	Present    int64  `json:"present"`
	Late       int64  `json:"late"`
}

// --- Internal DB scan structs ---

// DailyStatRow is the raw database row for daily stats aggregation.
type DailyStatRow struct {
	StatDate time.Time `gorm:"column:stat_date"`
	Present  int64     `gorm:"column:present"`
	Late     int64     `gorm:"column:late"`
}

// DepartmentStatRow is the raw database row for department stats aggregation.
type DepartmentStatRow struct {
	Department string `gorm:"column:department"`
	Total      int64  `gorm:"column:total"`
	Present    int64  `gorm:"column:present"`
	Late       int64  `gorm:"column:late"`
}
