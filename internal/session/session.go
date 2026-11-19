package session

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/rbrick/aoc-runner/internal/puzzle"
)

const (
	Domain     = "adventofcode.com"
	CookieName = "session"
	LoginURL   = puzzle.BaseURL + "/auth/login"
)

var (
	// only shown on pages when logged in
	loggedInMarker = []byte(`<div class="user">`)
)

type Session struct {
	mu          sync.RWMutex
	cookie      *http.Cookie
	readCookies CookieReader
}

func (s *Session) Login(ctx context.Context) error {
	cookies, err := s.readCookies(ctx)

	if err != nil {
		return err
	}

	// the browser can have old sessions that AoC no longer accepts
	for _, cookie := range cookies {
		valid, err := verify(ctx, cookie)

		if err != nil {
			return err
		}

		if valid {
			s.mu.Lock()
			s.cookie = cookie
			s.mu.Unlock()

			return nil
		}
	}

	return ErrNotLoggedIn
}

func (s *Session) LoggedIn() bool {
	return s.Cookie() != nil
}

func (s *Session) Cookie() *http.Cookie {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.cookie
}

// a client that sends the session cookie with every request to Advent of Code
func (s *Session) Client() *http.Client {
	return &http.Client{
		Transport: &sessionTransport{
			session: s,
		},
	}
}

func verify(ctx context.Context, cookie *http.Cookie) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, puzzle.BaseURL, nil)

	if err != nil {
		return false, err
	}

	req.AddCookie(cookie)

	response, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, nil
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return false, err
	}

	return bytes.Contains(body, loggedInMarker), nil
}

type sessionTransport struct {
	session *Session
}

func (t *sessionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cookie := t.session.Cookie()

	// never send the session anywhere other than Advent of Code
	if cookie != nil && req.URL.Hostname() == Domain {
		req = req.Clone(req.Context())
		req.AddCookie(cookie)
	}

	return http.DefaultTransport.RoundTrip(req)
}

func NewSession(readCookies CookieReader) *Session {
	return &Session{
		readCookies: readCookies,
	}
}
