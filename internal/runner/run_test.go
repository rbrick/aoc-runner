package runner

import "testing"

func TestParseAnswers(t *testing.T) {
	partOne, partTwo := ParseAnswers("123\x00456")

	if partOne != "123" {
		t.Errorf("unexpected part one, expected: 123, got: %s", partOne)
		return
	}

	if partTwo != "456" {
		t.Errorf("unexpected part two, expected: 456, got: %s", partTwo)
		return
	}
}

func TestParseAnswersMultipleLines(t *testing.T) {
	partOne, partTwo := ParseAnswers("#..#\n####\n#..#\x00\n42")

	if partOne != "#..#\n####\n#..#" {
		t.Errorf("unexpected part one, expected: %q, got: %q", "#..#\n####\n#..#", partOne)
		return
	}

	if partTwo != "\n42" {
		t.Errorf("unexpected part two, expected: %q, got: %q", "\n42", partTwo)
		return
	}
}

func TestParseAnswersMissingSeparator(t *testing.T) {
	partOne, partTwo := ParseAnswers("123")

	if partOne != "123" {
		t.Errorf("unexpected part one, expected: 123, got: %s", partOne)
		return
	}

	if partTwo != "" {
		t.Errorf("unexpected part two, expected empty, got: %s", partTwo)
		return
	}
}
