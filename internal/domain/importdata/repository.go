package importdata

import (
	"gorm.io/gorm"
)

// Repository defines the interface for staging table operations.
type Repository interface {
	CreateBatch(items []EmployeeStaging) error
	GetByBatchID(batchID string) ([]EmployeeStaging, error)
	GetValidByBatchID(batchID string) ([]EmployeeStaging, error)
	UpdateStatus(id uint, status, errorMsg string) error
	UpdateFields(id string, updates map[string]interface{}) error
	DeleteStaging(id uint) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new import staging repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateBatch(items []EmployeeStaging) error {
	return r.db.CreateInBatches(items, 100).Error
}

func (r *repository) GetByBatchID(batchID string) ([]EmployeeStaging, error) {
	var list []EmployeeStaging
	err := r.db.Where("import_batch_id = ?", batchID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *repository) GetValidByBatchID(batchID string) ([]EmployeeStaging, error) {
	var list []EmployeeStaging
	err := r.db.Where("import_batch_id = ? AND status IN ?", batchID, []string{"VALID", "UPDATE"}).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *repository) UpdateStatus(id uint, status, errorMsg string) error {
	return r.db.Model(&EmployeeStaging{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        status,
			"error_message": errorMsg,
		}).Error
}

func (r *repository) UpdateFields(id string, updates map[string]interface{}) error {
	return r.db.Model(&EmployeeStaging{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) DeleteStaging(id uint) error {
	return r.db.Delete(&EmployeeStaging{}, id).Error
}
