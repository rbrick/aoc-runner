package puzzle

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

const (
	SuccessfulEventFetch = `
<!DOCTYPE html>
<html lang="en-us">
<body>
<main>
<article>
<div class="eventlist-event"><a href="/">[2026]</a>          </div>
<div class="eventlist-event"><a href="/2025">[2025]</a>          </div>
<div class="eventlist-event"><a href="/2024">[2024]</a>          </div>
<div class="eventlist-event"><a href="/2023">[2023]</a>          </div>
<div class="eventlist-event"><a href="/2022">[2022]</a>          </div>
<div class="eventlist-event"><a href="/2021">[2021]</a>          </div>
<div class="eventlist-event"><a href="/2020">[2020]</a>          </div>
<div class="eventlist-event"><a href="/2019">[2019]</a>          </div>
<div class="eventlist-event"><a href="/2018">[2018]</a>          </div>
<div class="eventlist-event"><a href="/2017">[2017]</a>          </div>
<div class="eventlist-event"><a href="/2016">[2016]</a>          </div>
<div class="eventlist-event"><a href="/2015">[2015]</a>          </div>
</article>
</main>
</body>
</html>
	`
)

func setupHappyPath() {
	httpmock.RegisterResponder(http.MethodGet, EventsURL, func(r *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, SuccessfulEventFetch)
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
	events, err := fetcher.Fetch()

	if err != nil {
		t.Error(err)
		return
	}

	if len(events) != 12 {
		t.Errorf("unexpected event length: expected 3, got: %d", len(events))
		return
	}

	if events[0].Year != "[2026]" {
		t.Errorf("unexpected year, expected: [2026], got: %s", events[0].Year)
		return
	}
}
