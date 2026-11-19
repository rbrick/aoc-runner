package events

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/model"
	"github.com/rbrick/aoc-runner/internal/puzzle"
	"github.com/rbrick/aoc-runner/internal/session"
	"gorm.io/gorm"
)

type Handler struct {
	db      *gorm.DB
	fetcher *puzzle.EventFetcher
	session *session.Session
}

func (h *Handler) listEvents(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()
	all, err := gorm.G[model.Event](h.db).Find(ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var stars map[string]int

	if h.session.LoggedIn() {
		// stars are a nice to have, the years still list if AoC can't be reached
		stars, _ = h.fetcher.FetchEventStars(ctx)
	}

	return reqCtx.JSON(http.StatusOK, mapEvents(all, stars))
}

func (h *Handler) listEventDays(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()
	yearParam := reqCtx.Param("year")

	event, err := gorm.G[model.Event](h.db).Where("year = ?", yearParam).Preload("Puzzles", nil).First(ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var stars map[int]int

	if h.session.LoggedIn() {
		stars, _ = h.fetcher.FetchStars(ctx, yearParam)
	}

	return reqCtx.JSON(http.StatusOK, mapPuzzles(event.Puzzles, stars))
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

func (h *Handler) getInput(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()
	year := reqCtx.Param("year")

	day, err := strconv.Atoi(reqCtx.Param("day"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "day must be a number")
	}

	dbPuzzle, err := h.findPuzzle(ctx, year, day)

	if err != nil {
		return err
	}

	if dbPuzzle.Input != "" {
		return reqCtx.String(http.StatusOK, dbPuzzle.Input)
	}

	if !h.session.LoggedIn() {
		return echo.NewHTTPError(http.StatusUnauthorized, "log in to Advent of Code to get your puzzle input")
	}

	input, err := h.fetcher.FetchInput(ctx, year, day)

	var fetchErr *puzzle.FetchError

	// the session was accepted at login but AoC no longer takes it
	if errors.As(err, &fetchErr) && fetchErr.StatusCode == http.StatusBadRequest {
		return echo.NewHTTPError(http.StatusUnauthorized, "log in to Advent of Code to get your puzzle input")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	// AoC asks for inputs to be cached instead of downloaded again
	if _, err := gorm.G[model.Puzzle](h.db).Where("id = ?", dbPuzzle.ID).Update(ctx, "input", input); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return reqCtx.String(http.StatusOK, input)
}

// the stored puzzle for a year and day, as an echo error when it can't be found
func (h *Handler) findPuzzle(ctx context.Context, year string, day int) (*model.Puzzle, error) {
	event, err := gorm.G[model.Event](h.db).Where("year = ?", year).First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, echo.NewHTTPError(http.StatusNotFound, "event not found")
	}

	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	dbPuzzle, err := gorm.G[model.Puzzle](h.db).Where("event_id = ? AND day = ?", event.ID, fmt.Sprintf("Day %d", day)).First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, echo.NewHTTPError(http.StatusNotFound, "puzzle not found")
	}

	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return &dbPuzzle, nil
}

func (h *Handler) submitAnswer(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()
	year := reqCtx.Param("year")

	day, err := strconv.Atoi(reqCtx.Param("day"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "day must be a number")
	}

	var answer AnswerDTO

	if err := reqCtx.Bind(&answer); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	answer.Answer = strings.TrimSpace(answer.Answer)

	if answer.Answer == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "answer is required")
	}

	if answer.Level != 1 && answer.Level != 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "level must be 1 or 2")
	}

	if !h.session.LoggedIn() {
		return echo.NewHTTPError(http.StatusUnauthorized, "log in to Advent of Code to submit answers")
	}

	dbPuzzle, err := h.findPuzzle(ctx, year, day)

	if err != nil {
		return err
	}

	// a wrong answer is never right later, and resubmitting it only adds to AoC's timeout
	previous, err := gorm.G[model.Submission](h.db).
		Where("puzzle_id = ? AND level = ? AND answer = ? AND outcome = ?", dbPuzzle.ID, answer.Level, answer.Answer, string(puzzle.OutcomeIncorrect)).
		First(ctx)

	if err == nil {
		return reqCtx.JSON(http.StatusOK, &SubmitResultDTO{
			Outcome:   previous.Outcome,
			Message:   previous.Message,
			Duplicate: true,
		})
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	result, err := h.fetcher.SubmitAnswer(ctx, year, day, answer.Level, answer.Answer)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	submission := &model.Submission{
		PuzzleID: dbPuzzle.ID,
		Level:    answer.Level,
		Answer:   answer.Answer,
		Outcome:  string(result.Outcome),
		Message:  result.Message,
	}

	if err := gorm.G[model.Submission](h.db).Create(ctx, submission); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if result.Outcome == puzzle.OutcomeCorrect {
		status := model.StatusPartOneSolved

		if answer.Level == 2 {
			status = model.StatusBothSolved
		}

		if _, err := gorm.G[model.Puzzle](h.db).Where("id = ?", dbPuzzle.ID).Update(ctx, "status", status); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return reqCtx.JSON(http.StatusOK, &SubmitResultDTO{
		Outcome:     string(result.Outcome),
		Message:     result.Message,
		WaitSeconds: result.WaitSeconds,
	})
}

func NewHandler(db *gorm.DB, fetcher *puzzle.EventFetcher, session *session.Session) *Handler {
	return &Handler{
		db:      db,
		fetcher: fetcher,
		session: session,
	}
}
