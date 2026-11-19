package runner

import "context"

type Language interface {
	Build(ctx context.Context, code string) (*BuildInfo, error)
	Run(ctx context.Context, build *BuildInfo) (*RunInfo, error)
}
