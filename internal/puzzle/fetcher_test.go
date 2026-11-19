package puzzle

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"

	_ "embed"
)

const (
	EmptyBody = `<html></html>`
)

var (
	//go:embed testdata/events.html
	SuccessfulEventFetch string

	//go:embed testdata/days.html
	SuccessfulDayFetch string

	//go:embed testdata/problem.html
	SuccessfulProblemFetch string
)

func htmlResponder(str string) httpmock.Responder {
	return func(r *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, str)
		// set header type to text/html
		response.Header.Set("Content-Type", "text/html")

		return response, nil
	}
}

func setupHappyPath() {
	httpmock.RegisterResponder(http.MethodGet, EventsURL, htmlResponder(SuccessfulEventFetch))

	httpmock.RegisterResponder(
		http.MethodGet,
		`=~^https://adventofcode\.com/?$`,
		htmlResponder(EmptyBody),
	)

	httpmock.RegisterResponder(http.MethodGet, `=~^https://adventofcode.com/(\d+)/?$`, func(r *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, SuccessfulDayFetch)
		// set header type to text/html
		response.Header.Set("Content-Type", "text/html")
		return response, nil
	})
}

func TestFetchSuccess(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	setupHappyPath()

	bgCtx := context.Background()

	fetcher := NewEventFetcher(http.DefaultClient)
	events, err := fetcher.Fetch(bgCtx, true)

	if err != nil {
		t.Error(err)
		return
	}

	if len(events) != 12 {
		t.Errorf("unexpected event length: expected 3, got: %d", len(events))
		return
	}

	if events[0].Year != "2026" {
		t.Errorf("unexpected year, expected: 2026, got: %s", events[0].Year)
		return
	}

	days := events[1].Days

	if len(days) != 12 {
		t.Errorf("unexpected days, expected: 12, got: %d", len(days))
		return
	}

}

func TestYearExtractor(t *testing.T) {
	extracted := YearExtractor("[2026]")

	if extracted != "2026" {
		t.Errorf("unexpected extracted value, expected: 2026, got: %s", extracted)
		return
	}

}

func TestFetchProblem(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(
		http.MethodGet,
		`=~^https://adventofcode\.com/(\d+)/day/(\d+)/?$`,
		htmlResponder(SuccessfulProblemFetch),
	)

	bgCtx := context.Background()

	fetcher := NewEventFetcher(http.DefaultClient)
	problem, err := fetcher.FetchProblem(bgCtx, &Day{
		Day: "Day 1",
		URL: BaseURL + "/2025/day/1",
	})

	if err != nil {
		t.Error(err)
		return
	}

	if !strings.HasPrefix(problem.Body, `<article class="day-desc"><h2>--- Day 1: Test Puzzle ---</h2>`) {
		t.Errorf("unexpected problem body start, got: %s", problem.Body)
		return
	}

	if !strings.Contains(problem.Body, `<h2 id="part2">--- Part Two ---</h2>`) {
		t.Errorf("expected part two in problem body, got: %s", problem.Body)
		return
	}

	if strings.Contains(problem.Body, "Your puzzle answer was") {
		t.Errorf("unexpected content outside of the puzzle articles, got: %s", problem.Body)
		return
	}

	if len(problem.Examples) != 3 {
		t.Errorf("unexpected examples length, expected: 3, got: %d", len(problem.Examples))
		return
	}

	if problem.Examples[1] != "(()(\n<)(>\n" {
		t.Errorf("unexpected example, expected: %q, got: %q", "(()(\n<)(>\n", problem.Examples[1])
		return
	}

	if len(problem.ExampleAnswers) != 2 {
		t.Errorf("unexpected example answers length, expected: 2, got: %d", len(problem.ExampleAnswers))
		return
	}

	if problem.ExampleAnswers[0] != "5" {
		t.Errorf("unexpected part one example answer, expected: 5, got: %s", problem.ExampleAnswers[0])
		return
	}

	if problem.ExampleAnswers[1] != "" {
		t.Errorf("unexpected part two example answer, expected empty, got: %s", problem.ExampleAnswers[1])
		return
	}
}
