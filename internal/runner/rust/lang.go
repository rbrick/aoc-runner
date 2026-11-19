package rust

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
	DefaultSolutionCode = `struct Solution;

impl Solution {
    fn part_one(&self, input: &str) -> String {
        String::new()
    }

    fn part_two(&self, input: &str) -> String {
        String::new()
    }
}`

	SourceFileName = "code.rs"
	BinaryName     = "solution"
)

//go:embed rust.template
var rawRustTemplate string

// parse the rust template
var RustTemplate *template.Template = template.Must(template.New("rust").Parse(rawRustTemplate))

var (
	// the line the submitted code starts after in the template
	solutionOffset = runner.LineOffset(rawRustTemplate, runner.SolutionMarker)

	// matches line numbers in errors that point at the generated source
	linePattern = regexp.MustCompile(regexp.QuoteMeta(SourceFileName) + `:(\d+)|^\s*(\d+) \|`)
)

type TemplateData struct {
	Solution string
}

type RustLanguage struct{}

func (*RustLanguage) DefaultCode() string {
	return DefaultSolutionCode
}

func (*RustLanguage) Build(ctx context.Context, code string) (*runner.BuildInfo, error) {
	directory, err := os.MkdirTemp("", "aoc-runner-rust-")

	if err != nil {
		return nil, err
	}

	source := filepath.Join(directory, SourceFileName)

	if err := writeSource(source, code); err != nil {
		os.RemoveAll(directory)
		return nil, err
	}

	path := filepath.Join(directory, BinaryName)

	// warnings aren't errors, keep them out of the build output
	cmd := exec.CommandContext(ctx, "rustc", "-O", "-A", "warnings", "--edition", "2021", "-o", path, source)

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

func (*RustLanguage) Run(ctx context.Context, build *runner.BuildInfo) (*runner.RunInfo, error) {
	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, build.ArtifactPath)
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

	return RustTemplate.ExecuteTemplate(file, "rust", &TemplateData{
		Solution: code,
	})
}
