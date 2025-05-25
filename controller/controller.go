package controller

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lijuuu/EmployeeManagement/config"
	"github.com/lijuuu/EmployeeManagement/customerr"
	"github.com/lijuuu/EmployeeManagement/database"
	"github.com/lijuuu/EmployeeManagement/service"
	"golang.org/x/crypto/bcrypt"
)

// EmployeeController handles HTTP requests for employee operations
type EmployeeController struct {
	service service.EmployeeService
	cfg     *config.Config
}

func NewEmployeeController(service service.EmployeeService, cfg *config.Config) *EmployeeController {
	return &EmployeeController{service: service, cfg: cfg}
}

// Login godoc
// @Summary Admin login
// @Description Authenticate admin and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "Admin credentials"
// @Success 200 {object} map[string]string
// @Failure 401 {object} customerr.ErrorResponse
// @Router /login [post]
func (c *EmployeeController) Login(ctx echo.Context) error {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&credentials); err != nil {
		return customerr.NewError(ctx, http.StatusBadRequest, "Invalid request body")
	}

	if credentials.Email != c.cfg.AdminEmail || bcrypt.CompareHashAndPassword([]byte(c.cfg.AdminPassword), []byte(credentials.Password)) != nil {
		return customerr.NewError(ctx, http.StatusUnauthorized, "Invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": credentials.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(c.cfg.JWTSecret))
	if err != nil {
		return customerr.NewError(ctx, http.StatusInternalServerError, "Failed to generate token")
	}

	return ctx.JSON(http.StatusOK, map[string]string{"token": tokenString})
}

// CreateEmployee godoc
// @Summary Create a new employee
// @Description Create a new employee record
// @Tags employees
// @Accept json
// @Produce json
// @Param employee body database.Employee true "Employee data"
// @Success 201 {object} database.Employee
// @Failure 400 {object} customerr.ErrorResponse
// @Failure 401 {object} customerr.ErrorResponse
// @Router /employees [post]
func (c *EmployeeController) CreateEmployee(ctx echo.Context) error {
	var emp database.Employee
	if err := ctx.Bind(&emp); err != nil {
		return customerr.NewError(ctx, http.StatusBadRequest, "Invalid request body")
	}

	id, err := c.service.CreateEmployee(ctx.Request().Context(), &emp)
	if err != nil {
		return customerr.NewError(ctx, http.StatusInternalServerError, err.Error())
	}

	emp.ID = id
	return ctx.JSON(http.StatusCreated, emp)
}

// GetEmployee godoc
// @Summary Get employee by ID
// @Description Retrieve details of a specific employee
// @Tags employees
// @Accept json
// @Produce json
// @Param id path string true "Employee ID" format(uuid)
// @Success 200 {object} database.Employee
// @Failure 400 {object} customerr.ErrorResponse
// @Failure 404 {object} customerr.ErrorResponse
// @Router /employees/{id} [get]
func (c *EmployeeController) GetEmployee(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return customerr.NewError(ctx, http.StatusBadRequest, "Invalid employee ID")
	}

	emp, err := c.service.GetEmployeeByID(ctx.Request().Context(), id)
	if err != nil {
		return customerr.NewError(ctx, http.StatusNotFound, "Employee not found")
	}

	return ctx.JSON(http.StatusOK, emp)
}

// UpdateEmployee godoc
// @Summary Update an employee
// @Description Update details of a specific employee
// @Tags employees
// @Accept json
// @Produce json
// @Param id path string true "Employee ID" format(uuid)
// @Param employee body database.Employee true "Employee data"
// @Success 200 {object} database.Employee
// @Failure 400 {object} customerr.ErrorResponse
// @Failure 401 {object} customerr.ErrorResponse
// @Failure 404 {object} customerr.ErrorResponse
// @Router /employees/{id} [put]
func (c *EmployeeController) UpdateEmployee(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return customerr.NewError(ctx, http.StatusBadRequest, "Invalid employee ID")
	}

	var emp database.Employee
	if err := ctx.Bind(&emp); err != nil {
		return customerr.NewError(ctx, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.service.UpdateEmployee(ctx.Request().Context(), id, &emp); err != nil {
		return customerr.NewError(ctx, http.StatusNotFound, "Employee not found")
	}

	emp.ID = id
	return ctx.JSON(http.StatusOK, emp)
}

// DeleteEmployee godoc
// @Summary Delete an employee
// @Description Delete a specific employee
// @Tags employees
// @Accept json
// @Produce json
// @Param id path string true "Employee ID" format(uuid)
// @Success 204
// @Failure 400 {object} customerr.ErrorResponse
// @Failure 401 {object} customerr.ErrorResponse
// @Failure 404 {object} customerr.ErrorResponse
// @Router /employees/{id} [delete]
func (c *EmployeeController) DeleteEmployee(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return customerr.NewError(ctx, http.StatusBadRequest, "Invalid employee ID")
	}

	if err := c.service.DeleteEmployee(ctx.Request().Context(), id); err != nil {
		return customerr.NewError(ctx, http.StatusNotFound, "Employee not found")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ListEmployees godoc
// @Summary List all employees
// @Description Retrieve a list of all employees
// @Tags employees
// @Accept json
// @Produce json
// @Success 200 {array} database.Employee
// @Failure 500 {object} customerr.ErrorResponse
// @Router /employees [get]
func (c *EmployeeController) ListEmployees(ctx echo.Context) error {
	employees, err := c.service.ListEmployees(ctx.Request().Context())
	if err != nil {
		return customerr.NewError(ctx, http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, employees)
}
