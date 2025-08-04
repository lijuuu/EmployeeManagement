package main

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/lijuuu/EmployeeManagement/config"
	"github.com/lijuuu/EmployeeManagement/controller"
	"github.com/lijuuu/EmployeeManagement/database"
	_ "github.com/lijuuu/EmployeeManagement/docs"
	"github.com/lijuuu/EmployeeManagement/middleware"
	"github.com/lijuuu/EmployeeManagement/repo"
	"github.com/lijuuu/EmployeeManagement/routes"
	"github.com/lijuuu/EmployeeManagement/service"
)

// @title Employee Management API
// @version 1.0
// @description This is a sample server for managing employees.

// @contact.name Liju Thomas
// @contact.email liju@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host employeemanagement-69ga.onrender.com
// @BasePath /
// @schemes https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization


func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	db, err := database.NewPostgresConn(context.Background(), cfg.PostgresDSN)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close(context.Background())

	redisClient, err := database.InitRedis(cfg)
	if err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
		return
	}
	defer redisClient.Close()

	e := echo.New()
	e.Use(middleware.RequestLoggerMiddleware())

	employeeRepo := repo.NewEmployeeRepo(db)
	employeeService := service.NewEmployeeService(employeeRepo, redisClient)
	employeeController := controller.NewEmployeeController(employeeService, cfg)

	routes.SetupRoutes(e, employeeController, cfg)

	e.Start(":8080")
}
