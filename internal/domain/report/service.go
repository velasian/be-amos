package report

import (
	"bytes"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// Service handles attendance report generation and dashboard statistics.
type Service interface {
	ExportAttendanceExcel(startDate, endDate string, employeeID *uint) (*bytes.Buffer, string, error)
	GetDashboardStats(periodStart, periodEnd string) (*DashboardStats, error)
}

type service struct {
	repo Repository
}

// NewService creates a new report service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// GetDashboardStats assembles all dashboard statistics in a single call.
// Combines today's snapshot, period trends, and department breakdown.
func (s *service) GetDashboardStats(periodStart, periodEnd string) (*DashboardStats, error) {
	today := time.Now().Format("2006-01-02")

	// 1. Get total active employees
	totalEmployees, err := s.repo.GetTotalActiveEmployees()
	if err != nil {
		return nil, fmt.Errorf("failed to get employee count: %w", err)
	}

	// 2. Get today's stats (reuse daily stats with today's date)
	todayRows, err := s.repo.GetDailyStats(today, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get today stats: %w", err)
	}

	var todayPresent, todayLate int64
	if len(todayRows) > 0 {
		todayPresent = todayRows[0].Present
		todayLate = todayRows[0].Late
	}

	todayAbsent := totalEmployees - todayPresent
	if todayAbsent < 0 {
		todayAbsent = 0
	}

	var attendanceRate float64
	if totalEmployees > 0 {
		attendanceRate = float64(todayPresent) / float64(totalEmployees) * 100
		// Round to 2 decimal places
		attendanceRate = float64(int(attendanceRate*100)) / 100
	}

	todayStats := TodayStats{
		Date:           today,
		TotalEmployees: totalEmployees,
		Present:        todayPresent,
		Absent:         todayAbsent,
		Late:           todayLate,
		OnTime:         todayPresent - todayLate,
		AttendanceRate: attendanceRate,
	}

	// 3. Get period trend data
	periodRows, err := s.repo.GetDailyStats(periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get period stats: %w", err)
	}

	dailyStats := make([]DailyStat, 0, len(periodRows))
	for _, row := range periodRows {
		dailyStats = append(dailyStats, DailyStat{
			Date:    row.StatDate.Format("2006-01-02"),
			Present: row.Present,
			Late:    row.Late,
			OnTime:  row.Present - row.Late,
		})
	}

	// 4. Get department breakdown (today)
	deptRows, err := s.repo.GetDepartmentStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get department stats: %w", err)
	}

	departmentStats := make([]DepartmentStat, 0, len(deptRows))
	for _, row := range deptRows {
		departmentStats = append(departmentStats, DepartmentStat{
			Department: row.Department,
			Total:      row.Total,
			Present:    row.Present,
			Late:       row.Late,
		})
	}

	return &DashboardStats{
		Today: todayStats,
		Period: PeriodStats{
			StartDate:  periodStart,
			EndDate:    periodEnd,
			DailyStats: dailyStats,
		},
		DepartmentStats: departmentStats,
	}, nil
}

// ExportAttendanceExcel generates a styled Excel report of attendance records.
// Uses a single JOIN query for all data — no N+1 problem.
// Returns the buffer, filename, and error.
func (s *service) ExportAttendanceExcel(startDate, endDate string, employeeID *uint) (*bytes.Buffer, string, error) {
	// 1. Fetch enriched attendance data (single JOIN query)
	records, err := s.repo.GetAttendanceReport(startDate, endDate, employeeID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch attendance data: %w", err)
	}

	// 2. Create Excel file
	f := excelize.NewFile()
	defer f.Close()
	sheetName := "Laporan Presensi"
	f.SetSheetName("Sheet1", sheetName)

	// 3. Title row
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "1B5E20"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellValue(sheetName, "A1", fmt.Sprintf("Laporan Presensi: %s s/d %s", startDate, endDate))
	f.MergeCell(sheetName, "A1", "M1")
	f.SetCellStyle(sheetName, "A1", "M1", titleStyle)
	f.SetRowHeight(sheetName, 1, 30)

	// 4. Header row (row 3)
	headers := []string{
		"No", "NRP", "Nama Pegawai", "Departemen", "Jabatan", "Lokasi Kerja",
		"Tipe", "Tanggal", "Jam",
		"Dalam Geofence", "Jarak (m)", "Terlambat", "Menit Terlambat",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"1B5E20"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, 3, 24)

	// 5. Reusable styles
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})

	lateStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "D32F2F"},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})

	centerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})

	// 6. Data rows
	for i, rec := range records {
		row := i + 4 // Data starts at row 4

		geofenceStatus := "❌ Diluar"
		if rec.IsWithinGeofence {
			geofenceStatus = "✅ Dalam"
		}

		lateStatus := "-"
		if rec.IsLate {
			lateStatus = "⚠️ Terlambat"
		}

		typeLabel := "Masuk"
		if rec.Type == "clock_out" {
			typeLabel = "Pulang"
		}

		rowData := []interface{}{
			i + 1,
			rec.EmployeeNRP,
			rec.EmployeeName,
			rec.DepartmentName,
			rec.PositionName,
			rec.JobSiteName,
			typeLabel,
			rec.RecordedAt.Format("02-01-2006"),
			rec.RecordedAt.Format("15:04:05"),
			geofenceStatus,
			fmt.Sprintf("%.0f", rec.GeofenceDistance),
			lateStatus,
			rec.LateMinutes,
		}

		for j, val := range rowData {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheetName, cell, val)

			switch {
			case j == 11 && rec.IsLate:
				f.SetCellStyle(sheetName, cell, cell, lateStyle)
			case j == 0 || j == 6 || j == 8 || j == 9 || j == 12:
				f.SetCellStyle(sheetName, cell, cell, centerStyle)
			default:
				f.SetCellStyle(sheetName, cell, cell, dataStyle)
			}
		}
	}

	// 7. Set column widths
	colWidths := map[string]float64{
		"A": 5, "B": 14, "C": 28, "D": 20, "E": 20, "F": 22,
		"G": 10, "H": 14, "I": 12,
		"J": 16, "K": 12, "L": 16, "M": 16,
	}
	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	// 8. Freeze header row
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      3,
		TopLeftCell: "A4",
		ActivePane:  "bottomLeft",
	})

	// 9. Summary row
	summaryRow := len(records) + 5
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "1B5E20"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("Total Records: %d", len(records)))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), summaryStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), fmt.Sprintf("Generated: %s", time.Now().Format("02-01-2006 15:04")))

	// 10. Write to buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate Excel: %w", err)
	}

	filename := fmt.Sprintf("Laporan_Presensi_%s_%s.xlsx", startDate, endDate)
	return buf, filename, nil
}
