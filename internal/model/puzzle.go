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

	EventID string `gorm:"column:event_id;"`
	Day     string `gorm:"column:day;"`

	Status    Status    `gorm:"column:status;"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;"`
}
