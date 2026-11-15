package database

import (
	"context"
	"net/http"

	"github.com/rbrick/aoc-runner/internal/model"
	"github.com/rbrick/aoc-runner/internal/puzzle"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Seed(ctx context.Context, db *gorm.DB) error {
	fetcher := puzzle.NewEventFetcher(http.DefaultClient)

	events, err := fetcher.Fetch(ctx, true)

	if err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			dbEvent := &model.Event{
				Year: event.Year,
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "year"}},
				DoNothing: true,
			}).Where("year = ?", event.Year).
				FirstOrCreate(dbEvent).Error; err != nil {
				return err
			}

			var puzzles []*model.Puzzle

			for _, day := range event.Days {
				puzzle := mapToPuzzle(day)

				puzzle.EventID = dbEvent.ID

				puzzles = append(puzzles, puzzle)
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "event_id"}, {Name: "day"}},
				DoNothing: true,
			}).CreateInBatches(puzzles, len(puzzles)).Error; err != nil {
				return err
			}
		}

		return nil
	})

}

func mapToPuzzle(d puzzle.Day) *model.Puzzle {
	return &model.Puzzle{
		Day:    d.Day,
		Status: model.StatusIncomplete,
	}
}
