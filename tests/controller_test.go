package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lijuuu/EmployeeManagement/config"
	"github.com/lijuuu/EmployeeManagement/controller"
	"github.com/lijuuu/EmployeeManagement/database"
	"github.com/lijuuu/EmployeeManagement/repo"
	"github.com/lijuuu/EmployeeManagement/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//setupTestEnvironment sets up the test envt with database and redis connections
func setupTestEnvironment(t *testing.T) (*config.Config, *controller.EmployeeController, func()) {
	//load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("Skipping test: failed to load config: %v", err)
	}

	//set up database connection
	db, err := database.NewPostgresConn(context.Background(), cfg.PostgresDSN)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to database: %v", err)
	}

	//set up Redis client
	redisClient, err := database.InitRedis(cfg)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to Redis: %v", err)
	}

	//initialize dependencies
	repo := repo.NewEmployeeRepo(db)
	svc := service.NewEmployeeService(repo, redisClient)
	ctrl := controller.NewEmployeeController(svc, cfg)

	//return cleanup function
	cleanup := func() {
		db.Close(context.Background())
		redisClient.Close()
	}

	return cfg, ctrl, cleanup
}

//generateValidJWT creates a valid JWT token for testing
func generateValidJWT(cfg *config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": cfg.AdminEmail,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(cfg.JWTSecret))
}

func TestCreateEmployee(t *testing.T) {
	cfg, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()

	token, err := generateValidJWT(cfg)
	require.NoError(t, err)

	e := echo.New()

	reqBody := `{
		"name": "John Doe",
		"position": "Software Engineer",
		"salary": 60000,
		"hired_date": "2024-06-01T00:00:00Z"
	}`
	req := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = ctrl.CreateEmployee(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response controller.Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	empJSON, err := json.Marshal(response.Payload)
	require.NoError(t, err)

	var emp database.Employee
	err = json.Unmarshal(empJSON, &emp)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, emp.ID)
	assert.Equal(t, "John Doe", emp.Name)
	assert.Equal(t, "Software Engineer", emp.Position)
	assert.Equal(t, 60000.0, emp.Salary)
	assert.Equal(t, "2024-06-01", emp.HiredDate.Format("2006-01-02"))
	assert.False(t, emp.CreatedAt.IsZero())
	assert.False(t, emp.UpdatedAt.IsZero())
}

