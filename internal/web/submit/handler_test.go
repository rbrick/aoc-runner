package submit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	_ "embed"
)

var (
	//go:embed testdata/valid.java
	ValidSolution string

	//go:embed testdata/throws.java
	ThrowingSolution string
)

func skipWithoutJava(t *testing.T) {
	if _, err := exec.LookPath("javac"); err != nil {
		t.Skip("javac is not installed")
	}

	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java is not installed")
	}
}

func setupRouter() *echo.Echo {
	router := echo.New()

	NewHandler().RegisterRoutes(router.Group("/submit"))

	return router
}

func submit(router *echo.Echo, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/submit/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func submitCode(router *echo.Echo, language, code string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(&SubmissionDTO{
		Language: language,
		Code:     code,
	})

	return submit(router, string(body))
}

func TestSubmitSuccess(t *testing.T) {
	skipWithoutJava(t)

	rec := submitCode(setupRouter(), "java", ValidSolution)

	if rec.Code != http.StatusOK {
		t.Errorf("unexpected status, expected: 200, got: %d", rec.Code)
		return
	}

	var result ResultDTO

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Error(err)
		return
	}

	if result.Run.Output != "one\ntwo\n" {
		t.Errorf("unexpected output, expected: %q, got: %q", "one\ntwo\n", result.Run.Output)
		return
	}
}

func TestSubmitCompileError(t *testing.T) {
	skipWithoutJava(t)

	rec := submitCode(setupRouter(), "java", "not java")

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("unexpected status, expected: 422, got: %d", rec.Code)
		return
	}

	var result ResultDTO

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Error(err)
		return
	}

	if len(result.Build.Errors) == 0 {
		t.Error("expected compiler errors")
		return
	}

	if result.Run != nil {
		t.Error("expected no run result")
		return
	}
}

func TestSubmitRunError(t *testing.T) {
	skipWithoutJava(t)

	rec := submitCode(setupRouter(), "java", ThrowingSolution)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("unexpected status, expected: 422, got: %d", rec.Code)
		return
	}

	var result ResultDTO

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Error(err)
		return
	}

	if !strings.Contains(strings.Join(result.Run.Errors, "\n"), "RuntimeException: boom") {
		t.Errorf("expected exception in run errors, got: %q", result.Run.Errors)
		return
	}
}

func TestSubmitMissingCode(t *testing.T) {
	rec := submit(setupRouter(), `{"language":"java"}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("unexpected status, expected: 400, got: %d", rec.Code)
		return
	}
}

func TestSubmitUnsupportedLanguage(t *testing.T) {
	rec := submitCode(setupRouter(), "cobol", ValidSolution)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("unexpected status, expected: 400, got: %d", rec.Code)
		return
	}
}

func TestSubmitMalformedBody(t *testing.T) {
	rec := submit(setupRouter(), `{`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("unexpected status, expected: 400, got: %d", rec.Code)
		return
	}
}
