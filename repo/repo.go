package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lijuuu/EmployeeManagement/database"
	"github.com/redis/go-redis/v9"
)

type EmployeeRepo interface {
	CreateEmployee(ctx context.Context, emp *database.Employee) (uuid.UUID, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*database.Employee, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, emp *database.Employee) error
	DeleteEmployee(ctx context.Context, id uuid.UUID) error
	ListEmployees(ctx context.Context) ([]database.Employee, error)
}

type employeeRepo struct {
	queries *Queries // sqlc generated Queries struct in repo package
	redis   *redis.Client
}

func NewEmployeeRepo(db *pgx.Conn, redis *redis.Client) EmployeeRepo {
	return &employeeRepo{
		queries: New(db), // Initialize sqlc Queries from repo package
		redis:   redis,
	}
}

func (r *employeeRepo) CreateEmployee(ctx context.Context, emp *database.Employee) (uuid.UUID, error) {
	id := uuid.New() // Generate UUID in Go
	_, err := r.queries.CreateEmployee(ctx, CreateEmployeeParams{
		ID:        id,
		Name:      emp.Name,
		Position:  emp.Position,
		Salary:    emp.Salary,
		HiredDate: pgtype.Date{Time: emp.HiredDate, Valid: true},
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create employee: %v", err)
	}

	emp.ID = id
	empJSON, err := json.Marshal(emp)
	if err != nil {
		return id, fmt.Errorf("failed to marshal employee: %v", err)
	}
	if err := r.redis.Set(ctx, fmt.Sprintf("employee:%s", id.String()), empJSON, 1*time.Hour).Err(); err != nil {
		return id, fmt.Errorf("failed to cache employee: %v", err)
	}
	if err := r.redis.Del(ctx, "employees:list").Err(); err != nil {
		return id, fmt.Errorf("failed to invalidate list cache: %v", err)
	}
	return id, nil
}

func (r *employeeRepo) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*database.Employee, error) {
	cacheKey := fmt.Sprintf("employee:%s", id.String())
	if cached, err := r.redis.Get(ctx, cacheKey).Result(); err == nil {
		var emp database.Employee
		if err := json.Unmarshal([]byte(cached), &emp); err == nil {
			return &emp, nil
		}
	}

	dbEmp, err := r.queries.GetEmployeeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %v", err)
	}

	// Convert pgtype.Date to time.Time
	var hiredDate time.Time
	if dbEmp.HiredDate.Valid {
		hiredDate = dbEmp.HiredDate.Time
	} else {
		return nil, fmt.Errorf("invalid hired_date for employee ID %s", id.String())
	}

	emp := &database.Employee{
		ID:        dbEmp.ID,
		Name:      dbEmp.Name,
		Position:  dbEmp.Position,
		Salary:    dbEmp.Salary,
		HiredDate: hiredDate,
		CreatedAt: dbEmp.CreatedAt.Time,
		UpdatedAt: dbEmp.UpdatedAt.Time,
	}

	empJSON, err := json.Marshal(emp)
	if err != nil {
		return emp, fmt.Errorf("failed to marshal employee: %v", err)
	}
	if err := r.redis.Set(ctx, cacheKey, empJSON, 1*time.Hour).Err(); err != nil {
		return emp, fmt.Errorf("failed to cache employee: %v", err)
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
	empJSON, err := json.Marshal(emp)
	if err != nil {
		return fmt.Errorf("failed to marshal employee: %v", err)
	}
	if err := r.redis.Set(ctx, fmt.Sprintf("employee:%s", id.String()), empJSON, 1*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to cache employee: %v", err)
	}
	if err := r.redis.Del(ctx, "employees:list").Err(); err != nil {
		return fmt.Errorf("failed to invalidate list cache: %v", err)
	}
	return nil
}

func (r *employeeRepo) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteEmployee(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %v", err)
	}

	if err := r.redis.Del(ctx, fmt.Sprintf("employee:%s", id.String())).Err(); err != nil {
		return fmt.Errorf("failed to delete employee cache: %v", err)
	}
	if err := r.redis.Del(ctx, "employees:list").Err(); err != nil {
		return fmt.Errorf("failed to invalidate list cache: %v", err)
	}
	return nil
}

func (r *employeeRepo) ListEmployees(ctx context.Context) ([]database.Employee, error) {
	cacheKey := "employees:list"
	if cached, err := r.redis.Get(ctx, cacheKey).Result(); err == nil {
		var employees []database.Employee
		if err := json.Unmarshal([]byte(cached), &employees); err == nil {
			return employees, nil
		}
	}

	dbEmployees, err := r.queries.ListEmployees(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %v", err)
	}

	employees := make([]database.Employee, len(dbEmployees))
	for i, dbEmp := range dbEmployees {
		// Convert pgtype.Date to time.Time
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

	empJSON, err := json.Marshal(employees)
	if err != nil {
		return employees, fmt.Errorf("failed to marshal employees: %v", err)
	}
	if err := r.redis.Set(ctx, cacheKey, empJSON, 1*time.Hour).Err(); err != nil {
		return employees, fmt.Errorf("failed to cache employees: %v", err)
	}
	return employees, nil
}
