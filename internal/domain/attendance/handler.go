package attendance

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for attendance and IoT operations.
type Handler struct {
	service   Service
	sseBroker *SSEBroker
}

// NewHandler creates a new attendance handler.
func NewHandler(service Service, broker *SSEBroker) *Handler {
	return &Handler{service: service, sseBroker: broker}
}

// --- IoT Endpoints (authenticated via X-API-Key) ---

// ScanNFC processes an NFC card tap from an ESP32 device.
// POST /iot/scan
// Body: { "nfc_uid": "AB:CD:EF:12" }
func (h *Handler) ScanNFC(c *gin.Context) {
	deviceID, exists := c.Get("iotDeviceID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Device not authenticated",
		})
		return
	}

	var req struct {
		NfcUID string `json:"nfc_uid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "nfc_uid is required",
		})
		return
	}

	result, err := h.service.ProcessNFCScan(c.Request.Context(), req.NfcUID, deviceID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "NFC scan processed",
		"data":    result,
	})
}

// --- Mobile Endpoints (authenticated via JWT) ---

// VerifyAttendance processes selfie + GPS verification from the mobile app.
// POST /attendances/verify (multipart/form-data)
// Fields: session_id, latitude, longitude, selfie (file)
func (h *Handler) VerifyAttendance(c *gin.Context) {
	userID := c.GetUint("userID")

	sessionIDStr := c.PostForm("session_id")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil || sessionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Valid session_id is required",
		})
		return
	}

	lat, err := strconv.ParseFloat(c.PostForm("latitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Valid latitude is required",
		})
		return
	}

	lon, err := strconv.ParseFloat(c.PostForm("longitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Valid longitude is required",
		})
		return
	}

	// Build input — service resolves employee from userID
	var input VerifyInput
	input.SessionID = uint(sessionID)
	input.UserID = userID
	input.Latitude = lat
	input.Longitude = lon

	fileHeader, err := c.FormFile("selfie")
	if err == nil {
		f, _ := fileHeader.Open()
		input.Selfie = f
		input.SelfieSize = fileHeader.Size
		input.SelfieName = fileHeader.Filename
		defer f.Close()
	}

	result, err := h.service.VerifyAttendance(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Attendance verified successfully",
		"data":    result,
	})
}

// GetActiveSession returns the authenticated employee's pending NFC session.
// GET /attendances/session
func (h *Handler) GetActiveSession(c *gin.Context) {
	userID := c.GetUint("userID")

	result, err := h.service.GetActiveSession(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Active attendance session retrieved",
		"data":    result,
	})
}

// GetMyAttendances returns attendance history for the authenticated employee.
// GET /attendances/me?page=1&limit=20&start_date=2026-05-01&end_date=2026-05-31
func (h *Handler) GetMyAttendances(c *gin.Context) {
	filter, err := parseAttendanceListFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.GetMyAttendances(c.GetUint("userID"), filter)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Attendance history retrieved",
		"data":    result,
	})
}

// GetAllAttendances returns paginated attendance data for HR monitoring.
// GET /attendances?page=1&limit=20&search=budi&type=clock_in&department_id=1
func (h *Handler) GetAllAttendances(c *gin.Context) {
	filter, err := parseAttendanceListFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.GetAllAttendances(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Attendance data retrieved",
		"data":    result,
	})
}

// --- Admin Endpoints (authenticated via JWT) ---

// RegisterDevice registers a new IoT device with an auto-generated API key.
// POST /iot/devices
func (h *Handler) RegisterDevice(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		JobSiteID uint   `json:"job_site_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Generate secure API key
	apiKey := "iot_" + uuid.New().String()

	device, err := h.service.RegisterDevice(req.Name, req.JobSiteID, apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Return API key only once at creation time
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Device registered. Save the API key — it will not be shown again.",
		"data": gin.H{
			"id":          device.ID,
			"name":        device.Name,
			"job_site_id": device.JobSiteID,
			"api_key":     apiKey,
			"is_active":   device.IsActive,
		},
	})
}

// GetAllDevices lists all registered IoT devices (API keys hidden).
// GET /iot/devices
func (h *Handler) GetAllDevices(c *gin.Context) {
	devices, err := h.service.GetAllDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "IoT devices retrieved",
		"data":    devices,
	})
}

// --- NFC Registration Endpoints ---

// ListenNFC provides an SSE stream for admin to receive real-time NFC detections.
// GET /iot/listen
func (h *Handler) ListenNFC(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Subscribe to broker
	subID := uuid.New().String()
	ch := h.sseBroker.Subscribe(subID)
	defer h.sseBroker.Unsubscribe(subID)

	// Send initial heartbeat
	c.SSEvent("connected", fmt.Sprintf(`{"subscriber_id":"%s","listeners":%d}`, subID, h.sseBroker.ActiveCount()))
	c.Writer.Flush()

	// Stream events until client disconnects
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			c.SSEvent(event.Event, event.Data)
			c.Writer.Flush()
		}
	}
}

// ReportNFCUID is called by ESP32 in registration mode to report a detected NFC UID.
// POST /iot/assign { "nfc_uid": "AB:CD:EF:12" }
func (h *Handler) ReportNFCUID(c *gin.Context) {
	deviceID, exists := c.Get("iotDeviceID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Device not authenticated",
		})
		return
	}

	var req struct {
		NfcUID string `json:"nfc_uid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "nfc_uid is required",
		})
		return
	}

	if err := h.service.BroadcastNFCUID(req.NfcUID, deviceID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "NFC UID broadcasted to admin listeners",
	})
}

// AssignNFC assigns an NFC UID to an employee (admin action after seeing UID on SSE).
// POST /iot/assign-employee { "employee_id": 1, "nfc_uid": "AB:CD:EF:12" }
func (h *Handler) AssignNFC(c *gin.Context) {
	var req struct {
		EmployeeID uint   `json:"employee_id" binding:"required"`
		NfcUID     string `json:"nfc_uid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.AssignNFCToEmployee(req.EmployeeID, req.NfcUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "NFC UID assigned to employee successfully",
	})
}

func parseAttendanceListFilter(c *gin.Context) (AttendanceListFilter, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := AttendanceListFilter{
		Page:      page,
		Limit:     limit,
		Search:    c.Query("search"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Type:      c.Query("type"),
	}

	if filter.StartDate != "" {
		if _, err := time.Parse("2006-01-02", filter.StartDate); err != nil {
			return filter, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
		}
	}
	if filter.EndDate != "" {
		if _, err := time.Parse("2006-01-02", filter.EndDate); err != nil {
			return filter, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
		}
	}
	if filter.Type != "" && filter.Type != "clock_in" && filter.Type != "clock_out" {
		return filter, fmt.Errorf("invalid type, expected clock_in or clock_out")
	}

	employeeID, err := parseOptionalUintQuery(c, "employee_id")
	if err != nil {
		return filter, err
	}
	filter.EmployeeID = employeeID

	departmentID, err := parseOptionalUintQuery(c, "department_id")
	if err != nil {
		return filter, err
	}
	filter.DepartmentID = departmentID

	jobSiteID, err := parseOptionalUintQuery(c, "job_site_id")
	if err != nil {
		return filter, err
	}
	filter.JobSiteID = jobSiteID

	return filter, nil
}

func parseOptionalUintQuery(c *gin.Context, key string) (*uint, error) {
	value := c.Query(key)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		return nil, fmt.Errorf("invalid %s", key)
	}

	result := uint(parsed)
	return &result, nil
}
