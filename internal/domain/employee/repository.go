package employee

import (
	"gorm.io/gorm"
)

type Repository interface {
	CreateEmployee(e *Employee) error
	FindByID(id uint) (*Employee, error)
	FindByUserID(userID uint) (*Employee, error)
	FindByNRP(nrp string) (*Employee, error)
	FindByNfcUID(uid string) (*Employee, error)
	UpdateEmployee(e *Employee) error
	DeleteEmployee(id uint) error

	SaveDetail(d *EmployeeDetail) error
	SaveContractHistory(c *ContractHistory) error

	GetAllEmployeesPaginated(page, limit int, search string) ([]Employee, int64, error)
	GetAllWithDetails() ([]Employee, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateEmployee(e *Employee) error {
	return r.db.Create(e).Error
}

func (r *repository) FindByID(id uint) (*Employee, error) {
	var e Employee
	err := r.db.Preload("Detail").First(&e, id).Error
	return &e, err
}

func (r *repository) FindByUserID(userID uint) (*Employee, error) {
	var e Employee
	err := r.db.Preload("Detail").Where("user_id = ?", userID).First(&e).Error
	return &e, err
}

func (r *repository) FindByNRP(nrp string) (*Employee, error) {
	var e Employee
	err := r.db.Preload("Detail").Where("nrp = ?", nrp).First(&e).Error
	return &e, err
}

func (r *repository) FindByNfcUID(uid string) (*Employee, error) {
	var e Employee
	err := r.db.Where("nfc_uid = ?", uid).First(&e).Error
	return &e, err
}

func (r *repository) UpdateEmployee(e *Employee) error {
	return r.db.Save(e).Error
}

func (r *repository) DeleteEmployee(id uint) error {
	return r.db.Delete(&Employee{}, id).Error
}

func (r *repository) SaveDetail(d *EmployeeDetail) error {
	return r.db.Save(d).Error
}

func (r *repository) SaveContractHistory(c *ContractHistory) error {
	return r.db.Save(c).Error
}

func (r *repository) GetAllEmployeesPaginated(page, limit int, search string) ([]Employee, int64, error) {
	var e []Employee
	var total int64

	query := r.db.Model(&Employee{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR nrp ILIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	err := query.
		Preload("Detail").
		Offset(offset).
		Limit(limit).
		Order("name ASC").
		Find(&e).Error

	return e, total, err
}

func (r *repository) GetAllWithDetails() ([]Employee, error) {
	var e []Employee
	err := r.db.
		Preload("Detail").
		Preload("Position").
		Preload("Department").
		Preload("JobSite").
		Preload("ContractHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("id DESC").Preload("ContractType")
		}).
		Order("name ASC").
		Find(&e).Error
	return e, err
}
