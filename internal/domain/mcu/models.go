package mcu

import (
	"time"
)

type MCUSchedule struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EmployeeID   uint      `gorm:"not null" json:"employee_id"`
	ScheduleDate time.Time `gorm:"type:date" json:"schedule_date"`
	Status       string    `gorm:"type:varchar(50);default:'scheduled'" json:"status"`
	Notes        string    `gorm:"type:text" json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
