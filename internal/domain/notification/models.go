package notification

import (
	"time"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Type      string    `gorm:"type:varchar(50)" json:"type"` // attendance_required, leave_approved
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	Message   string    `gorm:"type:text" json:"message"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
