package events

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/model"
	"github.com/rbrick/aoc-runner/internal/puzzle"
	"gorm.io/gorm"
)

type Handler struct {
	db      *gorm.DB
	fetcher *puzzle.EventFetcher
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

func (h *Handler) getProblem(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()

	year, err := strconv.Atoi(reqCtx.Param("year"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "year must be a number")
	}

	day, err := strconv.Atoi(reqCtx.Param("day"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "day must be a number")
	}

	problem, err := h.fetcher.FetchProblem(ctx, &puzzle.Day{
		Day: fmt.Sprintf("Day %d", day),
		URL: fmt.Sprintf("%s/%d/day/%d", puzzle.BaseURL, year, day),
	})

	var fetchErr *puzzle.FetchError

	if errors.As(err, &fetchErr) && fetchErr.StatusCode == http.StatusNotFound {
		return echo.NewHTTPError(http.StatusNotFound, "puzzle not found")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	return reqCtx.JSON(http.StatusOK, mapProblem(problem))
}

func NewHandler(db *gorm.DB, fetcher *puzzle.EventFetcher) *Handler {
	return &Handler{
		db:      db,
		fetcher: fetcher,
	}
}
