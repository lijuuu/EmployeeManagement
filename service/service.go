package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lijuuu/EmployeeManagement/database"
	"github.com/lijuuu/EmployeeManagement/repo"
	"github.com/redis/go-redis/v9"
)

type EmployeeService interface {
	CreateEmployee(ctx context.Context, emp *database.Employee) (uuid.UUID, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*database.Employee, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, emp *database.Employee) error
	DeleteEmployee(ctx context.Context, id uuid.UUID) error
	ListEmployees(ctx context.Context) ([]database.Employee, error)
}

type employeeService struct {
	repo  repo.EmployeeRepo
	redis *redis.Client
}

func NewEmployeeService(repo repo.EmployeeRepo, redis *redis.Client) EmployeeService {
	return &employeeService{
		repo:  repo,
		redis: redis,
	}
}

func (s *employeeService) CreateEmployee(ctx context.Context, emp *database.Employee) (uuid.UUID, error) {
	emp.CreatedAt = time.Now()
	emp.HiredDate = time.Now()

	id, err := s.repo.CreateEmployee(ctx, emp)
	if err != nil {
		return uuid.Nil, err
	}

	empJSON, err := json.Marshal(emp)
	if err != nil {
		return id, fmt.Errorf("failed to marshal employee: %v", err)
	}
	cacheKey := fmt.Sprintf("employee:%s", id.String())
	if err := s.redis.Set(ctx, cacheKey, empJSON, 5*time.Minute).Err(); err != nil {
		return id, fmt.Errorf("failed to cache employee: %v", err)
	}
	if err := s.redis.Del(ctx, "employees:list").Err(); err != nil {
		return id, fmt.Errorf("failed to invalidate list cache: %v", err)
	}
	return id, nil
}

func (s *employeeService) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*database.Employee, error) {
	cacheKey := fmt.Sprintf("employee:%s", id.String())
	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var emp database.Employee
		if err := json.Unmarshal([]byte(cached), &emp); err == nil {
			return &emp, nil
		}
	}

	//actual db
	emp, err := s.repo.GetEmployeeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	empJSON, err := json.Marshal(emp)
	if err != nil {
		return emp, fmt.Errorf("failed to marshal employee: %v", err)
	}
	if err := s.redis.Set(ctx, cacheKey, empJSON, 5*time.Minute).Err(); err != nil {
		return emp, fmt.Errorf("failed to cache employee: %v", err)
	}
	return emp, nil
}

func (s *employeeService) UpdateEmployee(ctx context.Context, id uuid.UUID, emp *database.Employee) error {
	err := s.repo.UpdateEmployee(ctx, id, emp)
	if err != nil {
		return err
	}

	empJSON, err := json.Marshal(emp)
	if err != nil {
		return fmt.Errorf("failed to marshal employee: %v", err)
	}
	cacheKey := fmt.Sprintf("employee:%s", id.String())
	if err := s.redis.Set(ctx, cacheKey, empJSON, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to cache employee: %v", err)
	}
	if err := s.redis.Del(ctx, "employees:list").Err(); err != nil {
		return fmt.Errorf("failed to invalidate list cache: %v", err)
	}
	return nil
}

func (s *employeeService) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	err := s.repo.DeleteEmployee(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("employee:%s", id.String())
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("failed to delete employee cache: %v", err)
	}
	if err := s.redis.Del(ctx, "employees:list").Err(); err != nil {
		return fmt.Errorf("failed to invalidate list cache: %v", err)
	}
	return nil
}

func (s *employeeService) ListEmployees(ctx context.Context) ([]database.Employee, error) {
	cacheKey := "employees:list"
	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var employees []database.Employee
		if err := json.Unmarshal([]byte(cached), &employees); err == nil {
			return employees, nil
		}
	}

	employees, err := s.repo.ListEmployees(ctx)
	if err != nil {
		return nil, err
	}

	empJSON, err := json.Marshal(employees)
	if err != nil {
		return employees, fmt.Errorf("failed to marshal employees: %v", err)
	}
	if err := s.redis.Set(ctx, cacheKey, empJSON, 5*time.Minute).Err(); err != nil {
		return employees, fmt.Errorf("failed to cache employees: %v", err)
	}
	return employees, nil
}
