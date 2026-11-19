package runner

import "time"

type RunInfo struct {
	Output   string
	Duration time.Duration
	Errors   []string
}
