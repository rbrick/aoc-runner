package puzzle

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"

	_ "embed"
)

var (
	//go:embed testdata/calendar_logged_in.html
	LoggedInCalendar string

	//go:embed testdata/events_logged_in.html
	LoggedInEvents string
)

func TestFetchStars(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(http.MethodGet, BaseURL+"/2018", htmlResponder(LoggedInCalendar))

	stars, err := NewEventFetcher(http.DefaultClient).FetchStars(context.Background(), "2018")

	if err != nil {
		t.Error(err)
		return
	}

	expected := map[int]int{1: 2, 2: 1, 3: 0}

	for day, count := range expected {
		if stars[day] != count {
			t.Errorf("unexpected stars for day %d, expected: %d, got: %d", day, count, stars[day])
			return
		}
	}

	if len(stars) != len(expected) {
		t.Errorf("unexpected days, expected: %d, got: %d", len(expected), len(stars))
		return
	}
}

func TestFetchEventStars(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(http.MethodGet, EventsURL, htmlResponder(LoggedInEvents))

	stars, err := NewEventFetcher(http.DefaultClient).FetchEventStars(context.Background())

	if err != nil {
		t.Error(err)
		return
	}

	if stars["2018"] != 2 {
		t.Errorf("unexpected stars for 2018, expected: 2, got: %d", stars["2018"])
		return
	}

	if stars["2017"] != 50 {
		t.Errorf("unexpected stars for 2017, expected: 50, got: %d", stars["2017"])
		return
	}

	if _, found := stars["2019"]; found {
		t.Errorf("unexpected stars for 2019, expected none, got: %d", stars["2019"])
		return
	}
}

func TestFetchInput(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(http.MethodGet, BaseURL+"/2024/day/1/input",
		httpmock.NewStringResponder(http.StatusOK, "3   4\n4   3\n"))

	input, err := NewEventFetcher(http.DefaultClient).FetchInput(context.Background(), "2024", 1)

	if err != nil {
		t.Error(err)
		return
	}

	if input != "3   4\n4   3\n" {
		t.Errorf("unexpected input, expected: %q, got: %q", "3   4\n4   3\n", input)
		return
	}
}

func TestFetchInputNotLoggedIn(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(http.MethodGet, BaseURL+"/2024/day/1/input",
		httpmock.NewStringResponder(http.StatusBadRequest, "Puzzle inputs differ by user.  Please log in to get your puzzle input.\n"))

	_, err := NewEventFetcher(http.DefaultClient).FetchInput(context.Background(), "2024", 1)

	var fetchErr *FetchError

	if !errors.As(err, &fetchErr) || fetchErr.StatusCode != http.StatusBadRequest {
		t.Errorf("unexpected error, expected: a 400 FetchError, got: %v", err)
		return
	}
}
