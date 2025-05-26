package customerr

import (
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request body"`
}

func NewError(ctx echo.Context, status int, message string) error {
	return ctx.JSON(status, ErrorResponse{Error: message})
}
