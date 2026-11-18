package runner

import "context"

type Language interface {
	Build(ctx context.Context, code string) (string, error)
	Run(ctx context.Context, code string) (string, error)
}
