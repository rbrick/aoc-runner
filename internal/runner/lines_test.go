package runner

import (
	"regexp"
	"slices"
	"testing"
)

var (
	testLinePattern = regexp.MustCompile(`code\.java:(\d+)`)
)

func TestLineOffset(t *testing.T) {
	offset := LineOffset("import a;\nimport b;\n\n{{ .Solution }}\n\nvoid main() {}", SolutionMarker)

	if offset != 3 {
		t.Errorf("unexpected offset, expected: 3, got: %d", offset)
		return
	}
}

func TestLineOffsetMissingMarker(t *testing.T) {
	offset := LineOffset("import a;\nimport b;\n", SolutionMarker)

	if offset != 0 {
		t.Errorf("unexpected offset, expected: 0, got: %d", offset)
		return
	}
}

func TestMapLines(t *testing.T) {
	errors := []string{
		"code.java:13: error: ';' expected",
		"\tat code$Solution.partTwo(code.java:18)",
		"\tat code.main(code.java:34)",
		"1 error",
	}

	expected := []string{
		"code.java:3: error: ';' expected",
		"\tat code$Solution.partTwo(code.java:8)",
		"\tat code.main(code.java:34)",
		"1 error",
	}

	mapped := MapLines(errors, testLinePattern, 10, 9)

	if !slices.Equal(mapped, expected) {
		t.Errorf("unexpected mapped lines, expected: %q, got: %q", expected, mapped)
		return
	}
}

func TestMapLinesTemplateHeader(t *testing.T) {
	mapped := MapLines([]string{"code.java:4: error: bad import"}, testLinePattern, 10, 9)

	if mapped[0] != "code.java:4: error: bad import" {
		t.Errorf("unexpected mapped line, expected: code.java:4: error: bad import, got: %s", mapped[0])
		return
	}
}

func TestMapLinesMultipleGroups(t *testing.T) {
	pattern := regexp.MustCompile(`code\.rs:(\d+)|^\s*(\d+) \|`)

	errors := []string{
		" --> /tmp/code.rs:15:9",
		"15 |     let x: i32 = \"s\";",
		"   |              ^^^",
	}

	expected := []string{
		" --> /tmp/code.rs:5:9",
		"5 |     let x: i32 = \"s\";",
		"   |              ^^^",
	}

	mapped := MapLines(errors, pattern, 10, 9)

	if !slices.Equal(mapped, expected) {
		t.Errorf("unexpected mapped lines, expected: %q, got: %q", expected, mapped)
		return
	}
}
