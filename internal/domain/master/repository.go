package master

import "gorm.io/gorm"

type Repository interface {
	// JobSite
	GetAllJobSites() ([]JobSite, error)
	CreateJobSite(j *JobSite) error
	UpdateJobSite(j *JobSite) error
	DeleteJobSite(id uint) error

	// Department
	GetAllDepartments() ([]Department, error)
	CreateDepartment(d *Department) error
	UpdateDepartment(d *Department) error
	DeleteDepartment(id uint) error

	// Position
	GetAllPositions() ([]Position, error)
	GetPositionsByDepartment(deptID uint) ([]Position, error)
	CreatePosition(p *Position) error
	UpdatePosition(p *Position) error
	DeletePosition(id uint) error

	// ContractType
	GetAllContractTypes() ([]ContractType, error)
	CreateContractType(c *ContractType) error
	UpdateContractType(c *ContractType) error
	DeleteContractType(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// --- JobSite ---
func (r *repository) GetAllJobSites() ([]JobSite, error) {
	var sites []JobSite
	err := r.db.Find(&sites).Error
	return sites, err
}
func (r *repository) CreateJobSite(j *JobSite) error {
	return r.db.Create(j).Error
}
func (r *repository) UpdateJobSite(j *JobSite) error {
	return r.db.Save(j).Error
}
func (r *repository) DeleteJobSite(id uint) error {
	return r.db.Delete(&JobSite{}, id).Error
}

// --- Department ---
func (r *repository) GetAllDepartments() ([]Department, error) {
	var depts []Department
	err := r.db.Find(&depts).Error
	return depts, err
}
func (r *repository) CreateDepartment(d *Department) error {
	return r.db.Create(d).Error
}
func (r *repository) UpdateDepartment(d *Department) error {
	return r.db.Save(d).Error
}
func (r *repository) DeleteDepartment(id uint) error {
	return r.db.Delete(&Department{}, id).Error
}

// --- Position ---
func (r *repository) GetAllPositions() ([]Position, error) {
	var pos []Position
	err := r.db.Preload("Department").Find(&pos).Error
	return pos, err
}
func (r *repository) GetPositionsByDepartment(deptID uint) ([]Position, error) {
	var pos []Position
	err := r.db.Where("department_id = ?", deptID).Find(&pos).Error
	return pos, err
}
func (r *repository) CreatePosition(p *Position) error {
	return r.db.Create(p).Error
}
func (r *repository) UpdatePosition(p *Position) error {
	return r.db.Save(p).Error
}
func (r *repository) DeletePosition(id uint) error {
	return r.db.Delete(&Position{}, id).Error
}

// --- ContractType ---
func (r *repository) GetAllContractTypes() ([]ContractType, error) {
	var types []ContractType
	err := r.db.Find(&types).Error
	return types, err
}
func (r *repository) CreateContractType(c *ContractType) error {
	return r.db.Create(c).Error
}
func (r *repository) UpdateContractType(c *ContractType) error {
	return r.db.Save(c).Error
}
func (r *repository) DeleteContractType(id uint) error {
	return r.db.Delete(&ContractType{}, id).Error
}
