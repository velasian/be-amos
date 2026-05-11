package system

import (
	"gorm.io/gorm"
)

// Repository defines the interface for file metadata persistence.
type Repository interface {
	CreateFile(file *File) error
	FindFileByID(id uint) (*File, error)
	FindFilesByEntity(entityType string, entityID uint) ([]File, error)
	FindFilesByEntityAndCategory(entityType string, entityID uint, category string) ([]File, error)
	DeleteFile(id uint) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new system file repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateFile(file *File) error {
	return r.db.Create(file).Error
}

func (r *repository) FindFileByID(id uint) (*File, error) {
	var f File
	err := r.db.First(&f, id).Error
	return &f, err
}

func (r *repository) FindFilesByEntity(entityType string, entityID uint) ([]File, error) {
	var files []File
	err := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

func (r *repository) FindFilesByEntityAndCategory(entityType string, entityID uint, category string) ([]File, error) {
	var files []File
	err := r.db.Where("entity_type = ? AND entity_id = ? AND category = ?", entityType, entityID, category).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

func (r *repository) DeleteFile(id uint) error {
	return r.db.Delete(&File{}, id).Error
}
