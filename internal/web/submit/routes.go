package submit

import "github.com/labstack/echo/v5"

func (h *Handler) RegisterRoutes(group *echo.Group) {
	group.GET("/languages", h.listLanguages)
	group.POST("/", h.submit)
}
