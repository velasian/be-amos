package master

type Service interface {
	// JobSite
	GetJobSites() ([]JobSite, error)
	CreateJobSite(name string, lat, lng float64, radius int) (*JobSite, error)
	UpdateJobSite(id uint, name string, lat, lng float64, radius int) (*JobSite, error)
	DeleteJobSite(id uint) error

	// Department
	GetDepartments() ([]Department, error)
	CreateDepartment(name, code string) (*Department, error)
	UpdateDepartment(id uint, name, code string) (*Department, error)
	DeleteDepartment(id uint) error

	// Position
	GetPositions() ([]Position, error)
	GetPositionsByDept(deptID uint) ([]Position, error)
	CreatePosition(name string, deptID *uint) (*Position, error)
	UpdatePosition(id uint, name string, deptID *uint) (*Position, error)
	DeletePosition(id uint) error

	// ContractType
	GetContractTypes() ([]ContractType, error)
	CreateContractType(name string) (*ContractType, error)
	UpdateContractType(id uint, name string) (*ContractType, error)
	DeleteContractType(id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// --- JobSite ---
func (s *service) GetJobSites() ([]JobSite, error) {
	return s.repo.GetAllJobSites()
}
func (s *service) CreateJobSite(name string, lat, lng float64, radius int) (*JobSite, error) {
	site := &JobSite{
		Name:         name,
		Latitude:     lat,
		Longitude:    lng,
		RadiusMeters: radius,
	}
	if err := s.repo.CreateJobSite(site); err != nil {
		return nil, err
	}
	return site, nil
}
func (s *service) UpdateJobSite(id uint, name string, lat, lng float64, radius int) (*JobSite, error) {
	site := &JobSite{
		ID:           id,
		Name:         name,
		Latitude:     lat,
		Longitude:    lng,
		RadiusMeters: radius,
	}
	if err := s.repo.UpdateJobSite(site); err != nil {
		return nil, err
	}
	return site, nil
}
func (s *service) DeleteJobSite(id uint) error {
	return s.repo.DeleteJobSite(id)
}

// --- Department ---
func (s *service) GetDepartments() ([]Department, error) {
	return s.repo.GetAllDepartments()
}
func (s *service) CreateDepartment(name, code string) (*Department, error) {
	dept := &Department{Name: name, Code: code}
	if err := s.repo.CreateDepartment(dept); err != nil {
		return nil, err
	}
	return dept, nil
}
func (s *service) UpdateDepartment(id uint, name, code string) (*Department, error) {
	dept := &Department{ID: id, Name: name, Code: code}
	if err := s.repo.UpdateDepartment(dept); err != nil {
		return nil, err
	}
	return dept, nil
}
func (s *service) DeleteDepartment(id uint) error {
	return s.repo.DeleteDepartment(id)
}

// --- Position ---
func (s *service) GetPositions() ([]Position, error) {
	return s.repo.GetAllPositions()
}
func (s *service) GetPositionsByDept(deptID uint) ([]Position, error) {
	return s.repo.GetPositionsByDepartment(deptID)
}
func (s *service) CreatePosition(name string, deptID *uint) (*Position, error) {
	pos := &Position{
		Name:         name,
		DepartmentID: deptID,
	}
	if err := s.repo.CreatePosition(pos); err != nil {
		return nil, err
	}
	return pos, nil
}
func (s *service) UpdatePosition(id uint, name string, deptID *uint) (*Position, error) {
	pos := &Position{ID: id, Name: name, DepartmentID: deptID}
	if err := s.repo.UpdatePosition(pos); err != nil {
		return nil, err
	}
	return pos, nil
}
func (s *service) DeletePosition(id uint) error {
	return s.repo.DeletePosition(id)
}

// --- ContractType ---
func (s *service) GetContractTypes() ([]ContractType, error) {
	return s.repo.GetAllContractTypes()
}
func (s *service) CreateContractType(name string) (*ContractType, error) {
	ct := &ContractType{Name: name}
	if err := s.repo.CreateContractType(ct); err != nil {
		return nil, err
	}
	return ct, nil
}
func (s *service) UpdateContractType(id uint, name string) (*ContractType, error) {
	ct := &ContractType{ID: id, Name: name}
	if err := s.repo.UpdateContractType(ct); err != nil {
		return nil, err
	}
	return ct, nil
}
func (s *service) DeleteContractType(id uint) error {
	return s.repo.DeleteContractType(id)
}
