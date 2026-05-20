package importdata

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// Handler handles HTTP requests for employee import operations.
type Handler struct {
	service Service
}

// NewHandler creates a new import handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// DownloadTemplate generates and serves an Excel template for employee import.
// GET /import/template
func (h *Handler) DownloadTemplate(c *gin.Context) {
	xlsx := excelize.NewFile()
	sheetName := "Employee Data"
	xlsx.SetSheetName("Sheet1", sheetName)

	headers := []string{
		"NRP", "Nama Lengkap", "Email", "Password",
		"Jabatan", "Job Site",
		"Tempat Lahir", "Tanggal Lahir (YYYY-MM-DD)", "Jenis Kelamin (L/P)", "Agama", "Status Pernikahan", "Tanggal Pernikahan (YYYY-MM-DD)", "Golongan Darah", "Tinggi Badan (cm)", "Berat Badan (kg)",
		"NIK", "Alamat KTP", "Alamat Domisili",
		"No HP", "No HP Keluarga Darurat", "Nama Keluarga Darurat", "Hubungan Keluarga", "Nama Ibu Kandung",
		"No BPJS Kesehatan", "No BPJS Ketenagakerjaan", "No NPWP", "Nama Bank", "No Rekening", "Nama Pemilik Rekening",
		"Ukuran Baju", "Ukuran Sepatu", "Ukuran Celana",
		"Tipe Kontrak", "No SK", "Tanggal Mulai Kontrak (YYYY-MM-DD)", "Tanggal Selesai Kontrak (YYYY-MM-DD)",
	}

	// Header style
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E3A5F"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	// Set generic column width
	for i := 0; i < len(headers); i++ {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsx.SetColWidth(sheetName, colName, colName, 20)
	}

	xlsx.SetRowHeight(sheetName, 1, 30)

	// Write headers
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsx.SetCellValue(sheetName, cell, header)
		xlsx.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Example row
	exampleData := []string{
		"AM021001", "Ahmad Fauzi", "ahmad@company.com", "", "M",
		"Staff HRD", "Head Office Jakarta",
		"Jakarta", "1990-05-15", "Islam", "O", "Married", "Jl. Sudirman No. 123", "081234567890", "12.345.678.9-012.000",
		"PKWT I", "SK/HRD/001/2024", "2024-01-15", "2025-01-14",
	}

	dataStyle, _ := xlsx.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#666666", Italic: true},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#CCCCCC", Style: 1},
			{Type: "right", Color: "#CCCCCC", Style: 1},
			{Type: "top", Color: "#CCCCCC", Style: 1},
			{Type: "bottom", Color: "#CCCCCC", Style: 1},
		},
	})

	for i, data := range exampleData {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		xlsx.SetCellValue(sheetName, cell, data)
		xlsx.SetCellStyle(sheetName, cell, cell, dataStyle)
	}

	// Freeze header row
	xlsx.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=template_import_employee.xlsx")
	xlsx.Write(c.Writer)
}

// ParseExcel uploads an Excel file, parses it into staging, and triggers validation.
// POST /import/parse
func (h *Handler) ParseExcel(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Excel file is required",
		})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to open file",
		})
		return
	}
	defer f.Close()

	batchID, count, err := h.service.ProcessExcelToStaging(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Auto-validate immediately
	_ = h.service.ValidateStaging(batchID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data imported to staging and validated",
		"data": gin.H{
			"batch_id": batchID,
			"count":    count,
		},
	})
}

// GetStagingData returns all staging rows for a given batch.
// GET /import/staging/:batchId
func (h *Handler) GetStagingData(c *gin.Context) {
	batchID := c.Param("batchId")
	items, err := h.service.GetStagingData(batchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Batch not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Staging data retrieved",
		"data":    items,
	})
}

// UpdateStagingField allows partial editing of a staging row before commit.
// PATCH /import/staging/:id
func (h *Handler) UpdateStagingField(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Staging ID is required",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.UpdateStagingFields(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Staging record updated",
	})
}

// SubmitImport commits all valid staging data into the real tables.
// POST /import/commit
func (h *Handler) SubmitImport(c *gin.Context) {
	var req struct {
		BatchID string `json:"batch_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "batch_id is required",
		})
		return
	}

	// Re-validate before commit
	if err := h.service.ValidateStaging(req.BatchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Validation failed: " + err.Error(),
		})
		return
	}

	success, failed, err := h.service.CommitStaging(req.BatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Commit failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Import processed",
		"data": gin.H{
			"success": success,
			"failed":  failed,
		},
	})
}
