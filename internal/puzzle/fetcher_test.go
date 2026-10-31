package puzzle

import (
	"net/http"
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

	fetcher := NewEventFetcher(http.DefaultClient)
	events, err := fetcher.Fetch(true)

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
