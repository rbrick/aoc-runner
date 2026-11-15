package database

import (
	"context"

	"github.com/rbrick/aoc-runner/internal/model"
	"gorm.io/gorm"
)

func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(model.Event{}, model.Puzzle{})

}
