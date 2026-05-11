package leave

import (
	"time"
)

type LeaveRequest struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EmployeeID   uint      `gorm:"not null" json:"employee_id"`
	LeaveType    string    `gorm:"type:varchar(50)" json:"leave_type"`
	StartDate    time.Time `gorm:"type:date" json:"start_date"`
	EndDate      time.Time `gorm:"type:date" json:"end_date"`
	Reason       string    `gorm:"type:text" json:"reason"`
	Status       string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, approved, rejected
	ApprovedByID *uint     `json:"approved_by_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
