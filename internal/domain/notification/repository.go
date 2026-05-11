package notification

import (
	"gorm.io/gorm"
)

// Repository defines the interface for notification persistence.
type Repository interface {
	Create(n *Notification) error
	GetByUserID(userID uint, page, limit int) ([]Notification, int64, error)
	GetUnreadCount(userID uint) (int64, error)
	MarkAsRead(id, userID uint) error
	MarkAllAsRead(userID uint) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new notification repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(n *Notification) error {
	return r.db.Create(n).Error
}

func (r *repository) GetByUserID(userID uint, page, limit int) ([]Notification, int64, error) {
	var notifs []Notification
	var total int64

	query := r.db.Model(&Notification{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifs).Error
	return notifs, total, err
}

func (r *repository) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *repository) MarkAsRead(id, userID uint) error {
	return r.db.Model(&Notification{}).Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *repository) MarkAllAsRead(userID uint) error {
	return r.db.Model(&Notification{}).Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
