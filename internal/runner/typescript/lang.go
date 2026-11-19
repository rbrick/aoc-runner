package typescript

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
    partOne(input: string): string {
        return "";
    }

    partTwo(input: string): string {
        return "";
    }
}`

	SourceFileName = "code.ts"
	OutputName     = "code.js"
)

//go:embed typescript.template
var rawTypeScriptTemplate string

// parse the typescript template
var TypeScriptTemplate *template.Template = template.Must(template.New("typescript").Parse(rawTypeScriptTemplate))

var (
	// the line the submitted code starts after in the template
	solutionOffset = runner.LineOffset(rawTypeScriptTemplate, runner.SolutionMarker)

	// matches line numbers in errors that point at the generated source
	linePattern = regexp.MustCompile(regexp.QuoteMeta(SourceFileName) + `[(:](\d+)`)
)

type TemplateData struct {
	Solution string
}

type TypeScriptLanguage struct{}

func (*TypeScriptLanguage) DefaultCode() string {
	return DefaultSolutionCode
}

func (*TypeScriptLanguage) Build(ctx context.Context, code string) (*runner.BuildInfo, error) {
	directory, err := os.MkdirTemp("", "aoc-runner-typescript-")

	if err != nil {
		return nil, err
	}

	source := filepath.Join(directory, SourceFileName)

	if err := writeSource(source, code); err != nil {
		os.RemoveAll(directory)
		return nil, err
	}

	path := filepath.Join(directory, OutputName)

	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, "tsc",
		"--strict",
		"--target", "es2022",
		"--module", "commonjs",
		"--sourceMap",
		"--noEmitOnError",
		"--pretty", "false",
		SourceFileName,
	)
	// run from the build directory so errors use a short path
	cmd.Dir = directory
	cmd.Stdout = &stdout

	started := time.Now()
	compileErrors, err := runner.CaptureStderr(cmd)

	// tsc reports errors on stdout instead of stderr
	if output := strings.TrimSpace(stdout.String()); output != "" {
		compileErrors = append(compileErrors, strings.Split(output, "\n")...)
	}

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

func (*TypeScriptLanguage) Run(ctx context.Context, build *runner.BuildInfo, input string) (*runner.RunInfo, error) {
	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, "node", "--enable-source-maps", build.ArtifactPath)
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

	return TypeScriptTemplate.ExecuteTemplate(file, "typescript", &TemplateData{
		Solution: code,
	})
}
