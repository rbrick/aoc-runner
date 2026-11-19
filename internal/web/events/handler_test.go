package events

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/database"
	"github.com/rbrick/aoc-runner/internal/model"
	"github.com/rbrick/aoc-runner/internal/puzzle"
	"github.com/rbrick/aoc-runner/internal/session"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "embed"
)

const (
	InputURL  = puzzle.BaseURL + "/2018/day/1/input"
	AnswerURL = puzzle.BaseURL + "/2018/day/1/answer"
)

var (
	//go:embed testdata/calendar.html
	Calendar string

	//go:embed testdata/answer_correct.html
	CorrectAnswer string

	//go:embed testdata/answer_incorrect.html
	IncorrectAnswer string
)

func setupDatabase(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"))

	if err != nil {
		t.Fatal(err)
	}

	// every connection to :memory: is its own database
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	event := &model.Event{
		Year: "2018",
		Puzzles: []model.Puzzle{
			{Day: "Day 2"},
			{Day: "Day 1"},
			{Day: "Day 3"},
		},
	}

	if err := db.Create(event).Error; err != nil {
		t.Fatal(err)
	}

	return db
}

func setupSession(t *testing.T, loggedIn bool) *session.Session {
	httpmock.RegisterResponder(http.MethodGet, puzzle.BaseURL,
		httpmock.NewStringResponder(http.StatusOK, `<div class="user">someone</div>`))

	aocSession := session.NewSession(func(ctx context.Context) ([]*http.Cookie, error) {
		return []*http.Cookie{{Name: session.CookieName, Value: "good"}}, nil
	})

	if loggedIn {
		if err := aocSession.Login(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	return aocSession
}

func setupRouter(db *gorm.DB, aocSession *session.Session) *echo.Echo {
	router := echo.New()

	NewHandler(db, puzzle.NewEventFetcher(aocSession.Client()), aocSession).RegisterRoutes(router.Group("/events"))

	return router
}

func get(router *echo.Echo, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

	return rec
}

func TestListEventDaysWithStars(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(http.MethodGet, puzzle.BaseURL+"/2018", func(r *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, Calendar)
		response.Header.Set("Content-Type", "text/html")

		return response, nil
	})

	rec := get(setupRouter(setupDatabase(t), setupSession(t, true)), "/events/2018")

	var result EventDaysDTO

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Error(err)
		return
	}

	if len(result.Days) != 3 {
		t.Errorf("unexpected days length, expected: 3, got: %d", len(result.Days))
		return
	}

	for index, expected := range []int{2, 1, 0} {
		day := result.Days[index]

		if day.Number != index+1 {
			t.Errorf("unexpected day order, expected: %d, got: %d", index+1, day.Number)
			return
		}

		if day.Stars == nil || *day.Stars != expected {
			t.Errorf("unexpected stars for day %d, expected: %d, got: %v", day.Number, expected, day.Stars)
			return
		}
	}
}

func TestListEventDaysLoggedOut(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	rec := get(setupRouter(setupDatabase(t), setupSession(t, false)), "/events/2018")

	var result EventDaysDTO

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Error(err)
		return
	}

	if result.Days[0].Stars != nil {
		t.Errorf("unexpected stars when logged out, got: %d", *result.Days[0].Stars)
		return
	}
}

func TestGetInput(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(http.MethodGet, InputURL, httpmock.NewStringResponder(http.StatusOK, "+1\n-2\n"))

	db := setupDatabase(t)
	router := setupRouter(db, setupSession(t, true))

	for range 2 {
		rec := get(router, "/events/2018/1/input")

		if rec.Code != http.StatusOK {
			t.Errorf("unexpected status, expected: 200, got: %d", rec.Code)
			return
		}

		if rec.Body.String() != "+1\n-2\n" {
			t.Errorf("unexpected input, expected: %q, got: %q", "+1\n-2\n", rec.Body.String())
			return
		}
	}

	// the second request is served from the database
	if calls := httpmock.GetCallCountInfo()["GET "+InputURL]; calls != 1 {
		t.Errorf("unexpected input downloads, expected: 1, got: %d", calls)
		return
	}

	var stored model.Puzzle

	if err := db.Where("day = ?", "Day 1").First(&stored).Error; err != nil {
		t.Error(err)
		return
	}

	if stored.Input != "+1\n-2\n" {
		t.Errorf("unexpected stored input, expected: %q, got: %q", "+1\n-2\n", stored.Input)
		return
	}
}

