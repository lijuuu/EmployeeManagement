package service

import (
	"context"

	"github.com/lijuuu/EmployeeManagement/database"
	"github.com/lijuuu/EmployeeManagement/repo"
)

type EmployeeService interface {
	CreateEmployee(ctx context.Context, emp *database.Employee) (int, error)
	GetEmployeeByID(ctx context.Context, id int) (*database.Employee, error)
	UpdateEmployee(ctx context.Context, id int, emp *database.Employee) error
	DeleteEmployee(ctx context.Context, id int) error
	ListEmployees(ctx context.Context) ([]database.Employee, error)
}

type employeeService struct {
	repo repo.EmployeeRepo
}

func NewEmployeeService(repo repo.EmployeeRepo) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) CreateEmployee(ctx context.Context, emp *database.Employee) (int, error) {
	return s.repo.CreateEmployee(ctx, emp)
}

func (s *employeeService) GetEmployeeByID(ctx context.Context, id int) (*database.Employee, error) {
	return s.repo.GetEmployeeByID(ctx, id)
}

func (s *employeeService) UpdateEmployee(ctx context.Context, id int, emp *database.Employee) error {
	return s.repo.UpdateEmployee(ctx, id, emp)
}

func (s *employeeService) DeleteEmployee(ctx context.Context, id int) error {
	return s.repo.DeleteEmployee(ctx, id)
}

func (s *employeeService) ListEmployees(ctx context.Context) ([]database.Employee, error) {
	return s.repo.ListEmployees(ctx)
}
