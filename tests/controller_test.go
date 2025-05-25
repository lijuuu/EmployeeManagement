package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/lijuuu/EmployeeManagement/config"
	"github.com/lijuuu/EmployeeManagement/controller"
	"github.com/lijuuu/EmployeeManagement/database"
	"github.com/lijuuu/EmployeeManagement/repo"
	"github.com/lijuuu/EmployeeManagement/service"
	"github.com/stretchr/testify/assert"
)

func TestCreateEmployee(t *testing.T) {
	cfg := &config.Config{
		PostgresDSN:   "postgres://user:password@localhost:5432/testdb",
		RedisAddr:     "localhost:6379",
		AdminEmail:    "admin@example.com",
		AdminPassword: "$2a$10$...hashedpassword...",
		JWTSecret:     "secret",
	}

	db, err := database.InitDB(cfg)
	assert.NoError(t, err)
	defer db.Close(context.Background())

	redisClient, err := database.InitRedis(cfg)
	assert.NoError(t, err)
	defer redisClient.Close()

	e := echo.New()
	repo := repo.NewEmployeeRepo(db, redisClient)
	svc := service.NewEmployeeService(repo)
	ctrl := controller.NewEmployeeController(svc, cfg)

	reqBody := `{"name":"John Doe","position":"Software Engineer","salary":60000,"hired_date":"2024-06-01"}`
	req := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = ctrl.CreateEmployee(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var emp database.Employee
	err = json.Unmarshal(rec.Body.Bytes(), &emp)
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", emp.Name)
}
