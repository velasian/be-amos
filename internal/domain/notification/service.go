package notification

import (
	"amos-backend/internal/domain/auth"
	"amos-backend/pkg/firebase"
	"context"
	"log"
	"strconv"
)

// Service defines the interface for notification operations.
type Service interface {
	// Push + Inbox combined: sends FCM push AND saves to inbox
	SendToUser(ctx context.Context, userID uint, notifType, title, message string) error
	SendToMultipleUsers(ctx context.Context, userIDs []uint, notifType, title, message string) error

	// Inbox only
	GetInbox(userID uint, page, limit int) (*InboxResult, error)
	GetUnreadCount(userID uint) (int64, error)
	MarkAsRead(id, userID uint) error
	MarkAllAsRead(userID uint) error
}

type service struct {
	repo     Repository
	authRepo auth.Repository
	fcm      *firebase.Client
}

// NewService creates a new notification service.
func NewService(repo Repository, authRepo auth.Repository, fcm *firebase.Client) Service {
	return &service{
		repo:     repo,
		authRepo: authRepo,
		fcm:      fcm,
	}
}

// InboxResult wraps paginated notification data.
type InboxResult struct {
	Data        []Notification `json:"data"`
	Total       int64          `json:"total"`
	UnreadCount int64          `json:"unread_count"`
	Page        int            `json:"page"`
	Limit       int            `json:"limit"`
}

// SendToUser sends a push notification AND saves it to the user's inbox.
func (s *service) SendToUser(ctx context.Context, userID uint, notifType, title, message string) error {
	// 1. Save to inbox (always, even if FCM fails)
	notif := &Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
	}
	if err := s.repo.Create(notif); err != nil {
		return err
	}

	// 2. Send push notification via FCM (best effort)
	if s.fcm.IsAvailable() {
		user, err := s.authRepo.FindUserByID(userID)
		if err != nil {
			log.Printf("[NOTIF] User %d not found for push notification", userID)
			return nil // Inbox was saved, push is optional
		}

		data := map[string]string{
			"type":            notifType,
			"notification_id": strconv.FormatUint(uint64(notif.ID), 10),
		}

		for _, fcmToken := range user.FCMTokens {
			if err := s.fcm.SendToToken(ctx, fcmToken.Token, title, message, data); err != nil {
				log.Printf("[NOTIF] FCM send failed for user %d token %s: %v", userID, fcmToken.Token[:10]+"...", err)
			}
		}
	}

	return nil
}

// SendToMultipleUsers sends push + inbox notifications to multiple users.
func (s *service) SendToMultipleUsers(ctx context.Context, userIDs []uint, notifType, title, message string) error {
	for _, userID := range userIDs {
		if err := s.SendToUser(ctx, userID, notifType, title, message); err != nil {
			log.Printf("[NOTIF] Failed to send to user %d: %v", userID, err)
		}
	}
	return nil
}

func (s *service) GetInbox(userID uint, page, limit int) (*InboxResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	notifs, total, err := s.repo.GetByUserID(userID, page, limit)
	if err != nil {
		return nil, err
	}

	unread, _ := s.repo.GetUnreadCount(userID)

	return &InboxResult{
		Data:        notifs,
		Total:       total,
		UnreadCount: unread,
		Page:        page,
		Limit:       limit,
	}, nil
}

func (s *service) GetUnreadCount(userID uint) (int64, error) {
	return s.repo.GetUnreadCount(userID)
}

func (s *service) MarkAsRead(id, userID uint) error {
	return s.repo.MarkAsRead(id, userID)
}

func (s *service) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}
