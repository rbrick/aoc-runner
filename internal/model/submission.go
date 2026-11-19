package model

import "time"

// an answer sent to Advent of Code, kept so a wrong answer is never sent twice
type Submission struct {
	ID int `gorm:"column:id;primaryKey;autoIncrement"`

	PuzzleID int    `gorm:"column:puzzle_id;index:idx_puzzle_level_answer"`
	Level    int    `gorm:"column:level;index:idx_puzzle_level_answer"`
	Answer   string `gorm:"column:answer;index:idx_puzzle_level_answer"`

	Outcome string `gorm:"column:outcome;"`
	Message string `gorm:"column:message;"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;"`
}
