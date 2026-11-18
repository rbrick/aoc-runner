package web

import (
	"context"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/web/events"
	"gorm.io/gorm"
)

func Start(ctx context.Context, db *gorm.DB, host string) error {
	errChan := make(chan error, 1)
	go func(errChan chan error) {
		router := echo.New()

		// register the event handler
		events.NewHandler(db).RegisterRoutes(router.Group("/events"))

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
