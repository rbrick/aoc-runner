package events

import "github.com/labstack/echo/v5"

func (h *Handler) RegisterRoutes(group *echo.Group) {
	group.GET("/", h.listEvents)
	group.GET("/:year", h.listEventDays)
}
