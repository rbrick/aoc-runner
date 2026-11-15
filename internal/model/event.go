package model

type Event struct {
	ID      int      `gorm:"column:id;autoIncrement;primaryKey"`
	Year    string   `gorm:"column:year;unique;"`
	Puzzles []Puzzle `gorm:"foreignKey:event_id"`
}
