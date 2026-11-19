package runner

import (
	"bufio"
	"io"
	"os/exec"
)

const (
	// the longest line of stderr we'll read
	MaxLineSize = 1024 * 1024
)

type ErrorReader struct {
	source io.Reader
	lines  []string
}

func (r *ErrorReader) Capture() error {
	scanner := bufio.NewScanner(r.source)
	scanner.Buffer(make([]byte, 0, 64*1024), MaxLineSize)

	for scanner.Scan() {
		r.lines = append(r.lines, scanner.Text())
	}

	return scanner.Err()
}

func (r *ErrorReader) Lines() []string {
	return r.lines
}

func CaptureStderr(cmd *exec.Cmd) ([]string, error) {
	stderr, err := cmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	reader := NewErrorReader(stderr)

	if err := reader.Capture(); err != nil {
		// stopped reading early, kill the process so Wait doesn't block on it
		cmd.Process.Kill()
		cmd.Wait()

		return reader.Lines(), err
	}

	// stderr has to be read before calling Wait, Wait closes the pipe
	if err := cmd.Wait(); err != nil {
		return reader.Lines(), &CommandError{
			Command: cmd.String(),
			Err:     err,
		}
	}

	return reader.Lines(), nil
}

func NewErrorReader(source io.Reader) *ErrorReader {
	return &ErrorReader{
		source: source,
	}
}
