package puzzle

import (
	"io"
	"mime"
	"net/http"
	"regexp"

	"golang.org/x/net/html"
)

// The URL for the series of events
const (
	EventsURL = "https://adventofcode.com/events"
	BaseURL   = "https://adventofcode.com"
)

var (
	yearPattern         = regexp.MustCompile(`\[([0-9]{4})\]`)
	calendaryDayPattern = regexp.MustCompile(`calendar-day([0-9]{1,2})`)
)

func YearExtractor(year string) string {
	matches := yearPattern.FindStringSubmatch(year)

	if len(matches) <= 1 {
		// no groups
		return ""
	}

	return matches[1]
}

type Event struct {
	Year string `json:"year"`
	URL  string `json:"url"`
	Days []Day  `json:"days"`
}

type Day struct {
	Day     string `json:"day"`
	URL     string `json:"url"`
	Problem string `json:"problem"`
}

type EventFetcher struct {
	client *http.Client
}

func (f *EventFetcher) makeGet(url string) (*http.Response, error) {
	response, err := f.client.Get(url)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, &FetchError{
			URL:        url,
			StatusCode: response.StatusCode,
			Err:        err,
		}
	}

	contentType := response.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)

	if err != nil {
		return nil, &FetchError{
			URL:        url,
			StatusCode: response.StatusCode,
			Err:        err,
		}
	}

	if mediaType != "text/html" {
		// unsupported/unexpected body
		return nil, UnexpectedContentType("text/html", mediaType)
	}

	return response, nil
}

func (f *EventFetcher) Fetch(eager bool) ([]Event, error) {
	resp, err := f.makeGet(EventsURL)

	if err != nil {
		return nil, err
	}

	// Parse HTML body
	return f.Parse(resp.Body, eager)
}

func checkAttr(attrs []html.Attribute, k, v string) bool {
	for _, attr := range attrs {
		if attr.Key == k && attr.Val == v {
			return true
		}
	}

	return false
}

func checkAttrPattern(attrs []html.Attribute, k string, pattern *regexp.Regexp) bool {
	for _, attr := range attrs {
		if attr.Key == k && pattern.MatchString(attr.Val) {
			return true
		}
	}

	return false
}

func extractAttr(attrs []html.Attribute, key string) string {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val
		}
	}

	return ""
}

func (f *EventFetcher) fetchDays(event *Event) ([]Day, error) {
	eventResponse, err := f.makeGet(event.URL)

	if err != nil {
		return nil, err
	}

	return f.parseDays(eventResponse.Body)
}

func (f *EventFetcher) parseDays(htmlBody io.Reader) ([]Day, error) {
	var days []Day

	doc, err := html.Parse(htmlBody)

	if err != nil {
		return nil, err
	}

	filter := func(n *html.Node) bool {
		return checkAttrPattern(n.Attr, "class", calendaryDayPattern) &&
			n.Type == html.ElementNode &&
			n.Data == "a"
	}

	apply := func(n *html.Node) error {
		link := extractAttr(n.Attr, "href")
		name := extractAttr(n.Attr, "aria-label")

		days = append(days, Day{
			URL:     BaseURL + link,
			Day:     name,
			Problem: "",
		})

		return nil
	}

	if err := f.eachNodeFunc(doc, filter, apply); err != nil {
		return nil, err
	}

	return days, nil
}

func (f *EventFetcher) eachNodeFunc(root *html.Node, filter func(*html.Node) bool, apply func(*html.Node) error) error {
	for n := range root.Descendants() {
		if filter(n) {
			if err := apply(n); err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *EventFetcher) Parse(htmlBody io.Reader, eager bool) ([]Event, error) {
	var events []Event

	doc, err := html.Parse(htmlBody)

	if err != nil {
		return nil, err
	}

	parseEvent := func(child *html.Node) (*Event, error) {
		link := child.Attr[0].Val
		text := child.FirstChild.Data

		event := &Event{
			URL:  BaseURL + link,
			Year: YearExtractor(text),
			Days: make([]Day, 0),
		}

		if eager {
			days, err := f.fetchDays(event)

			if err != nil {
				return nil, err
			}

			event.Days = days
		}

		return event, nil
	}

	filter := func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" &&
			checkAttr(n.Attr, "class", "eventlist-event") &&
			n.FirstChild != nil
	}

	apply := func(n *html.Node) error {
		event, err := parseEvent(n.FirstChild)

		if err != nil {
			return err
		}

		events = append(events, *event)
		return nil
	}

	if err := f.eachNodeFunc(doc, filter, apply); err != nil {
		return nil, err
	}

	return events, nil
}

func NewEventFetcher(client *http.Client) *EventFetcher {
	return &EventFetcher{
		client: client,
	}
}
