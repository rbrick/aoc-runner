package main

import (
	"log"
	"net/http"

	"github.com/rbrick/aoc-runner/internal/puzzle"
)

func main() {
	fetcher := puzzle.NewEventFetcher(http.DefaultClient)

	events, err := fetcher.Fetch(true)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(events)
}
