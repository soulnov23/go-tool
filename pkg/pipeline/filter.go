package pipeline

import (
	"context"
)

type Filter interface {
	Name() string
	Exec(ctx context.Context, source any, sink any) error
}
