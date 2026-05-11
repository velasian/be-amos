package notification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for notification operations.
type Handler struct {
	service Service
}

// NewHandler creates a new notification handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetInbox returns paginated notifications for the authenticated user.
// GET /notifications?page=1&limit=20
func (h *Handler) GetInbox(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.service.GetInbox(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Notifications retrieved",
		"data":    result,
	})
}

// GetUnreadCount returns the number of unread notifications.
// GET /notifications/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := c.GetUint("userID")

	count, err := h.service.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Unread count",
		"data": gin.H{
			"unread_count": count,
		},
	})
}

// MarkAsRead marks a single notification as read.
// PATCH /notifications/:id/read
func (h *Handler) MarkAsRead(c *gin.Context) {
	userID := c.GetUint("userID")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid notification ID",
		})
		return
	}

	if err := h.service.MarkAsRead(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead marks all notifications as read for the authenticated user.
// PATCH /notifications/read-all
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetUint("userID")

	if err := h.service.MarkAllAsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "All notifications marked as read",
	})
}

// SendTestNotification sends a test push notification (admin only, for debugging).
// POST /notifications/test
func (h *Handler) SendTestNotification(c *gin.Context) {
	var input struct {
		UserID  uint   `json:"user_id" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SendToUser(c.Request.Context(), input.UserID, "test", input.Title, input.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Test notification sent",
	})
}
