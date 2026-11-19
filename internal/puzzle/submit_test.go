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
	AnswerURL = BaseURL + "/2024/day/1/answer"
)

var (
	//go:embed testdata/answers/correct.html
	CorrectAnswer string

	//go:embed testdata/answers/incorrect.html
	IncorrectAnswer string

	//go:embed testdata/answers/rate_limited.html
	RateLimitedAnswer string

	//go:embed testdata/answers/wrong_level.html
	WrongLevelAnswer string
)

func submitWithResponse(t *testing.T, page string) *SubmitResult {
	httpmock.RegisterResponder(http.MethodPost, AnswerURL, func(r *http.Request) (*http.Response, error) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}

		// AoC expects a form post with the level and answer
		if r.PostForm.Get("level") != "1" || r.PostForm.Get("answer") != "11" {
			return httpmock.NewStringResponse(http.StatusBadRequest, "bad form"), nil
		}

		return htmlResponder(page)(r)
	})

	result, err := NewEventFetcher(http.DefaultClient).SubmitAnswer(context.Background(), "2024", 1, 1, "11")

	if err != nil {
		t.Fatal(err)
	}

	return result
}

func TestSubmitCorrect(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	result := submitWithResponse(t, CorrectAnswer)

	if result.Outcome != OutcomeCorrect {
		t.Errorf("unexpected outcome, expected: %s, got: %s", OutcomeCorrect, result.Outcome)
		return
	}

	expected := "That's the right answer! You are one gold star closer to finding the Chief Historian."

	if result.Message != expected {
		t.Errorf("unexpected message, expected: %s, got: %s", expected, result.Message)
		return
	}
}

func TestSubmitIncorrect(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	result := submitWithResponse(t, IncorrectAnswer)

	if result.Outcome != OutcomeIncorrect {
		t.Errorf("unexpected outcome, expected: %s, got: %s", OutcomeIncorrect, result.Outcome)
		return
	}

	if !strings.HasPrefix(result.Message, "That's not the right answer; your answer is too high.") {
		t.Errorf("unexpected message, got: %s", result.Message)
		return
	}
}

func TestSubmitRateLimited(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	result := submitWithResponse(t, RateLimitedAnswer)

	if result.Outcome != OutcomeRateLimited {
		t.Errorf("unexpected outcome, expected: %s, got: %s", OutcomeRateLimited, result.Outcome)
		return
	}

	if result.WaitSeconds != 65 {
		t.Errorf("unexpected wait, expected: 65, got: %d", result.WaitSeconds)
		return
	}
}

func TestSubmitWrongLevel(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	result := submitWithResponse(t, WrongLevelAnswer)

	if result.Outcome != OutcomeWrongLevel {
		t.Errorf("unexpected outcome, expected: %s, got: %s", OutcomeWrongLevel, result.Outcome)
		return
	}
}

func TestSubmitUnknownPage(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	result := submitWithResponse(t, `<html><body><main><article><p>Something new.</p></article></main></body></html>`)

	if result.Outcome != OutcomeUnknown {
		t.Errorf("unexpected outcome, expected: %s, got: %s", OutcomeUnknown, result.Outcome)
		return
	}

	if result.Message != "Something new." {
		t.Errorf("unexpected message, expected: Something new., got: %s", result.Message)
		return
	}
}

func TestWaitSeconds(t *testing.T) {
	for message, expected := range map[string]int{
		"You have 49s left to wait.":    49,
		"You have 1m 5s left to wait.":  65,
		"You have 10m 0s left to wait.": 600,
		"no wait mentioned":             0,
	} {
		if seconds := waitSeconds(message); seconds != expected {
			t.Errorf("unexpected wait for %q, expected: %d, got: %d", message, expected, seconds)
			return
		}
	}
}
