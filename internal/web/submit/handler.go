package submit

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/runner"
	"github.com/rbrick/aoc-runner/internal/runner/golang"
	"github.com/rbrick/aoc-runner/internal/runner/java"
	"github.com/rbrick/aoc-runner/internal/runner/odin"
	"github.com/rbrick/aoc-runner/internal/runner/python"
	"github.com/rbrick/aoc-runner/internal/runner/rust"
	"github.com/rbrick/aoc-runner/internal/runner/typescript"
)

const (
	// how long a solution can run before it's killed
	RunTimeout = 10 * time.Second
)

type Handler struct {
	languages map[string]runner.Language
}

func (h *Handler) submit(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()

	var submission SubmissionDTO

	if err := reqCtx.Bind(&submission); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if submission.Code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "code is required")
	}

	language, ok := h.languages[submission.Language]

	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported language: "+submission.Language)
	}

	build, err := language.Build(ctx, submission.Code)

	if build == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err != nil {
		// failed to compile, send back the compiler errors
		return reqCtx.JSON(http.StatusUnprocessableEntity, &ResultDTO{
			Build: mapBuild(build),
		})
	}

	defer os.RemoveAll(filepath.Dir(build.ArtifactPath))

	runCtx, cancel := context.WithTimeout(ctx, RunTimeout)
	defer cancel()

	run, err := language.Run(runCtx, build)

	if run == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	result := &ResultDTO{
		Build: mapBuild(build),
		Run:   mapRun(run),
	}

	if err != nil {
		// failed while running, send back what was written to stderr
		return reqCtx.JSON(http.StatusUnprocessableEntity, result)
	}

	return reqCtx.JSON(http.StatusOK, result)
}

func NewHandler() *Handler {
	return &Handler{
		languages: map[string]runner.Language{
			"java":       &java.JavaLanguage{},
			"python":     &python.PythonLanguage{},
			"go":         &golang.GoLanguage{},
			"typescript": &typescript.TypeScriptLanguage{},
			"rust":       &rust.RustLanguage{},
			"odin":       &odin.OdinLanguage{},
		},
	}
}
