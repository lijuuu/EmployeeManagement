package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lijuuu/EmployeeManagement/database"
	"github.com/redis/go-redis/v9"
)

type EmployeeRepo interface {
	CreateEmployee(ctx context.Context, emp *database.Employee) (int, error)
	GetEmployeeByID(ctx context.Context, id int) (*database.Employee, error)
	UpdateEmployee(ctx context.Context, id int, emp *database.Employee) error
	DeleteEmployee(ctx context.Context, id int) error
	ListEmployees(ctx context.Context) ([]database.Employee, error)
}

type employeeRepo struct {
	db    *pgx.Conn
	redis *redis.Client
}

func NewEmployeeRepo(db *pgx.Conn, redis *redis.Client) EmployeeRepo {
	return &employeeRepo{db: db, redis: redis}
}

func (r *employeeRepo) CreateEmployee(ctx context.Context, emp *database.Employee) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `
		INSERT INTO employees (name, position, salary, hired_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, emp.Name, emp.Position, emp.Salary, emp.HiredDate).Scan(&id)
	if err != nil {
		return 0, err
	}

	emp.ID = id
	empJSON, _ := json.Marshal(emp)
	r.redis.Set(ctx, fmt.Sprintf("employee:%d", id), empJSON, 1*time.Hour)
	return id, nil
}

func (r *employeeRepo) GetEmployeeByID(ctx context.Context, id int) (*database.Employee, error) {
	cacheKey := fmt.Sprintf("employee:%d", id)
	if cached, err := r.redis.Get(ctx, cacheKey).Result(); err == nil {
		var emp database.Employee
		if err := json.Unmarshal([]byte(cached), &emp); err == nil {
			return &emp, nil
		}
	}

	var emp database.Employee
	err := r.db.QueryRow(ctx, `
		SELECT id, name, position, salary, hired_date, created_at, updated_at
		FROM employees WHERE id = $1
	`, id).Scan(&emp.ID, &emp.Name, &emp.Position, &emp.Salary, &emp.HiredDate, &emp.CreatedAt, &emp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	empJSON, _ := json.Marshal(emp)
	r.redis.Set(ctx, cacheKey, empJSON, 1*time.Hour)
	return &emp, nil
}

func (r *employeeRepo) UpdateEmployee(ctx context.Context, id int, emp *database.Employee) error {
	_, err := r.db.Exec(ctx, `
		UPDATE employees
		SET name = $1, position = $2, salary = $3, hired_date = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`, emp.Name, emp.Position, emp.Salary, emp.HiredDate, id)
	if err != nil {
		return err
	}

	emp.ID = id
	empJSON, _ := json.Marshal(emp)
	r.redis.Set(ctx, fmt.Sprintf("employee:%d", id), empJSON, 1*time.Hour)
	return nil
}

func (r *employeeRepo) DeleteEmployee(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM employees WHERE id = $1`, id)
	if err != nil {
		return err
	}

	r.redis.Del(ctx, fmt.Sprintf("employee:%d", id))
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

	rows, err := r.db.Query(ctx, `
		SELECT id, name, position, salary, hired_date, created_at, updated_at
		FROM employees
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []database.Employee
	for rows.Next() {
		var emp database.Employee
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.Position, &emp.Salary, &emp.HiredDate, &emp.CreatedAt, &emp.UpdatedAt); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}

	empJSON, _ := json.Marshal(employees)
	r.redis.Set(ctx, cacheKey, empJSON, 1*time.Hour)
	return employees, nil
}
