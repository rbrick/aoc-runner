package haskell

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
	DefaultSolutionCode = `partOne :: String -> String
partOne input = ""

partTwo :: String -> String
partTwo input = ""`

	SourceFileName = "code.hs"
	BinaryName     = "solution"
)

//go:embed haskell.template
var rawHaskellTemplate string

// parse the haskell template
var HaskellTemplate *template.Template = template.Must(template.New("haskell").Parse(rawHaskellTemplate))

var (
	// the line the submitted code starts after in the template
	solutionOffset = runner.LineOffset(rawHaskellTemplate, runner.SolutionMarker)

	// matches line numbers in errors that point at the generated source,
	// ghc also prints the source line with a gutter, like "5 |"
	linePattern = regexp.MustCompile(regexp.QuoteMeta(SourceFileName) + `:(\d+)|^\s*(\d+) \|`)
)

type TemplateData struct {
	Solution string
}

type HaskellLanguage struct{}

func (*HaskellLanguage) DefaultCode() string {
	return DefaultSolutionCode
}

func (*HaskellLanguage) Build(ctx context.Context, code string) (*runner.BuildInfo, error) {
	directory, err := os.MkdirTemp("", "aoc-runner-haskell-")

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
	cmd := exec.CommandContext(ctx, "ghc", "-O1", "-w", "-outputdir", filepath.Join(directory, "build"), "-o", path, source)

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

func (*HaskellLanguage) Run(ctx context.Context, build *runner.BuildInfo, input string) (*runner.RunInfo, error) {
	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, build.ArtifactPath)
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

	return HaskellTemplate.ExecuteTemplate(file, "haskell", &TemplateData{
		Solution: code,
	})
}
