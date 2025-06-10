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

	e.Logger.Fatal(e.Start(":8080"))
}
