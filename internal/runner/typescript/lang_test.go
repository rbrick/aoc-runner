package typescript

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	_ "embed"
)

var (
	//go:embed testdata/valid.ts
	ValidSolution string

	//go:embed testdata/invalid.ts
	InvalidSolution string

	//go:embed testdata/throws.ts
	ThrowingSolution string
)

func skipWithoutToolchain(t *testing.T) {
	if _, err := exec.LookPath("tsc"); err != nil {
		t.Skip("tsc is not installed")
	}

	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
}

func TestBuildAndRun(t *testing.T) {
	skipWithoutToolchain(t)

	bgCtx := context.Background()
	language := &TypeScriptLanguage{}

	build, err := language.Build(bgCtx, ValidSolution)

	if err != nil {
		t.Error(err)
		return
	}

	defer os.RemoveAll(filepath.Dir(build.ArtifactPath))

	run, err := language.Run(bgCtx, build)

	if err != nil {
		t.Error(err)
		return
	}

	if run.PartOne != "one" {
		t.Errorf("unexpected part one, expected: one, got: %s", run.PartOne)
		return
	}

	if run.PartTwo != "two" {
		t.Errorf("unexpected part two, expected: two, got: %s", run.PartTwo)
		return
	}
}

func TestBuildErrors(t *testing.T) {
	skipWithoutToolchain(t)

	bgCtx := context.Background()
	language := &TypeScriptLanguage{}

	build, err := language.Build(bgCtx, InvalidSolution)

	if err == nil {
		t.Error("expected compilation to fail")
		return
	}

	if !slices.ContainsFunc(build.Errors, func(line string) bool {
		return strings.Contains(line, `code.ts(7,`)
	}) {
		t.Errorf("expected error on line of submitted code, got: %q", build.Errors)
		return
	}

	if len(build.Errors) == 0 {
		t.Error("expected compiler errors")
		return
	}

	if _, err := os.Stat(filepath.Dir(build.ArtifactPath)); !os.IsNotExist(err) {
		t.Errorf("expected build directory to be removed, got: %v", err)
		return
	}
}

func TestRunErrors(t *testing.T) {
	skipWithoutToolchain(t)

	bgCtx := context.Background()
	language := &TypeScriptLanguage{}

	build, err := language.Build(bgCtx, ThrowingSolution)

	if err != nil {
		t.Error(err)
		return
	}

	defer os.RemoveAll(filepath.Dir(build.ArtifactPath))

	run, err := language.Run(bgCtx, build)

	if err == nil {
		t.Error("expected the run to fail")
		return
	}

	if len(run.Errors) < 2 {
		t.Errorf("unexpected errors length, expected at least: 2, got: %d", len(run.Errors))
		return
	}

	if run.Errors[0] != "debugging" {
		t.Errorf("unexpected first line, expected: debugging, got: %s", run.Errors[0])
		return
	}

	if !slices.ContainsFunc(run.Errors, func(line string) bool {
		return strings.Contains(line, "Error: boom")
	}) {
		t.Errorf("expected exception in errors, got: %q", run.Errors)
		return
	}

	if !slices.ContainsFunc(run.Errors, func(line string) bool {
		return strings.Contains(line, `code.ts:8`)
	}) {
		t.Errorf("expected stack trace on line of submitted code, got: %q", run.Errors)
		return
	}
}
