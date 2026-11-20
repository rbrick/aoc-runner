package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/puzzle"
	"github.com/rbrick/aoc-runner/internal/session"
)

func setupRouter(cookies ...string) *echo.Echo {
	httpmock.RegisterResponder(http.MethodGet, puzzle.BaseURL,
		httpmock.NewStringResponder(http.StatusOK, `<div class="user">someone</div>`))

	aocSession := session.NewSession(func(ctx context.Context) ([]*http.Cookie, error) {
		var found []*http.Cookie

		for _, value := range cookies {
			found = append(found, &http.Cookie{
				Name:  session.CookieName,
				Value: value,
			})
		}

		return found, nil
	})

	router := echo.New()
	NewHandler(aocSession).RegisterRoutes(router.Group("/auth"))

	return router
}

func request(router *echo.Echo, method, path string) (*httptest.ResponseRecorder, *SessionDTO, error) {
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader("")))

	var result SessionDTO

	err := json.Unmarshal(rec.Body.Bytes(), &result)

	return rec, &result, err
}

func TestLogin(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	router := setupRouter("good")

	rec, result, err := request(router, http.MethodPost, "/auth/login")

	if err != nil {
		t.Error(err)
		return
	}

	if rec.Code != http.StatusOK {
		t.Errorf("unexpected status, expected: 200, got: %d", rec.Code)
		return
	}

	if !result.LoggedIn {
		t.Error("expected to be logged in")
		return
	}

	_, result, err = request(router, http.MethodGet, "/auth/")

	if err != nil {
		t.Error(err)
		return
	}

	if !result.LoggedIn {
		t.Error("expected status to be logged in after login")
		return
	}
}

func TestLoginNoSession(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	router := setupRouter()

	rec, result, err := request(router, http.MethodPost, "/auth/login")

	if err != nil {
		t.Error(err)
		return
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unexpected status, expected: 401, got: %d", rec.Code)
		return
	}

	if result.LoginURL != session.LoginURL {
		t.Errorf("unexpected login url, expected: %s, got: %s", session.LoginURL, result.LoginURL)
		return
	}
}

func TestLogout(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	router := setupRouter("good")

	if _, _, err := request(router, http.MethodPost, "/auth/login"); err != nil {
		t.Error(err)
		return
	}

	rec, result, err := request(router, http.MethodPost, "/auth/logout")

	if err != nil {
		t.Error(err)
		return
	}

	if rec.Code != http.StatusOK {
		t.Errorf("unexpected status, expected: 200, got: %d", rec.Code)
		return
	}

	if result.LoggedIn {
		t.Error("expected to be logged out")
		return
	}

	_, result, err = request(router, http.MethodGet, "/auth/")

	if err != nil {
		t.Error(err)
		return
	}

	if result.LoggedIn {
		t.Error("expected status to be logged out after logout")
		return
	}
}

func TestStatusLoggedOut(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	_, result, err := request(setupRouter(), http.MethodGet, "/auth/")

	if err != nil {
		t.Error(err)
		return
	}

	if result.LoggedIn {
		t.Error("expected to be logged out")
		return
	}

	if result.LoginURL != session.LoginURL {
		t.Errorf("unexpected login url, expected: %s, got: %s", session.LoginURL, result.LoginURL)
		return
	}
}
