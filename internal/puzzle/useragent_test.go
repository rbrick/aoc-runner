package puzzle

import (
	"context"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestRequestsSendUserAgent(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	agents := map[string]string{}

	record := func(name, page string) httpmock.Responder {
		return func(r *http.Request) (*http.Response, error) {
			agents[name] = r.Header.Get("User-Agent")
			return htmlResponder(page)(r)
		}
	}

	httpmock.RegisterResponder(http.MethodGet, BaseURL+"/2025/day/1", record("problem", SuccessfulProblemFetch))
	httpmock.RegisterResponder(http.MethodGet, BaseURL+"/2025/day/1/input", record("input", "1\n"))
	httpmock.RegisterResponder(http.MethodPost, BaseURL+"/2025/day/1/answer", record("answer", WrongLevelAnswer))

	bgCtx := context.Background()
	fetcher := NewEventFetcher(http.DefaultClient)

	if _, err := fetcher.FetchProblem(bgCtx, &Day{URL: BaseURL + "/2025/day/1"}); err != nil {
		t.Error(err)
		return
	}

	if _, err := fetcher.FetchInput(bgCtx, "2025", 1); err != nil {
		t.Error(err)
		return
	}

	if _, err := fetcher.SubmitAnswer(bgCtx, "2025", 1, 1, "1"); err != nil {
		t.Error(err)
		return
	}

	for _, name := range []string{"problem", "input", "answer"} {
		if agents[name] != UserAgent {
			t.Errorf("unexpected user agent for %s, expected: %s, got: %q", name, UserAgent, agents[name])
			return
		}
	}
}
