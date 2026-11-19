package auth

import "github.com/labstack/echo/v5"

func (h *Handler) RegisterRoutes(group *echo.Group) {
	group.GET("/", h.status)
	group.POST("/login", h.login)
}
