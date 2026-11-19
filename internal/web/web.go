package web

import (
	"context"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/web/events"
	"github.com/rbrick/aoc-runner/internal/web/submit"
	"gorm.io/gorm"
)

func Start(ctx context.Context, db *gorm.DB, host string) error {
	errChan := make(chan error, 1)
	go func(errChan chan error) {
		router := echo.New()

		// register the event handler
		events.NewHandler(db).RegisterRoutes(router.Group("/events"))
		// register the submission handler
		submit.NewHandler().RegisterRoutes(router.Group("/submit"))

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
