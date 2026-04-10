package common

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c echo.Context, status int, code, message string) error {
	return c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

func ValidationError(c echo.Context, errors map[string]string) error {
	return c.JSON(http.StatusUnprocessableEntity, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Details: errors,
		},
	})
}
