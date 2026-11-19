package java

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
	DefaultSolutionCode = `class Solution {
   public String partOne(Reader reader) {
       return "";
   }

   public String partTwo(Reader reader) {
       return "";
   }
}`

	SourceFileName = "code.java"
)

//go:embed java.template
var rawJavaTemplate string

// parse the java template
var JavaTemplate *template.Template = template.Must(template.New("java").Parse(rawJavaTemplate))

var (
	// the line the submitted code starts after in the template
	solutionOffset = runner.LineOffset(rawJavaTemplate, runner.SolutionMarker)

	// matches line numbers in errors that point at the generated source
	linePattern = regexp.MustCompile(regexp.QuoteMeta(SourceFileName) + `:(\d+)`)
)

type TemplateData struct {
	Solution string
}

type JavaLanguage struct{}

func (*JavaLanguage) DefaultCode() string {
	return DefaultSolutionCode
}

func (*JavaLanguage) Build(ctx context.Context, code string) (*runner.BuildInfo, error) {
	directory, err := os.MkdirTemp("", "aoc-runner-java-")

	if err != nil {
		return nil, err
	}

	path := filepath.Join(directory, SourceFileName)

	if err := writeSource(path, code); err != nil {
		os.RemoveAll(directory)
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "javac", path)

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

func (*JavaLanguage) Run(ctx context.Context, build *runner.BuildInfo, input string) (*runner.RunInfo, error) {
	var stdout bytes.Buffer

	classPath := filepath.Dir(build.ArtifactPath)
	className := strings.TrimSuffix(filepath.Base(build.ArtifactPath), filepath.Ext(build.ArtifactPath))

	cmd := exec.CommandContext(ctx, "java", "-cp", classPath, className)
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

	return JavaTemplate.ExecuteTemplate(file, "java", &TemplateData{
		Solution: code,
	})
}
