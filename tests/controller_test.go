package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lijuuu/EmployeeManagement/config"
	"github.com/lijuuu/EmployeeManagement/controller"
	"github.com/lijuuu/EmployeeManagement/database"
	"github.com/lijuuu/EmployeeManagement/repo"
	"github.com/lijuuu/EmployeeManagement/service"
	"github.com/stretchr/testify/assert"
)

func TestCreateEmployee(t *testing.T) {
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	db, err := database.NewPostgresConn(context.Background(), cfg.PostgresDSN)
	assert.NoError(t, err)
	defer db.Close(context.Background())

	redisClient, err := database.InitRedis(cfg)
	assert.NoError(t, err)
	defer redisClient.Close()

	e := echo.New()
	repo := repo.NewEmployeeRepo(db, redisClient)
	svc := service.NewEmployeeService(repo)
	ctrl := controller.NewEmployeeController(svc, cfg)

	reqBody := `{
		"name": "John Doe",
		"position": "Software Engineer",
		"salary": 60000,
		"hired_date": "2024-06-01T00:00:00Z"
	}`
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
	assert.NotEqual(t, uuid.Nil, emp.ID)
	assert.Equal(t, "John Doe", emp.Name)
	assert.Equal(t, "Software Engineer", emp.Position)
	assert.Equal(t, 60000.0, emp.Salary)
	assert.Equal(t, "2024-06-01", emp.HiredDate.Format("2006-01-02"))

	_, err = db.Exec(context.Background(), "DELETE FROM employees WHERE id = $1", emp.ID)
	assert.NoError(t, err)
}
