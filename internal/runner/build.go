package runner

import "time"

type BuildInfo struct {
	ArtifactPath string
	Duration     time.Duration
	Errors       []string
}
