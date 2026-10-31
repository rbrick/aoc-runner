package puzzle

import "fmt"

func UnexpectedContentType(expected, actual string) error {
	return fmt.Errorf("unexpected content type: expected %s, actual: %s", expected, actual)
}

type FetchError struct {
	StatusCode int
	URL        string
	Err        error
}

func (e *FetchError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("failed to fetch. status_code = %d, url = %s", e.StatusCode, e.URL)
	}

	return fmt.Sprintf("failed to fetch. status_code = %d, url = %s, err = %s", e.StatusCode, e.URL, e.Err)
}
