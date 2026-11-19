package runner

import "context"

type Language interface {
	DefaultCode() string
	Build(ctx context.Context, code string) (*BuildInfo, error)
	Run(ctx context.Context, build *BuildInfo) (*RunInfo, error)
}
