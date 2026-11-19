package golang

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
	DefaultSolutionCode = `import "io"

type Solution struct{}

func (s *Solution) PartOne(reader io.Reader) string {
	return ""
}

func (s *Solution) PartTwo(reader io.Reader) string {
	return ""
}`

	SourceFileName = "solution.go"
	BinaryName     = "solution"

	// the submitted code is built as its own module
	GoMod = "module solution\n\ngo 1.22\n"
)

var (
	//go:embed solution.template
	rawSolutionTemplate string

	//go:embed main.template
	MainSource string
)

// parse the go template
var GoTemplate *template.Template = template.Must(template.New("go").Parse(rawSolutionTemplate))

var (
	// the line the submitted code starts after in the template
	solutionOffset = runner.LineOffset(rawSolutionTemplate, runner.SolutionMarker)

	// matches line numbers in errors that point at the generated source
	linePattern = regexp.MustCompile(regexp.QuoteMeta(SourceFileName) + `:(\d+)`)
)

type TemplateData struct {
	Solution string
}

type GoLanguage struct{}

func (*GoLanguage) Build(ctx context.Context, code string) (*runner.BuildInfo, error) {
	directory, err := os.MkdirTemp("", "aoc-runner-go-")

	if err != nil {
		return nil, err
	}

	if err := writeModule(directory, code); err != nil {
		os.RemoveAll(directory)
		return nil, err
	}

	path := filepath.Join(directory, BinaryName)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", path, ".")
	cmd.Dir = directory
	// don't pick up a go.work from a parent directory or download toolchains
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")

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

func (*GoLanguage) Run(ctx context.Context, build *runner.BuildInfo) (*runner.RunInfo, error) {
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

func writeModule(directory, code string) error {
	if err := os.WriteFile(filepath.Join(directory, "go.mod"), []byte(GoMod), 0o644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(directory, "main.go"), []byte(MainSource), 0o644); err != nil {
		return err
	}

	return writeSource(filepath.Join(directory, SourceFileName), code)
}

func writeSource(path, code string) error {
	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()

	return GoTemplate.ExecuteTemplate(file, "go", &TemplateData{
		Solution: code,
	})
}
