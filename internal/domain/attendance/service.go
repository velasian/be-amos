package attendance

import (
	"amos-backend/internal/domain/employee"
	"amos-backend/internal/domain/master"
	"amos-backend/internal/domain/notification"
	"amos-backend/pkg/geofence"
	"amos-backend/pkg/storage"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"time"
)

const sessionExpiry = 5 * time.Minute // NFC session valid for 5 minutes

// Service defines the interface for attendance operations.
type Service interface {
	// IoT Device Management
	RegisterDevice(name string, jobSiteID uint, apiKey string) (*IoTDevice, error)
	GetAllDevices() ([]IoTDevice, error)

	// Attendance query endpoints for web and mobile clients
	GetActiveSession(userID uint) (*ActiveSessionResult, error)
	GetMyAttendances(userID uint, filter AttendanceListFilter) (*AttendancePaginatedResult, error)
	GetAllAttendances(filter AttendanceListFilter) (*AttendancePaginatedResult, error)

	// NFC Scan Flow (called by ESP32)
	ProcessNFCScan(ctx context.Context, nfcUID string, deviceID uint) (*ScanResult, error)

	// Mobile Verification Flow (called by employee app)
	VerifyAttendance(ctx context.Context, input VerifyInput) (*VerifyResult, error)

	// NFC Registration Flow
	BroadcastNFCUID(nfcUID string, deviceID uint) error
	AssignNFCToEmployee(employeeID uint, nfcUID string) error
}

// ScanResult is returned to the ESP32 after a successful NFC scan.
type ScanResult struct {
	SessionID    uint   `json:"session_id"`
	EmployeeName string `json:"employee_name"`
	EmployeeNRP  string `json:"employee_nrp"`
	Status       string `json:"status"` // "session_created"
	ExpiresAt    string `json:"expires_at"`
}

type VerifyInput struct {
	SessionID  uint
	UserID     uint // From JWT — will be resolved to employee
	Latitude   float64
	Longitude  float64
	Selfie     multipart.File
	SelfieSize int64
	SelfieName string
}

// VerifyResult is returned to the mobile app after verification.
type VerifyResult struct {
	AttendanceID     uint    `json:"attendance_id"`
	Type             string  `json:"type"` // clock_in or clock_out
	IsWithinGeofence bool    `json:"is_within_geofence"`
	DistanceMeters   float64 `json:"distance_meters"`
	IsLate           bool    `json:"is_late"`
	LateMinutes      int     `json:"late_minutes"`
	RecordedAt       string  `json:"recorded_at"`
}

// workStartHour and workStartMinute define the start of work for late calculation.
const (
	workStartHour   = 8
	workStartMinute = 0
)

type service struct {
	repo         Repository
	employeeRepo employee.Repository
	masterRepo   master.Repository
	notifService notification.Service
	storage      storage.Client
	sseBroker    *SSEBroker
}

// NewService creates a new attendance service.
func NewService(repo Repository, employeeRepo employee.Repository, masterRepo master.Repository, notifService notification.Service, storageClient storage.Client, broker *SSEBroker) Service {
	return &service{
		repo:         repo,
		employeeRepo: employeeRepo,
		masterRepo:   masterRepo,
		notifService: notifService,
		storage:      storageClient,
		sseBroker:    broker,
	}
}

// RegisterDevice creates a new IoT device entry.
func (s *service) RegisterDevice(name string, jobSiteID uint, apiKey string) (*IoTDevice, error) {
	device := &IoTDevice{
		Name:     name,
		APIKey:   apiKey,
		IsActive: true,
	}
	if jobSiteID > 0 {
		device.JobSiteID = &jobSiteID
	}

	if err := s.repo.CreateDevice(device); err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}
	return device, nil
}

func (s *service) GetAllDevices() ([]IoTDevice, error) {
	return s.repo.GetAllDevices()
}

func (s *service) GetActiveSession(userID uint) (*ActiveSessionResult, error) {
	emp, err := s.employeeRepo.FindByUserID(userID)
	if err != nil || emp.ID == 0 {
		return nil, fmt.Errorf("no employee linked to this user account")
	}

	_ = s.repo.ExpireStaleSessions()

	session, err := s.repo.FindPendingSessionByEmployee(emp.ID)
	if err != nil || session.ID == 0 {
		return &ActiveSessionResult{
			HasActiveSession: false,
			Session:          nil,
		}, nil
	}

	return &ActiveSessionResult{
		HasActiveSession: true,
		Session:          session,
	}, nil
}

func (s *service) GetMyAttendances(userID uint, filter AttendanceListFilter) (*AttendancePaginatedResult, error) {
	emp, err := s.employeeRepo.FindByUserID(userID)
	if err != nil || emp.ID == 0 {
		return nil, fmt.Errorf("no employee linked to this user account")
	}

	filter.EmployeeID = &emp.ID
	return s.GetAllAttendances(filter)
}