func TestGetInputLoggedOut(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	rec := get(setupRouter(setupDatabase(t), setupSession(t, false)), "/events/2018/1/input")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unexpected status, expected: 401, got: %d", rec.Code)
		return
	}

	if calls := httpmock.GetCallCountInfo()["GET "+InputURL]; calls != 0 {
		t.Errorf("unexpected input downloads, expected: 0, got: %d", calls)
		return
	}
}

func TestGetInputUnknownDay(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	rec := get(setupRouter(setupDatabase(t), setupSession(t, true)), "/events/2018/9/input")

	if rec.Code != http.StatusNotFound {
		t.Errorf("unexpected status, expected: 404, got: %d", rec.Code)
		return
	}
}

func answerResponder(page string) httpmock.Responder {
	return func(r *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, page)
		response.Header.Set("Content-Type", "text/html")

		return response, nil
	}
}

func submit(router *echo.Echo, body string) (*httptest.ResponseRecorder, *SubmitResultDTO) {
	req := httptest.NewRequest(http.MethodPost, "/events/2018/1/answer", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var result SubmitResultDTO
	json.Unmarshal(rec.Body.Bytes(), &result)

	return rec, &result
}

func TestSubmitAnswerCorrect(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(http.MethodPost, AnswerURL, answerResponder(CorrectAnswer))

	db := setupDatabase(t)
	rec, result := submit(setupRouter(db, setupSession(t, true)), `{"level":1,"answer":"576"}`)

	if rec.Code != http.StatusOK {
		t.Errorf("unexpected status, expected: 200, got: %d", rec.Code)
		return
	}

	if result.Outcome != string(puzzle.OutcomeCorrect) {
		t.Errorf("unexpected outcome, expected: correct, got: %s", result.Outcome)
		return
	}

	var stored model.Puzzle

	if err := db.Where("day = ?", "Day 1").First(&stored).Error; err != nil {
		t.Error(err)
		return
	}

	if stored.Status != model.StatusPartOneSolved {
		t.Errorf("unexpected puzzle status, expected: %d, got: %d", model.StatusPartOneSolved, stored.Status)
		return
	}
}

func TestSubmitAnswerIncorrectIsNotResent(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(http.MethodPost, AnswerURL, answerResponder(IncorrectAnswer))

	router := setupRouter(setupDatabase(t), setupSession(t, true))

	_, first := submit(router, `{"level":1,"answer":"999"}`)

	if first.Outcome != string(puzzle.OutcomeIncorrect) || first.Duplicate {
		t.Errorf("unexpected first result, expected: incorrect, got: %+v", first)
		return
	}

	// surrounding whitespace doesn't make it a different answer
	_, second := submit(router, `{"level":1,"answer":" 999\n"}`)

	if second.Outcome != string(puzzle.OutcomeIncorrect) || !second.Duplicate {
		t.Errorf("unexpected second result, expected: a duplicate incorrect, got: %+v", second)
		return
	}

	if calls := httpmock.GetCallCountInfo()["POST "+AnswerURL]; calls != 1 {
		t.Errorf("unexpected answers sent to AoC, expected: 1, got: %d", calls)
		return
	}

	// the same answer for the other part is a different submission
	submit(router, `{"level":2,"answer":"999"}`)

	if calls := httpmock.GetCallCountInfo()["POST "+AnswerURL]; calls != 2 {
		t.Errorf("unexpected answers sent to AoC, expected: 2, got: %d", calls)
		return
	}
}

func TestSubmitAnswerLoggedOut(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	rec, _ := submit(setupRouter(setupDatabase(t), setupSession(t, false)), `{"level":1,"answer":"576"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unexpected status, expected: 401, got: %d", rec.Code)
		return
	}

	if calls := httpmock.GetCallCountInfo()["POST "+AnswerURL]; calls != 0 {
		t.Errorf("unexpected answers sent to AoC, expected: 0, got: %d", calls)
		return
	}
}

func TestSubmitAnswerValidation(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	router := setupRouter(setupDatabase(t), setupSession(t, true))

	for _, body := range []string{`{"level":3,"answer":"1"}`, `{"level":1,"answer":"  "}`, `{`} {
		rec, _ := submit(router, body)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("unexpected status for %s, expected: 400, got: %d", body, rec.Code)
			return
		}
	}

	if calls := httpmock.GetCallCountInfo()["POST "+AnswerURL]; calls != 0 {
		t.Errorf("unexpected answers sent to AoC, expected: 0, got: %d", calls)
		return
	}
}
