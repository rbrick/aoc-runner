package puzzle

import (
	"context"
	"io"
	"mime"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

// The URL for the series of events
const (
	EventsURL = "https://adventofcode.com/events"
	BaseURL   = "https://adventofcode.com"

	// sent with every request, AoC asks automated tools to identify themselves
	UserAgent = "aoc-runner/1.0"
)

var (
	yearPattern         = regexp.MustCompile(`^\[([0-9]{4})\]$`)
	calendaryDayPattern = regexp.MustCompile(`^calendar-day([0-9]{1,2})$`)
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

type Problem struct {
	Body     string   `json:"body"`
	Examples []string `json:"examples"`
	// one per part, empty when the answer couldn't be found
	ExampleAnswers []string `json:"example_answers"`
}

type EventFetcher struct {
	client *http.Client
}

func (f *EventFetcher) makeGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	response, err := f.client.Do(req)

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

func (f *EventFetcher) Fetch(ctx context.Context, eager bool) ([]Event, error) {

	resp, err := f.makeGet(ctx, EventsURL)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	// Parse HTML body
	return f.Parse(ctx, resp.Body, eager)
}

func checkAttr(attrs []html.Attribute, k, v string) bool {
	for _, attr := range attrs {
		if attr.Key == k {
			if slices.Contains(strings.Fields(attr.Val), v) {
				return true
			}
		}
	}
	return false
}

func checkAttrPattern(attrs []html.Attribute, k string, pattern *regexp.Regexp) bool {
	for _, attr := range attrs {
		if attr.Key == k {
			if slices.ContainsFunc(strings.Fields(attr.Val), pattern.MatchString) {
				return true
			}
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

// highlighted code is either <code><em>142</em></code>, <em><code>24000</code></em>
// or the result at the end of a calculation, <code>1749 - 161 = <em>1588</em></code>
func isHighlightedCode(n *html.Node) bool {
	if n.Type != html.ElementNode || n.Parent == nil {
		return false
	}

	if n.Data == "code" && n.Parent.Data == "em" {
		// only when it's wrapped exactly, not <em>sum to <code>2020</code></em>
		return n.Parent.FirstChild == n && n.Parent.LastChild == n
	}

	if n.Data == "em" && n.Parent.Data == "code" {
		// highlights inside an example block aren't answers
		if n.Parent.Parent != nil && n.Parent.Parent.Data == "pre" {
			return false
		}

		return n.Parent.LastChild == n
	}

	return false
}

// joins all the text inside a node, examples can have <em> highlights inside them
func extractText(n *html.Node) string {
	var text strings.Builder

	for child := range n.Descendants() {
		if child.Type == html.TextNode {
			text.WriteString(child.Data)
		}
	}

	return text.String()
}

func (f *EventFetcher) fetchDays(ctx context.Context, event *Event) ([]Day, error) {
	eventResponse, err := f.makeGet(ctx, event.URL)
	if err != nil {
		return nil, err
	}

	defer eventResponse.Body.Close()

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

func (f *EventFetcher) FetchProblem(ctx context.Context, day *Day) (*Problem, error) {
	dayResponse, err := f.makeGet(ctx, day.URL)

	if err != nil {
		return nil, err
	}

	defer dayResponse.Body.Close()

	return f.parseProblem(dayResponse.Body)
}

func (f *EventFetcher) parseProblem(htmlBody io.Reader) (*Problem, error) {
	var body strings.Builder
	var examples []string
	var exampleAnswers []string

	doc, err := html.Parse(htmlBody)

	if err != nil {
		return nil, err
	}

	// part two is a second article, only shown once part one is solved
	filter := func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "article" &&
			checkAttr(n.Attr, "class", "day-desc")
	}

	// examples are in <pre><code> blocks, not every block is an input though (some are diagrams)
	exampleFilter := func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "pre"
	}

	exampleApply := func(n *html.Node) error {
		examples = append(examples, extractText(n))
		return nil
	}

	// the example answer is usually the last highlighted value in a part
	answer := ""

	answerApply := func(n *html.Node) error {
		answer = extractText(n)
		return nil
	}

	apply := func(n *html.Node) error {
		if err := f.eachNodeFunc(n, exampleFilter, exampleApply); err != nil {
			return err
		}

		answer = ""

		if err := f.eachNodeFunc(n, isHighlightedCode, answerApply); err != nil {
			return err
		}

		exampleAnswers = append(exampleAnswers, answer)

		return html.Render(&body, n)
	}

	if err := f.eachNodeFunc(doc, filter, apply); err != nil {
		return nil, err
	}

	return &Problem{
		Body:           body.String(),
		Examples:       examples,
		ExampleAnswers: exampleAnswers,
	}, nil
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

func (f *EventFetcher) Parse(ctx context.Context, htmlBody io.Reader, eager bool) ([]Event, error) {
	var events []Event

	doc, err := html.Parse(htmlBody)

	if err != nil {
		return nil, err
	}

	parseEvent := func(child *html.Node) (*Event, error) {
		link := extractAttr(child.Attr, "href")
		text := child.FirstChild.Data

		event := &Event{
			URL:  BaseURL + link,
			Year: YearExtractor(text),
			Days: make([]Day, 0),
		}

		if eager {
			days, err := f.fetchDays(ctx, event)

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
