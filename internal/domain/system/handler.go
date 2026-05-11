package system

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for file operations.
type Handler struct {
	service Service
}

// NewHandler creates a new file handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// UploadFile handles file upload for any entity.
// POST /files/upload
// Form fields: entity_type, entity_id, category
// Form file:   file
func (h *Handler) UploadFile(c *gin.Context) {
	entityType := c.PostForm("entity_type")
	entityIDStr := c.PostForm("entity_id")
	category := c.PostForm("category")

	if entityType == "" || entityIDStr == "" || category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "entity_type, entity_id, and category are required",
		})
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid entity_id format",
		})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "File is required",
		})
		return
	}

	fileRecord, err := h.service.UploadFile(c.Request.Context(), entityType, uint(entityID), category, fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "File uploaded successfully",
		"data":    fileRecord,
	})
}

// GetFilesByEntity returns all files for a given entity.
// GET /files?entity_type=employee&entity_id=1&category=profile_photo
func (h *Handler) GetFilesByEntity(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")
	category := c.Query("category")

	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "entity_type and entity_id query params are required",
		})
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid entity_id format",
		})
		return
	}

	var files []File
	if category != "" {
		files, err = h.service.GetFilesByEntityAndCategory(entityType, uint(entityID), category)
	} else {
		files, err = h.service.GetFilesByEntity(entityType, uint(entityID))
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Files retrieved",
		"data":    files,
	})
}

// GetFileDownloadURL generates a temporary presigned download link.
// GET /files/:id/download
func (h *Handler) GetFileDownloadURL(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid file ID",
		})
		return
	}

	url, err := h.service.GetFileDownloadURL(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Download link generated",
		"data": gin.H{
			"download_url": url,
		},
	})
}

// DeleteFile removes a file from storage and database.
// DELETE /files/:id
func (h *Handler) DeleteFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid file ID",
		})
		return
	}

	if err := h.service.DeleteFile(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File deleted successfully",
	})
}
