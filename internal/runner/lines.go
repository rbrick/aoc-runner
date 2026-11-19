package runner

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	// where the submitted code is placed in a language template
	SolutionMarker = "{{ .Solution }}"
)

// the number of template lines that come before the submitted code
func LineOffset(template, marker string) int {
	index := strings.Index(template, marker)

	if index == -1 {
		return 0
	}

	return strings.Count(template[:index], "\n")
}

// shifts line numbers captured by pattern so they point at the submitted code
// instead of the generated source. numbers outside the submitted code are left alone.
func MapLines(errors []string, pattern *regexp.Regexp, offset, codeLines int) []string {
	var mapped []string

	for _, line := range errors {
		mapped = append(mapped, mapLine(line, pattern, offset, codeLines))
	}

	return mapped
}

func mapLine(line string, pattern *regexp.Regexp, offset, codeLines int) string {
	var builder strings.Builder

	last := 0

	for _, match := range pattern.FindAllStringSubmatchIndex(line, -1) {
		start, end := matchedGroup(match)

		if start == -1 {
			continue
		}

		number, err := strconv.Atoi(line[start:end])

		if err != nil {
			continue
		}

		number -= offset

		if number < 1 || number > codeLines {
			// part of the template, not the submitted code
			continue
		}

		builder.WriteString(line[last:start])
		builder.WriteString(strconv.Itoa(number))

		last = end
	}

	builder.WriteString(line[last:])

	return builder.String()
}

// the first capture group that took part in the match, patterns can have
// more than one when a line number shows up in different formats
func matchedGroup(match []int) (int, int) {
	for group := 2; group < len(match); group += 2 {
		if match[group] != -1 {
			return match[group], match[group+1]
		}
	}

	return -1, -1
}