func (s *service) GetAllAttendances(filter AttendanceListFilter) (*AttendancePaginatedResult, error) {
	filter = normalizeAttendanceListFilter(filter)

	list, total, err := s.repo.GetAttendanceList(filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	return &AttendancePaginatedResult{
		Data:       list,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func normalizeAttendanceListFilter(filter AttendanceListFilter) AttendanceListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	return filter
}

// ProcessNFCScan handles the core NFC tap flow:
// 1. Find employee by NFC UID
// 2. Expire any stale sessions
// 3. Check for duplicate pending sessions
// 4. Create new attendance session (pending, 5-min expiry)
// 5. Send FCM push notification to employee's phone
func (s *service) ProcessNFCScan(ctx context.Context, nfcUID string, deviceID uint) (*ScanResult, error) {
	// 1. Find employee by NFC UID
	emp, err := s.employeeRepo.FindByNfcUID(nfcUID)
	if err != nil || emp.ID == 0 {
		return nil, fmt.Errorf("NFC UID '%s' not registered to any employee", nfcUID)
	}

	// 2. Expire stale sessions
	_ = s.repo.ExpireStaleSessions()

	// 3. Check for existing pending session (prevent duplicate taps)
	existing, _ := s.repo.FindPendingSessionByEmployee(emp.ID)
	if existing != nil && existing.ID != 0 {
		return &ScanResult{
			SessionID:    existing.ID,
			EmployeeName: emp.Name,
			EmployeeNRP:  emp.NRP,
			Status:       "session_exists",
			ExpiresAt:    existing.ExpiresAt.Format(time.RFC3339),
		}, nil
	}

	// 4. Create new attendance session
	now := time.Now()
	session := &AttendanceSession{
		EmployeeID:  emp.ID,
		IoTDeviceID: &deviceID,
		Status:      "pending",
		ScannedAt:   now,
		ExpiresAt:   now.Add(sessionExpiry),
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create attendance session: %w", err)
	}

	log.Printf("[IOT] Session %d created for %s (NRP: %s, expires: %s)",
		session.ID, emp.Name, emp.NRP, session.ExpiresAt.Format("15:04:05"))

	// 5. Send FCM push notification to employee's phone
	if emp.UserID != nil {
		go func() {
			notifCtx := context.Background()
			title := "Presensi Terdeteksi"
			msg := fmt.Sprintf("Halo %s, kartu NFC Anda terdeteksi. Buka aplikasi untuk verifikasi selfie dalam %d menit.", emp.Name, int(sessionExpiry.Minutes()))

			if err := s.notifService.SendToUser(notifCtx, *emp.UserID, "attendance_required", title, msg); err != nil {
				log.Printf("[IOT] Failed to send FCM for session %d: %v", session.ID, err)
			}
		}()
	}

	return &ScanResult{
		SessionID:    session.ID,
		EmployeeName: emp.Name,
		EmployeeNRP:  emp.NRP,
		Status:       "session_created",
		ExpiresAt:    session.ExpiresAt.Format(time.RFC3339),
	}, nil
}

// VerifyAttendance handles the mobile app verification flow:
// 1. Validate session (pending, not expired, belongs to employee)
// 2. Upload selfie to MinIO
// 3. Determine clock_in or clock_out
// 4. Run geofence check against job site
// 5. Calculate lateness (for clock_in only)
// 6. Create final attendance record
// 7. Mark session as verified
func (s *service) VerifyAttendance(ctx context.Context, input VerifyInput) (*VerifyResult, error) {
	// 0. Resolve employee from user ID
	emp, err := s.employeeRepo.FindByUserID(input.UserID)
	if err != nil || emp.ID == 0 {
		return nil, fmt.Errorf("no employee linked to this user account")
	}
	employeeID := emp.ID

	// 1. Validate session
	session, err := s.repo.FindSessionByID(input.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}
	if session.EmployeeID != employeeID {
		return nil, fmt.Errorf("session does not belong to this employee")
	}
	if session.Status != "pending" {
		return nil, fmt.Errorf("session already %s", session.Status)
	}
	if time.Now().After(session.ExpiresAt) {
		session.Status = "expired"
		_ = s.repo.UpdateSession(session)
		return nil, fmt.Errorf("session expired, please tap NFC again")
	}

	// 2. Run geofence check before storing the selfie.
	geoResult, err := s.checkSessionGeofence(session, input.Latitude, input.Longitude)
	if err != nil {
		return nil, err
	}
	if !geoResult.IsWithinFence {
		return nil, fmt.Errorf("outside geofence: distance %.0f meters from allowed job site", geoResult.Distance)
	}

	// 3. Upload selfie to MinIO
	selfieURL := ""
	if input.Selfie != nil {
		objectName := storage.GenerateObjectName("attendances", employeeID, "selfie", input.SelfieName)
		url, err := s.storage.Upload(ctx, objectName, input.Selfie, input.SelfieSize, "image/jpeg")
		if err != nil {
			log.Printf("[ATTENDANCE] Selfie upload failed: %v", err)
		} else {
			selfieURL = url
		}
	}

	// 4. Determine clock_in or clock_out
	todayRecords, _ := s.repo.GetTodayAttendanceByEmployee(employeeID)
	attendanceType := "clock_in"
	if len(todayRecords) > 0 {
		// If last record is clock_in, this is clock_out
		last := todayRecords[len(todayRecords)-1]
		if last.Type == "clock_in" {
			attendanceType = "clock_out"
		}
	}

	// 5. Calculate lateness (clock_in only)
	now := time.Now()
	isLate := false
	lateMinutes := 0
	if attendanceType == "clock_in" {
		workStart := time.Date(now.Year(), now.Month(), now.Day(), workStartHour, workStartMinute, 0, 0, now.Location())
		if now.After(workStart) {
			isLate = true
			lateMinutes = int(now.Sub(workStart).Minutes())
		}
	}

	// 6. Create final attendance record
	attendance := &Attendance{
		SessionID:        session.ID,
		EmployeeID:       employeeID,
		Type:             attendanceType,
		SelfieURL:        selfieURL,
		Latitude:         input.Latitude,
		Longitude:        input.Longitude,
		IsWithinGeofence: geoResult.IsWithinFence,
		GeofenceDistance: geoResult.Distance,
		IsLate:           isLate,
		LateMinutes:      lateMinutes,
		RecordedAt:       now,
	}

	if err := s.repo.CreateAttendance(attendance); err != nil {
		return nil, fmt.Errorf("failed to save attendance: %w", err)
	}

	// 7. Mark session as verified
	session.Status = "verified"
	_ = s.repo.UpdateSession(session)

	log.Printf("[ATTENDANCE] %s recorded for employee %d (session %d, geofence: %v, late: %v)",
		attendanceType, employeeID, session.ID, geoResult.IsWithinFence, isLate)

	return &VerifyResult{
		AttendanceID:     attendance.ID,
		Type:             attendanceType,
		IsWithinGeofence: geoResult.IsWithinFence,
		DistanceMeters:   geoResult.Distance,
		IsLate:           isLate,
		LateMinutes:      lateMinutes,
		RecordedAt:       now.Format(time.RFC3339),
	}, nil
}

func (s *service) checkSessionGeofence(session *AttendanceSession, latitude, longitude float64) (geofence.Result, error) {
	if session.IoTDeviceID == nil {
		return geofence.Result{}, fmt.Errorf("attendance session has no IoT device assigned")
	}

	device, err := s.repo.FindDeviceByID(*session.IoTDeviceID)
	if err != nil || device.ID == 0 {
		return geofence.Result{}, fmt.Errorf("IoT device not found")
	}
	if device.JobSiteID == nil {
		return geofence.Result{}, fmt.Errorf("IoT device has no job site configured")
	}

	jobSite, err := s.masterRepo.FindJobSiteByID(*device.JobSiteID)
	if err != nil {
		return geofence.Result{}, fmt.Errorf("job site not found")
	}

	return geofence.Check(
		jobSite.Latitude,
		jobSite.Longitude,
		latitude,
		longitude,
		jobSite.RadiusMeters,
	), nil
}

// BroadcastNFCUID is called by ESP32 in registration mode.
// It broadcasts the detected NFC UID to all admin SSE listeners.
func (s *service) BroadcastNFCUID(nfcUID string, deviceID uint) error {
	device, err := s.repo.FindDeviceByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found")
	}

	deviceName := device.Name
	log.Printf("[NFC-REG] UID '%s' detected on device '%s' (ID: %d), broadcasting to %d listeners",
		nfcUID, deviceName, deviceID, s.sseBroker.ActiveCount())

	s.sseBroker.PublishNFCDetected(nfcUID, deviceID, deviceName)
	return nil
}

// AssignNFCToEmployee assigns an NFC UID to an employee.
// Called by admin after seeing the UID appear on their SSE stream.
func (s *service) AssignNFCToEmployee(employeeID uint, nfcUID string) error {
	emp, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		return fmt.Errorf("employee not found")
	}

	// Check if NFC UID is already assigned
	existing, _ := s.employeeRepo.FindByNfcUID(nfcUID)
	if existing != nil && existing.ID != 0 && existing.ID != employeeID {
		return fmt.Errorf("NFC UID already assigned to employee '%s' (NRP: %s)", existing.Name, existing.NRP)
	}

	emp.NfcUID = &nfcUID
	if err := s.employeeRepo.UpdateEmployee(emp); err != nil {
		return fmt.Errorf("failed to assign NFC: %w", err)
	}

	log.Printf("[NFC-REG] UID '%s' assigned to %s (NRP: %s)", nfcUID, emp.Name, emp.NRP)
	return nil
}