func TestGetEmployee(t *testing.T) {
	cfg, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()

	token, err := generateValidJWT(cfg)
	require.NoError(t, err)

	e := echo.New()

	createReqBody := `{
		"name": "Jane Smith",
		"position": "Product Manager",
		"salary": 75000
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewBufferString(createReqBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err = ctrl.CreateEmployee(createCtx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResponse controller.Response
	err = json.Unmarshal(createRec.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	empJSON, err := json.Marshal(createResponse.Payload)
	require.NoError(t, err)
	var createdEmp database.Employee
	err = json.Unmarshal(empJSON, &createdEmp)
	require.NoError(t, err)

	getReq := httptest.NewRequest(http.MethodGet, "/employees/"+createdEmp.ID.String(), nil)
	getRec := httptest.NewRecorder()
	getCtx := e.NewContext(getReq, getRec)
	getCtx.SetParamNames("id")
	getCtx.SetParamValues(createdEmp.ID.String())

	err = ctrl.GetEmployee(getCtx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getRec.Code)

	var getResponse controller.Response
	err = json.Unmarshal(getRec.Body.Bytes(), &getResponse)
	require.NoError(t, err)

	assert.Equal(t, "success", getResponse.Status)
	assert.Equal(t, http.StatusOK, getResponse.StatusCode)

	empJSON, err = json.Marshal(getResponse.Payload)
	require.NoError(t, err)
	var retrievedEmp database.Employee
	err = json.Unmarshal(empJSON, &retrievedEmp)
	require.NoError(t, err)

	assert.Equal(t, createdEmp.ID, retrievedEmp.ID)
	assert.Equal(t, "Jane Smith", retrievedEmp.Name)
	assert.Equal(t, "Product Manager", retrievedEmp.Position)
	assert.Equal(t, 75000.0, retrievedEmp.Salary)
}

func TestListEmployees(t *testing.T) {
	_, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/employees", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ctrl.ListEmployees(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response controller.Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	empListJSON, err := json.Marshal(response.Payload)
	require.NoError(t, err)
	var employees []database.Employee
	err = json.Unmarshal(empListJSON, &employees)
	require.NoError(t, err)

	assert.IsType(t, []database.Employee{}, employees)
}

func TestUpdateEmployee(t *testing.T) {
	cfg, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()

	token, err := generateValidJWT(cfg)
	require.NoError(t, err)

	e := echo.New()

	createReqBody := `{
		"name": "Bob Wilson",
		"position": "Developer",
		"salary": 65000
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewBufferString(createReqBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err = ctrl.CreateEmployee(createCtx)
	require.NoError(t, err)

	var createResponse controller.Response
	err = json.Unmarshal(createRec.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	empJSON, err := json.Marshal(createResponse.Payload)
	require.NoError(t, err)
	var createdEmp database.Employee
	err = json.Unmarshal(empJSON, &createdEmp)
	require.NoError(t, err)

	updateReqBody := `{
		"name": "Bob Wilson Jr",
		"position": "Senior Developer",
		"salary": 80000
	}`
	updateReq := httptest.NewRequest(http.MethodPut, "/employees/"+createdEmp.ID.String(), bytes.NewBufferString(updateReqBody))
	updateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRec := httptest.NewRecorder()
	updateCtx := e.NewContext(updateReq, updateRec)
	updateCtx.SetParamNames("id")
	updateCtx.SetParamValues(createdEmp.ID.String())

	err = ctrl.UpdateEmployee(updateCtx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, updateRec.Code)

	var updateResponse controller.Response
	err = json.Unmarshal(updateRec.Body.Bytes(), &updateResponse)
	require.NoError(t, err)

	assert.Equal(t, "success", updateResponse.Status)
	assert.Equal(t, http.StatusOK, updateResponse.StatusCode)

	empJSON, err = json.Marshal(updateResponse.Payload)
	require.NoError(t, err)
	var updatedEmp database.Employee
	err = json.Unmarshal(empJSON, &updatedEmp)
	require.NoError(t, err)

	assert.Equal(t, createdEmp.ID, updatedEmp.ID)
	assert.Equal(t, "Bob Wilson Jr", updatedEmp.Name)
	assert.Equal(t, "Senior Developer", updatedEmp.Position)
	assert.Equal(t, 80000.0, updatedEmp.Salary)
}

func TestDeleteEmployee(t *testing.T) {
	cfg, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()

	token, err := generateValidJWT(cfg)
	require.NoError(t, err)

	e := echo.New()

	createReqBody := `{
		"name": "Alice Brown",
		"position": "Designer",
		"salary": 55000
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewBufferString(createReqBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err = ctrl.CreateEmployee(createCtx)
	require.NoError(t, err)

	var createResponse controller.Response
	err = json.Unmarshal(createRec.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	empJSON, err := json.Marshal(createResponse.Payload)
	require.NoError(t, err)
	var createdEmp database.Employee
	err = json.Unmarshal(empJSON, &createdEmp)
	require.NoError(t, err)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/employees/"+createdEmp.ID.String(), nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRec := httptest.NewRecorder()
	deleteCtx := e.NewContext(deleteReq, deleteRec)
	deleteCtx.SetParamNames("id")
	deleteCtx.SetParamValues(createdEmp.ID.String())

	err = ctrl.DeleteEmployee(deleteCtx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, deleteRec.Code)

	getReq := httptest.NewRequest(http.MethodGet, "/employees/"+createdEmp.ID.String(), nil)
	getRec := httptest.NewRecorder()
	getCtx := e.NewContext(getReq, getRec)
	getCtx.SetParamNames("id")
	getCtx.SetParamValues(createdEmp.ID.String())

	err = ctrl.GetEmployee(getCtx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getRec.Code)
}

func TestLogin(t *testing.T) {
	cfg, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()

	e := echo.New()

	loginReqBody := `{
		"email": "` + cfg.AdminEmail + `",
		"password": "` + cfg.AdminPassword + `"
	}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginReqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ctrl.Login(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response controller.Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.NotEmpty(t, response.Payload)

	tokenString, ok := response.Payload.(string)
	assert.True(t, ok)
	assert.NotEmpty(t, tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)
}

func TestCreateEmployeeWithoutHiredDate(t *testing.T) {
	cfg, ctrl, cleanup := setupTestEnvironment(t)
	defer cleanup()

	token, err := generateValidJWT(cfg)
	require.NoError(t, err)

	e := echo.New()

	reqBody := `{
		"name": "Test Employee",
		"position": "Tester",
		"salary": 50000
	}`
	req := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = ctrl.CreateEmployee(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response controller.Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	empJSON, err := json.Marshal(response.Payload)
	require.NoError(t, err)

	var emp database.Employee
	err = json.Unmarshal(empJSON, &emp)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, emp.ID)
	assert.Equal(t, "Test Employee", emp.Name)
	assert.Equal(t, "Tester", emp.Position)
	assert.Equal(t, 50000.0, emp.Salary)
	assert.Equal(t, time.Now().Format("2006-01-02"), emp.HiredDate.Format("2006-01-02"))
	assert.False(t, emp.CreatedAt.IsZero())
	assert.False(t, emp.UpdatedAt.IsZero())
}
