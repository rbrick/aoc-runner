package events

import "github.com/rbrick/aoc-runner/internal/model"

type EventDTO struct {
	Year string `json:"year"`
}

type EventsDTO struct {
	Events []EventDTO `json:"events"`
}

type EventDayDTO struct {
	Day string `json:"day"`
}

type EventDaysDTO struct {
	Days []EventDayDTO `json:"days"`
}

func mapEvents(events []model.Event) *EventsDTO {
	var dtos []EventDTO

	for _, event := range events {
		dtos = append(dtos, EventDTO{
			Year: event.Year,
		})
	}

	return &EventsDTO{
		Events: dtos,
	}
}

func mapPuzzles(puzzles []model.Puzzle) *EventDaysDTO {
	var dtos []EventDayDTO

	for _, puzzle := range puzzles {
		dtos = append(dtos, EventDayDTO{
			Day: puzzle.Day,
		})
	}

	return &EventDaysDTO{
		Days: dtos,
	}

}
