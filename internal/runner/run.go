package runner

import (
	"strings"
	"time"
)

const (
	// separates the part one and part two answers in a solution's output
	AnswerSeparator = "\x00"
)

type RunInfo struct {
	PartOne  string
	PartTwo  string
	Duration time.Duration
	Errors   []string
}

// splits a solution's output into the part one and part two answers
func ParseAnswers(output string) (string, string) {
	partOne, partTwo, found := strings.Cut(output, AnswerSeparator)

	if !found {
		// the solution didn't finish, everything is part one
		return output, ""
	}

	return partOne, partTwo
}
