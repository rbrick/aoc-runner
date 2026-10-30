package puzzle

import (
	"io"
	"mime"
	"net/http"

	"golang.org/x/net/html"
)

// The URL for the series of events
const EventsURL = "https://adventofcode.com/events"

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

func (f *EventFetcher) Fetch() ([]Event, error) {
	response, err := f.client.Get(EventsURL)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, &FetchError{
			URL:        EventsURL,
			StatusCode: response.StatusCode,
			Err:        err,
		}
	}

	contentType := response.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)

	if err != nil {
		return nil, &FetchError{
			URL:        EventsURL,
			StatusCode: response.StatusCode,
			Err:        err,
		}
	}

	if mediaType != "text/html" {
		// unsupported/unexpected body
		return nil, UnexpectedContentType("text/html", mediaType)
	}

	// Parse HTML body
	return f.Parse(response.Body)
}

func (f *EventFetcher) Parse(htmlBody io.Reader) ([]Event, error) {
	var events []Event

	doc, err := html.Parse(htmlBody)

	if err != nil {
		return nil, err
	}

	checkAttr := func(attrs []html.Attribute) bool {
		for _, attr := range attrs {
			if attr.Key == "class" && attr.Val == "eventlist-event" {
				return true
			}
		}

		return false
	}

	parseEvent := func(child *html.Node) *Event {
		link := child.Attr[0].Val
		text := child.FirstChild.Data

		return &Event{
			URL:  EventsURL + link,
			Year: text,
			Days: make([]Day, 0),
		}
	}

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "div" {
			// check attr
			if checkAttr(n.Attr) && n.FirstChild != nil {
				event := parseEvent(n.FirstChild)
				if event != nil {
					events = append(events, *event)
				}
			}
		}
	}

	return events, nil
}

func NewEventFetcher(client *http.Client) *EventFetcher {
	return &EventFetcher{
		client: client,
	}
}
