package puzzle

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// stars earned for each day of a year, only shown on the calendar when logged in
func (f *EventFetcher) FetchStars(ctx context.Context, year string) (map[int]int, error) {
	calendarResponse, err := f.makeGet(ctx, BaseURL+"/"+year)

	if err != nil {
		return nil, err
	}

	defer calendarResponse.Body.Close()

	return f.parseStars(calendarResponse.Body)
}

func (f *EventFetcher) parseStars(htmlBody io.Reader) (map[int]int, error) {
	stars := make(map[int]int)

	doc, err := html.Parse(htmlBody)

	if err != nil {
		return nil, err
	}

	filter := func(n *html.Node) bool {
		return checkAttrPattern(n.Attr, "class", calendaryDayPattern) &&
			n.Type == html.ElementNode &&
			n.Data == "a"
	}

	// <a aria-label="Day 1, two stars" class="calendar-day1 calendar-verycomplete">
	apply := func(n *html.Node) error {
		day := 0

		for _, class := range strings.Fields(extractAttr(n.Attr, "class")) {
			if matches := calendaryDayPattern.FindStringSubmatch(class); len(matches) > 1 {
				day, _ = strconv.Atoi(matches[1])
			}
		}

		switch {
		case checkAttr(n.Attr, "class", "calendar-verycomplete"):
			stars[day] = 2
		case checkAttr(n.Attr, "class", "calendar-complete"):
			stars[day] = 1
		default:
			stars[day] = 0
		}

		return nil
	}

	if err := f.eachNodeFunc(doc, filter, apply); err != nil {
		return nil, err
	}

	return stars, nil
}

// stars earned for each year, only shown on the events page when logged in
func (f *EventFetcher) FetchEventStars(ctx context.Context) (map[string]int, error) {
	eventsResponse, err := f.makeGet(ctx, EventsURL)

	if err != nil {
		return nil, err
	}

	defer eventsResponse.Body.Close()

	return f.parseEventStars(eventsResponse.Body)
}

func (f *EventFetcher) parseEventStars(htmlBody io.Reader) (map[string]int, error) {
	stars := make(map[string]int)

	doc, err := html.Parse(htmlBody)

	if err != nil {
		return nil, err
	}

	filter := func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" &&
			checkAttr(n.Attr, "class", "eventlist-event") &&
			n.FirstChild != nil && n.FirstChild.FirstChild != nil
	}

	// <div class="eventlist-event"><a href="/2018">[2018]</a> <span class="star-count"> 2*</span></div>
	apply := func(n *html.Node) error {
		year := YearExtractor(n.FirstChild.FirstChild.Data)

		for child := range n.ChildNodes() {
			if child.Type == html.ElementNode && checkAttr(child.Attr, "class", "star-count") {
				count, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(extractText(child), "*")))

				if err == nil {
					stars[year] = count
				}
			}
		}

		return nil
	}

	if err := f.eachNodeFunc(doc, filter, apply); err != nil {
		return nil, err
	}

	return stars, nil
}

// the user's puzzle input, only available when logged in
func (f *EventFetcher) FetchInput(ctx context.Context, year string, day int) (string, error) {
	url := fmt.Sprintf("%s/%s/day/%d/input", BaseURL, year, day)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", UserAgent)

	response, err := f.client.Do(req)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	// AoC answers 400 "Puzzle inputs differ by user. Please log in..." without a session
	if response.StatusCode != http.StatusOK {
		return "", &FetchError{
			URL:        url,
			StatusCode: response.StatusCode,
			Err:        fmt.Errorf("%s", strings.TrimSpace(string(body))),
		}
	}

	return string(body), nil
}
