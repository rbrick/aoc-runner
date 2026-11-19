package puzzle

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Outcome string

const (
	OutcomeCorrect     Outcome = "correct"
	OutcomeIncorrect   Outcome = "incorrect"
	OutcomeRateLimited Outcome = "rate_limited"
	// the level is already solved, or part two isn't unlocked yet
	OutcomeWrongLevel Outcome = "wrong_level"
	OutcomeUnknown    Outcome = "unknown"
)

var (
	waitPattern = regexp.MustCompile(`You have (?:(\d+)m )?(\d+)s left to wait`)
)

type SubmitResult struct {
	Outcome Outcome
	// AoC's own message, like "That's not the right answer; your answer is too high."
	Message string
	// how long until another answer can be submitted, when rate limited
	WaitSeconds int
}

// submits an answer for part one (level 1) or part two (level 2), only works when logged in
func (f *EventFetcher) SubmitAnswer(ctx context.Context, year string, day, level int, answer string) (*SubmitResult, error) {
	answerURL := fmt.Sprintf("%s/%s/day/%d/answer", BaseURL, year, day)

	form := url.Values{
		"level":  {strconv.Itoa(level)},
		"answer": {answer},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, answerURL, strings.NewReader(form.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	response, err := f.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)

		return nil, &FetchError{
			URL:        answerURL,
			StatusCode: response.StatusCode,
			Err:        fmt.Errorf("%s", strings.TrimSpace(string(body))),
		}
	}

	return f.parseSubmitResult(response.Body)
}

func (f *EventFetcher) parseSubmitResult(htmlBody io.Reader) (*SubmitResult, error) {
	var message string

	doc, err := html.Parse(htmlBody)

	if err != nil {
		return nil, err
	}

	filter := func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "article"
	}

	apply := func(n *html.Node) error {
		message = strings.Join(strings.Fields(extractText(n)), " ")
		return nil
	}

	if err := f.eachNodeFunc(doc, filter, apply); err != nil {
		return nil, err
	}

	// drop the trailing link, like "[Return to Day 1]"
	if index := strings.LastIndex(message, " ["); index != -1 && strings.HasSuffix(message, "]") {
		message = message[:index]
	}

	result := &SubmitResult{
		Outcome: OutcomeUnknown,
		Message: message,
	}

	switch {
	case strings.Contains(message, "That's the right answer"):
		result.Outcome = OutcomeCorrect
	case strings.Contains(message, "That's not the right answer"):
		result.Outcome = OutcomeIncorrect
	case strings.Contains(message, "You gave an answer too recently"):
		result.Outcome = OutcomeRateLimited
		result.WaitSeconds = waitSeconds(message)
	case strings.Contains(message, "You don't seem to be solving the right level"):
		result.Outcome = OutcomeWrongLevel
	}

	return result, nil
}

// "You have 1m 5s left to wait" is 65 seconds
func waitSeconds(message string) int {
	matches := waitPattern.FindStringSubmatch(message)

	if len(matches) != 3 {
		return 0
	}

	minutes, _ := strconv.Atoi(matches[1])
	seconds, _ := strconv.Atoi(matches[2])

	return minutes*60 + seconds
}
