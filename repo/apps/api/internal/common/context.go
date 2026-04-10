package common

import (
	"github.com/labstack/echo/v4"
)

const (
	ContextKeyUserID    = "user_id"
	ContextKeyEmail     = "email"
	ContextKeyRoles     = "roles"
	ContextKeyRequestID = "request_id"
)

func GetUserID(c echo.Context) string {
	v, _ := c.Get(ContextKeyUserID).(string)
	return v
}

func GetEmail(c echo.Context) string {
	v, _ := c.Get(ContextKeyEmail).(string)
	return v
}

func GetRoles(c echo.Context) []string {
	v, _ := c.Get(ContextKeyRoles).([]string)
	return v
}

func GetRequestID(c echo.Context) string {
	v, _ := c.Get(ContextKeyRequestID).(string)
	return v
}
