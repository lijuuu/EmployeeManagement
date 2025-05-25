package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	// Load configuration
	fmt.Println(os.Getwd())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("Skipping TestCreateEmployee: failed to load config: %v", err)
	}

	// Set up database connection
	db, err := database.NewPostgresConn(context.Background(), cfg.PostgresDSN)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(context.Background())

	// Set up Redis client
	redisClient, err := database.InitRedis(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize Echo and dependencies
	e := echo.New()
	repo := repo.NewEmployeeRepo(db)
	svc := service.NewEmployeeService(repo, redisClient)
	ctrl := controller.NewEmployeeController(svc, cfg)

	// Prepare request
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

	// Execute the handler
	err = ctrl.CreateEmployee(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify response
	var emp database.Employee
	err = json.Unmarshal(rec.Body.Bytes(), &emp)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, emp.ID)
	assert.Equal(t, "John Doe", emp.Name)
	assert.Equal(t, "Software Engineer", emp.Position)
	assert.Equal(t, 60000.0, emp.Salary)
	assert.Equal(t, "2024-06-01", emp.HiredDate.Format("2006-01-02"))

	// Cleanup
	_, err = db.Exec(context.Background(), "DELETE FROM employees WHERE id = $1", emp.ID)
	assert.NoError(t, err)
}
