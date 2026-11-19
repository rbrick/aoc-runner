package runner

import "time"

type BuildInfo struct {
	ArtifactPath string
	CodeLines    int
	Duration     time.Duration
	Errors       []string
}
