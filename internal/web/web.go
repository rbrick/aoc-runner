package web

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/puzzle"
	"github.com/rbrick/aoc-runner/internal/session"
	"github.com/rbrick/aoc-runner/internal/web/auth"
	"github.com/rbrick/aoc-runner/internal/web/events"
	"github.com/rbrick/aoc-runner/internal/web/submit"
	"github.com/rbrick/aoc-runner/ui"
	"gorm.io/gorm"
)

func Start(ctx context.Context, db *gorm.DB, session *session.Session, host string) error {
	errChan := make(chan error, 1)
	go func(errChan chan error) {
		router := echo.New()

		// register the event handler, puzzles are fetched as the logged in user
		events.NewHandler(db, puzzle.NewEventFetcher(session.Client()), session).RegisterRoutes(router.Group("/events"))
		// register the auth handler
		auth.NewHandler(session).RegisterRoutes(router.Group("/auth"))
		// register the submission handler
		submit.NewHandler().RegisterRoutes(router.Group("/submit"))

		// serve the web interface
		if ui.Built() {
			router.StaticFS("/", ui.Dist())
		} else {
			router.GET("/", func(reqCtx *echo.Context) error {
				return reqCtx.String(http.StatusOK, "the web interface hasn't been built, run: go generate ./ui")
			})
		}

		if err := router.Start(host); err != nil {
			errChan <- err
		}
	}(errChan)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
