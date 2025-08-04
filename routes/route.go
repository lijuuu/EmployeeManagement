package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/lijuuu/EmployeeManagement/config"
	"github.com/lijuuu/EmployeeManagement/controller"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo, ctrl *controller.EmployeeController, cfg *config.Config) {
	// Swagger route
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Login route (non-protected)
	e.POST("/login", ctrl.Login)

	// Grouped routes for JWT-protected endpoints
	protected := e.Group("/employees")
	// Uncomment and fix JWT middleware if needed
	// protected.Use(middleware.JWTAuthMiddleware(cfg))

	// Protected routes
	protected.POST("", ctrl.CreateEmployee)
	protected.PUT("/:id", ctrl.UpdateEmployee)
	protected.DELETE("/:id", ctrl.DeleteEmployee)

	// Non-protected read routes
	e.GET("/employees", ctrl.ListEmployees)
	e.GET("/employees/:id", ctrl.GetEmployee)
}
