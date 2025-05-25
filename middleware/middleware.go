package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/lijuuu/EmployeeManagement/config"
	customerr "github.com/lijuuu/EmployeeManagement/customerr"
)

// JWTAuthMiddleware validates JWT tokens for protected routes
func JWTAuthMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")
			if tokenString == "" {
				return customerr.NewError(c, http.StatusUnauthorized, "Missing Authorization header")
			}

			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			} else {
				return customerr.NewError(c, http.StatusUnauthorized, "Invalid Authorization header format")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, customerr.NewError(c, http.StatusUnauthorized, "Invalid signing method")
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				return customerr.NewError(c, http.StatusUnauthorized, "Invalid or expired token")
			}

			return next(c)
		}
	}
}

// RequestLoggerMiddleware logs incoming requests
func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			log.Printf("Request: %s %s from %s", c.Request().Method, c.Request().URL.Path, c.RealIP())
			err := next(c)
			log.Printf("Response: %s %s - Status: %d, Duration: %v",
				c.Request().Method, c.Request().URL.Path, c.Response().Status, time.Since(start))
			return err
		}
	}
}

// ErrorHandlerMiddleware customizes error responses
func ErrorHandlerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				switch e := err.(type) {
				case *echo.HTTPError:
					return customerr.NewError(c, e.Code, e.Message.(string))
				default:
					return customerr.NewError(c, http.StatusInternalServerError, "Internal server error")
				}
			}
			return nil
		}
	}
}
