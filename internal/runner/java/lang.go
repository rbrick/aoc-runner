package java

import (
	"bytes"
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
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

type TemplateData struct {
	Solution string
}

type JavaLanguage struct{}

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

	build := &runner.BuildInfo{
		ArtifactPath: path,
		Duration:     time.Since(started),
		Errors:       compileErrors,
	}

	if err != nil {
		// nothing to run, clean up the build directory
		os.RemoveAll(directory)
		return build, err
	}

	return build, nil
}

func (*JavaLanguage) Run(ctx context.Context, artifactPath string) (*runner.RunInfo, error) {
	var stdout bytes.Buffer

	classPath := filepath.Dir(artifactPath)
	className := strings.TrimSuffix(filepath.Base(artifactPath), filepath.Ext(artifactPath))

	cmd := exec.CommandContext(ctx, "java", "-cp", classPath, className)
	cmd.Stdout = &stdout

	started := time.Now()
	runErrors, err := runner.CaptureStderr(cmd)

	run := &runner.RunInfo{
		Output:   stdout.String(),
		Duration: time.Since(started),
		Errors:   runErrors,
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
