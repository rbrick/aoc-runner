package events

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/model"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func (h *Handler) listEvents(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()
	all, err := gorm.G[model.Event](h.db).Find(ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return reqCtx.JSON(http.StatusOK, mapEvents(all))
}

func (h *Handler) listEventDays(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()
	yearParam := reqCtx.Param("year")

	event, err := gorm.G[model.Event](h.db).Where("year = ?", yearParam).Preload("Puzzles", nil).First(ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return reqCtx.JSON(http.StatusOK, mapPuzzles(event.Puzzles))
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}
