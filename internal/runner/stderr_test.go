package runner

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	_ "embed"
)

var (
	//go:embed testdata/stderr.txt
	StderrOutput string
)

func TestCaptureLines(t *testing.T) {
	reader := NewErrorReader(strings.NewReader(StderrOutput))

	if err := reader.Capture(); err != nil {
		t.Error(err)
		return
	}

	expected := []string{"first", "second", "", "last"}

	if !slices.Equal(reader.Lines(), expected) {
		t.Errorf("unexpected lines, expected: %q, got: %q", expected, reader.Lines())
		return
	}
}

func TestCaptureStderrLongLine(t *testing.T) {
	// a line one byte past MaxLineSize, then a process that would keep running
	script := fmt.Sprintf("head -c %d /dev/zero | tr '\\0' x >&2; sleep 5", MaxLineSize+1)
	cmd := exec.Command("sh", "-c", script)

	started := time.Now()
	_, err := CaptureStderr(cmd)

	if err == nil {
		t.Error("expected an error for a line longer than MaxLineSize")
		return
	}

	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Errorf("expected the process to be killed, took: %s", elapsed)
		return
	}
}

func TestCaptureStderr(t *testing.T) {
	var stdout strings.Builder

	cmd := exec.Command("sh", "-c", "echo out; echo one >&2; echo two >&2; exit 3")
	cmd.Stdout = &stdout

	lines, err := CaptureStderr(cmd)

	var commandErr *CommandError

	if !errors.As(err, &commandErr) {
		t.Errorf("unexpected error, expected: *CommandError, got: %v", err)
		return
	}

	if !slices.Equal(lines, []string{"one", "two"}) {
		t.Errorf("unexpected lines, expected: [one two], got: %q", lines)
		return
	}

	if stdout.String() != "out\n" {
		t.Errorf("unexpected stdout, expected: %q, got: %q", "out\n", stdout.String())
		return
	}
}
