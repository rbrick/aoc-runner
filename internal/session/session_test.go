package session

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rbrick/aoc-runner/internal/puzzle"

	_ "embed"
)

const (
	ValidSession = "good"
)

var (
	//go:embed testdata/logged_in.html
	LoggedInPage string

	//go:embed testdata/logged_out.html
	LoggedOutPage string
)

func cookies(values ...string) CookieReader {
	return func(ctx context.Context) ([]*http.Cookie, error) {
		var found []*http.Cookie

		for _, value := range values {
			found = append(found, &http.Cookie{
				Name:  CookieName,
				Value: value,
			})
		}

		return found, nil
	}
}

func setupAdventOfCode() {
	// AoC only shows the user when the session is valid
	httpmock.RegisterResponder(http.MethodGet, puzzle.BaseURL, func(r *http.Request) (*http.Response, error) {
		page := LoggedOutPage

		if cookie, err := r.Cookie(CookieName); err == nil && cookie.Value == ValidSession {
			page = LoggedInPage
		}

		return httpmock.NewStringResponse(http.StatusOK, page), nil
	})
}

func TestLogin(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	setupAdventOfCode()

	session := NewSession(cookies("expired", ValidSession))

	if err := session.Login(context.Background()); err != nil {
		t.Error(err)
		return
	}

	if !session.LoggedIn() {
		t.Error("expected to be logged in")
		return
	}

	if session.Cookie().Value != ValidSession {
		t.Errorf("unexpected session, expected: %s, got: %s", ValidSession, session.Cookie().Value)
		return
	}
}

func TestLoginNoSession(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	setupAdventOfCode()

	session := NewSession(cookies())

	if err := session.Login(context.Background()); !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("unexpected error, expected: %s, got: %v", ErrNotLoggedIn, err)
		return
	}

	if session.LoggedIn() {
		t.Error("expected to be logged out")
		return
	}
}

func TestLoginExpiredSession(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	setupAdventOfCode()

	session := NewSession(cookies("expired"))

	if err := session.Login(context.Background()); !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("unexpected error, expected: %s, got: %v", ErrNotLoggedIn, err)
		return
	}
}

func TestLoginReadError(t *testing.T) {
	readErr := errors.New("keyring locked")

	session := NewSession(func(ctx context.Context) ([]*http.Cookie, error) {
		return nil, readErr
	})

	if err := session.Login(context.Background()); !errors.Is(err, readErr) {
		t.Errorf("unexpected error, expected: %s, got: %v", readErr, err)
		return
	}
}

func TestClientSendsSession(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	setupAdventOfCode()

	var sentToAdventOfCode, sentElsewhere bool

	httpmock.RegisterResponder(http.MethodGet, puzzle.BaseURL+"/2024/day/1", func(r *http.Request) (*http.Response, error) {
		_, err := r.Cookie(CookieName)
		sentToAdventOfCode = err == nil

		return httpmock.NewStringResponse(http.StatusOK, ""), nil
	})

	httpmock.RegisterResponder(http.MethodGet, "https://example.com/", func(r *http.Request) (*http.Response, error) {
		_, err := r.Cookie(CookieName)
		sentElsewhere = err == nil

		return httpmock.NewStringResponse(http.StatusOK, ""), nil
	})

	session := NewSession(cookies(ValidSession))

	if err := session.Login(context.Background()); err != nil {
		t.Error(err)
		return
	}

	client := session.Client()

	if _, err := client.Get(puzzle.BaseURL + "/2024/day/1"); err != nil {
		t.Error(err)
		return
	}

	if _, err := client.Get("https://example.com/"); err != nil {
		t.Error(err)
		return
	}

	if !sentToAdventOfCode {
		t.Error("expected the session to be sent to Advent of Code")
		return
	}

	if sentElsewhere {
		t.Error("expected the session to only be sent to Advent of Code")
		return
	}
}
