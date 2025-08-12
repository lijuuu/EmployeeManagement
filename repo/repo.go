package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lijuuu/EmployeeManagement/database"
)

type EmployeeRepo interface {
	CreateEmployee(ctx context.Context, emp *database.Employee) (uuid.UUID, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*database.Employee, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, emp *database.Employee) error
	DeleteEmployee(ctx context.Context, id uuid.UUID) error
	ListEmployees(ctx context.Context) ([]database.Employee, error)
}

type employeeRepo struct {
	queries *Queries 
}

func NewEmployeeRepo(db *pgx.Conn) EmployeeRepo {
	return &employeeRepo{
		queries: New(db), 
	}
}

func (r *employeeRepo) CreateEmployee(ctx context.Context, emp *database.Employee) (uuid.UUID, error) {
		id := uuid.New() 
	_, err := r.queries.CreateEmployee(ctx, CreateEmployeeParams{
		ID:        id,
		Name:      emp.Name,
		Position:  emp.Position,
		Salary:    emp.Salary,
		HiredDate: pgtype.Date{Time: emp.HiredDate, Valid: true},
		CreatedAt: pgtype.Timestamp{Time:emp.CreatedAt,Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: emp.UpdatedAt,Valid: true},
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create employee: %v", err)
	}
	emp.ID = id
	return id, nil
}

func (r *employeeRepo) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*database.Employee, error) {
	dbEmp, err := r.queries.GetEmployeeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %v", err)
	}

	emp := &database.Employee{
		ID:        dbEmp.ID,
		Name:      dbEmp.Name,
		Position:  dbEmp.Position,
		Salary:    dbEmp.Salary,
		HiredDate: dbEmp.HiredDate.Time,
		CreatedAt: dbEmp.CreatedAt.Time,
		UpdatedAt: dbEmp.UpdatedAt.Time,
	}
	return emp, nil
}

func (r *employeeRepo) UpdateEmployee(ctx context.Context, id uuid.UUID, emp *database.Employee) error {
	err := r.queries.UpdateEmployee(ctx, UpdateEmployeeParams{
		Name:      emp.Name,
		Position:  emp.Position,
		Salary:    emp.Salary,
		HiredDate: pgtype.Date{Time: emp.HiredDate, Valid: true},
		ID:        id,
	})
	if err != nil {
		return fmt.Errorf("failed to update employee: %v", err)
	}
	emp.ID = id
	return nil
}

func (r *employeeRepo) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteEmployee(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %v", err)
	}
	return nil
}

func (r *employeeRepo) ListEmployees(ctx context.Context) ([]database.Employee, error) {
	dbEmployees, err := r.queries.ListEmployees(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %v", err)
	}

	employees := make([]database.Employee, len(dbEmployees))
	for i, dbEmp := range dbEmployees {
		var hiredDate time.Time
		if dbEmp.HiredDate.Valid {
			hiredDate = dbEmp.HiredDate.Time
		} else {
			return nil, fmt.Errorf("invalid hired_date for employee ID %s", dbEmp.ID.String())
		}

		employees[i] = database.Employee{
			ID:        dbEmp.ID,
			Name:      dbEmp.Name,
			Position:  dbEmp.Position,
			Salary:    dbEmp.Salary,
			HiredDate: hiredDate,
			CreatedAt: dbEmp.CreatedAt.Time,
			UpdatedAt: dbEmp.UpdatedAt.Time,
		}
	}
	return employees, nil
}
