package submit

import "github.com/rbrick/aoc-runner/internal/runner"

type SubmissionDTO struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type BuildDTO struct {
	DurationMS int64    `json:"duration_ms"`
	Errors     []string `json:"errors"`
}

type RunDTO struct {
	DurationMS int64    `json:"duration_ms"`
	PartOne    string   `json:"part_one"`
	PartTwo    string   `json:"part_two"`
	Errors     []string `json:"errors"`
}

type ResultDTO struct {
	Build *BuildDTO `json:"build"`
	Run   *RunDTO   `json:"run,omitempty"`
}

func mapBuild(build *runner.BuildInfo) *BuildDTO {
	return &BuildDTO{
		DurationMS: build.Duration.Milliseconds(),
		Errors:     build.Errors,
	}
}

func mapRun(run *runner.RunInfo) *RunDTO {
	return &RunDTO{
		DurationMS: run.Duration.Milliseconds(),
		PartOne:    run.PartOne,
		PartTwo:    run.PartTwo,
		Errors:     run.Errors,
	}
}
