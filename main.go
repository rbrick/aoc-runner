package main

import (
	"context"
	"log"

	"github.com/rbrick/aoc-runner/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	db, err := gorm.Open(sqlite.Open("aoc.db"), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	database.Migrate(context.Background(), db)

	database.Seed(context.Background(), db)

}
