package model

import "time"

type Status int

const (
	StatusIncomplete Status = iota
	StatusPartOneSolved
	StatusBothSolved
)

type Puzzle struct {
	ID int `gorm:"column:id;primaryKey;autoIncrement"`

	EventID int    `gorm:"column:event_id;uniqueIndex:idx_event_day"`
	Day     string `gorm:"column:day;uniqueIndex:idx_event_day"`

	Status Status `gorm:"column:status;"`
	// the user's puzzle input, cached since AoC asks not to download it repeatedly
	Input     string    `gorm:"column:input;"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;"`
}
