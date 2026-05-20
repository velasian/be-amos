package report

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for report operations.
type Handler struct {
	service Service
}

// NewHandler creates a new report handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetStats returns dashboard statistics for attendance.
// GET /reports/stats?period_start=2026-04-01&period_end=2026-04-30
// Defaults to the last 30 days if no period params are given.
func (h *Handler) GetStats(c *gin.Context) {
	now := time.Now()
	periodStart := c.DefaultQuery("period_start", now.AddDate(0, 0, -30).Format("2006-01-02"))
	periodEnd := c.DefaultQuery("period_end", now.Format("2006-01-02"))

	// Validate date formats
	if _, err := time.Parse("2006-01-02", periodStart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid period_start format, expected YYYY-MM-DD",
		})
		return
	}
	if _, err := time.Parse("2006-01-02", periodEnd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid period_end format, expected YYYY-MM-DD",
		})
		return
	}

	stats, err := h.service.GetDashboardStats(periodStart, periodEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Dashboard statistics",
		"data":    stats,
	})
}

// ExportAttendance exports attendance records to an Excel file.
// GET /reports/attendance/export?start_date=2026-01-01&end_date=2026-01-31&employee_id=2
func (h *Handler) ExportAttendance(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required (format: YYYY-MM-DD)",
		})
		return
	}

	// Validate date format
	parsedStart, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format, expected YYYY-MM-DD",
		})
		return
	}

	parsedEnd, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format, expected YYYY-MM-DD",
		})
		return
	}

	// Validate date range
	if parsedEnd.Before(parsedStart) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "end_date must be after or equal to start_date",
		})
		return
	}

	// Optional employee filter
	var employeeID *uint
	if empIDStr := c.Query("employee_id"); empIDStr != "" {
		id, err := strconv.ParseUint(empIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "invalid employee_id",
			})
			return
		}
		eid := uint(id)
		employeeID = &eid
	}

	buf, filename, err := h.service.ExportAttendanceExcel(startDate, endDate, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
