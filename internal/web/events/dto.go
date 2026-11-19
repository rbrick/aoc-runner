package events

import (
	"slices"
	"strconv"
	"strings"

	"github.com/rbrick/aoc-runner/internal/model"
	"github.com/rbrick/aoc-runner/internal/puzzle"
)

type EventDTO struct {
	Year string `json:"year"`
	// only set when logged in
	Stars *int `json:"stars,omitempty"`
}

type EventsDTO struct {
	Events []EventDTO `json:"events"`
}

type EventDayDTO struct {
	Day    string `json:"day"`
	Number int    `json:"number"`
	// only set when logged in
	Stars *int `json:"stars,omitempty"`
}

type EventDaysDTO struct {
	Days []EventDayDTO `json:"days"`
}

func mapEvents(events []model.Event, stars map[string]int) *EventsDTO {
	var dtos []EventDTO

	for _, event := range events {
		dtos = append(dtos, EventDTO{
			Year:  event.Year,
			Stars: starsFor(stars, event.Year),
		})
	}

	return &EventsDTO{
		Events: dtos,
	}
}

func mapPuzzles(puzzles []model.Puzzle, stars map[int]int) *EventDaysDTO {
	var dtos []EventDayDTO

	for _, puzzle := range puzzles {
		number := dayNumber(puzzle.Day)

		dtos = append(dtos, EventDayDTO{
			Day:    puzzle.Day,
			Number: number,
			Stars:  starsFor(stars, number),
		})
	}

	slices.SortFunc(dtos, func(a, b EventDayDTO) int {
		return a.Number - b.Number
	})

	return &EventDaysDTO{
		Days: dtos,
	}
}

// nil when stars weren't fetched, so logged out users don't see zero stars everywhere
func starsFor[K comparable](stars map[K]int, key K) *int {
	if stars == nil {
		return nil
	}

	count := stars[key]

	return &count
}

// days are stored by name, like "Day 1"
func dayNumber(day string) int {
	number, _ := strconv.Atoi(strings.TrimPrefix(day, "Day "))
	return number
}

type ProblemDTO struct {
	Body           string   `json:"body"`
	Examples       []string `json:"examples"`
	ExampleAnswers []string `json:"example_answers"`
}

func mapProblem(problem *puzzle.Problem) *ProblemDTO {
	return &ProblemDTO{
		Body:           problem.Body,
		Examples:       problem.Examples,
		ExampleAnswers: problem.ExampleAnswers,
	}
}

type AnswerDTO struct {
	// 1 for part one, 2 for part two
	Level  int    `json:"level"`
	Answer string `json:"answer"`
}

type SubmitResultDTO struct {
	Outcome     string `json:"outcome"`
	Message     string `json:"message"`
	WaitSeconds int    `json:"wait_seconds,omitempty"`
	// true when the answer was already rejected before, so it wasn't sent again
	Duplicate bool `json:"duplicate,omitempty"`
}
