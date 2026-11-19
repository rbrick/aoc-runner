package python

import (
	"bytes"
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/rbrick/aoc-runner/internal/runner"
)

const (
	DefaultSolutionCode = `class Solution:
    def part_one(self, reader):
        return ""

    def part_two(self, reader):
        return ""`

	SourceFileName = "code.py"
)

//go:embed python.template
var rawPythonTemplate string

// parse the python template
var PythonTemplate *template.Template = template.Must(template.New("python").Parse(rawPythonTemplate))

var (
	// the line the submitted code starts after in the template
	solutionOffset = runner.LineOffset(rawPythonTemplate, runner.SolutionMarker)

	// matches line numbers in errors that point at the generated source
	linePattern = regexp.MustCompile(regexp.QuoteMeta(SourceFileName) + `", line (\d+)`)
)

type TemplateData struct {
	Solution string
}

type PythonLanguage struct{}

func (*PythonLanguage) DefaultCode() string {
	return DefaultSolutionCode
}

func (*PythonLanguage) Build(ctx context.Context, code string) (*runner.BuildInfo, error) {
	directory, err := os.MkdirTemp("", "aoc-runner-python-")

	if err != nil {
		return nil, err
	}

	path := filepath.Join(directory, SourceFileName)

	if err := writeSource(path, code); err != nil {
		os.RemoveAll(directory)
		return nil, err
	}

	// python isn't compiled, but this catches syntax errors before running
	cmd := exec.CommandContext(ctx, "python3", "-m", "py_compile", path)

	started := time.Now()
	compileErrors, err := runner.CaptureStderr(cmd)

	codeLines := strings.Count(code, "\n") + 1

	build := &runner.BuildInfo{
		ArtifactPath: path,
		CodeLines:    codeLines,
		Duration:     time.Since(started),
		Errors:       runner.MapLines(compileErrors, linePattern, solutionOffset, codeLines),
	}

	if err != nil {
		// nothing to run, clean up the build directory
		os.RemoveAll(directory)
		return build, err
	}

	return build, nil
}

func (*PythonLanguage) Run(ctx context.Context, build *runner.BuildInfo, input string) (*runner.RunInfo, error) {
	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, "python3", build.ArtifactPath)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &stdout

	started := time.Now()
	runErrors, err := runner.CaptureStderr(cmd)

	partOne, partTwo := runner.ParseAnswers(stdout.String())

	run := &runner.RunInfo{
		PartOne:  partOne,
		PartTwo:  partTwo,
		Duration: time.Since(started),
		Errors:   runner.MapLines(runErrors, linePattern, solutionOffset, build.CodeLines),
	}

	if err != nil {
		return run, err
	}

	return run, nil
}

func writeSource(path, code string) error {
	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()

	return PythonTemplate.ExecuteTemplate(file, "python", &TemplateData{
		Solution: code,
	})
}
