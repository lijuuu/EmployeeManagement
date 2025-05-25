package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/lijuuu/EmployeeManagement/config"
	"github.com/lijuuu/EmployeeManagement/controller"
	swagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo, ctrl *controller.EmployeeController, cfg *config.Config) {
	e.GET("/swagger/*", swagger.WrapHandler)

	e.POST("/login", ctrl.Login)

	// Protected routes with JWT
	protected := e.Group("/employees")
	// protected.Use(middleware.JWTWithConfig(middleware.JWTConfig{
	// 	SigningKey: []byte(cfg.JWTSecret),
	// }))

	protected.POST("", ctrl.CreateEmployee)
	protected.PUT("/:id", ctrl.UpdateEmployee)
	protected.DELETE("/:id", ctrl.DeleteEmployee)

	// Public routes
	e.GET("/employees", ctrl.ListEmployees)
	e.GET("/employees/:id", ctrl.GetEmployee)
}
