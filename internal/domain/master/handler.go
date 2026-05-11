package master

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// --- JobSites ---
func (h *Handler) GetJobSites(c *gin.Context) {
	sites, err := h.service.GetJobSites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data Job Sites",
		"data":    sites,
	})
}

func (h *Handler) CreateJobSite(c *gin.Context) {
	var input struct {
		Name         string  `json:"name" binding:"required"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		RadiusMeters int     `json:"radius_meters"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	site, err := h.service.CreateJobSite(input.Name, input.Latitude, input.Longitude, input.RadiusMeters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Job Site created successfully",
		"data":    site,
	})
}

func (h *Handler) UpdateJobSite(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var input struct {
		Name         string  `json:"name" binding:"required"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		RadiusMeters int     `json:"radius_meters"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	site, err := h.service.UpdateJobSite(uint(id), input.Name, input.Latitude, input.Longitude, input.RadiusMeters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Job Site updated successfully",
		"data":    site,
	})
}

func (h *Handler) DeleteJobSite(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := h.service.DeleteJobSite(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Job Site deleted successfully"})
}

// --- Departments ---
func (h *Handler) GetDepartments(c *gin.Context) {
	depts, err := h.service.GetDepartments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data Departments",
		"data":    depts,
	})
}

func (h *Handler) CreateDepartment(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dept, err := h.service.CreateDepartment(input.Name, input.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Department created successfully",
		"data":    dept,
	})
}

func (h *Handler) UpdateDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var input struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dept, err := h.service.UpdateDepartment(uint(id), input.Name, input.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Department updated successfully",
		"data":    dept,
	})
}

func (h *Handler) DeleteDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := h.service.DeleteDepartment(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Department deleted successfully"})
}

// --- Positions ---
func (h *Handler) GetPositions(c *gin.Context) {
	deptIDStr := c.Query("department_id")
	if deptIDStr != "" {
		deptID, _ := strconv.Atoi(deptIDStr)
		pos, err := h.service.GetPositionsByDept(uint(deptID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Data Positions by Department",
			"data":    pos,
		})
		return
	}

	pos, err := h.service.GetPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data Positions",
		"data":    pos,
	})
}

func (h *Handler) CreatePosition(c *gin.Context) {
	var input struct {
		Name         string `json:"name" binding:"required"`
		DepartmentID *uint  `json:"department_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pos, err := h.service.CreatePosition(input.Name, input.DepartmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Position created successfully",
		"data":    pos,
	})
}

func (h *Handler) UpdatePosition(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var input struct {
		Name         string `json:"name" binding:"required"`
		DepartmentID *uint  `json:"department_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pos, err := h.service.UpdatePosition(uint(id), input.Name, input.DepartmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Position updated successfully",
		"data":    pos,
	})
}

func (h *Handler) DeletePosition(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := h.service.DeletePosition(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Position deleted successfully"})
}

// --- ContractTypes ---
func (h *Handler) GetContractTypes(c *gin.Context) {
	types, err := h.service.GetContractTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data Contract Types",
		"data":    types,
	})
}

func (h *Handler) CreateContractType(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ct, err := h.service.CreateContractType(input.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Contract Type created successfully",
		"data":    ct,
	})
}

func (h *Handler) UpdateContractType(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ct, err := h.service.UpdateContractType(uint(id), input.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Contract Type updated successfully",
		"data":    ct,
	})
}

func (h *Handler) DeleteContractType(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := h.service.DeleteContractType(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Contract Type deleted successfully"})
}
