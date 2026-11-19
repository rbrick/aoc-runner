package odin

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
	DefaultSolutionCode = `part_one :: proc(input: string) -> string {
	return ""
}

part_two :: proc(input: string) -> string {
	return ""
}`

	SourceFileName = "code.odin"
	BinaryName     = "solution"
)

//go:embed odin.template
var rawOdinTemplate string

// parse the odin template
var OdinTemplate *template.Template = template.Must(template.New("odin").Parse(rawOdinTemplate))

var (
	// the line the submitted code starts after in the template
	solutionOffset = runner.LineOffset(rawOdinTemplate, runner.SolutionMarker)

	// matches line numbers in errors that point at the generated source
	linePattern = regexp.MustCompile(regexp.QuoteMeta(SourceFileName) + `\((\d+)`)
)

type TemplateData struct {
	Solution string
}

type OdinLanguage struct{}

func (*OdinLanguage) DefaultCode() string {
	return DefaultSolutionCode
}

func (*OdinLanguage) Build(ctx context.Context, code string) (*runner.BuildInfo, error) {
	directory, err := os.MkdirTemp("", "aoc-runner-odin-")

	if err != nil {
		return nil, err
	}

	source := filepath.Join(directory, SourceFileName)

	if err := writeSource(source, code); err != nil {
		os.RemoveAll(directory)
		return nil, err
	}

	path := filepath.Join(directory, BinaryName)

	cmd := exec.CommandContext(ctx, "odin", "build", source, "-file", "-out:"+path)

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

func (*OdinLanguage) Run(ctx context.Context, build *runner.BuildInfo) (*runner.RunInfo, error) {
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

	return OdinTemplate.ExecuteTemplate(file, "odin", &TemplateData{
		Solution: code,
	})
}
