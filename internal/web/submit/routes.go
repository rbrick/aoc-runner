package submit

import "github.com/labstack/echo/v5"

func (h *Handler) RegisterRoutes(group *echo.Group) {
	group.POST("/", h.submit)
}
