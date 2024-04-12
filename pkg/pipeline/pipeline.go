package pipeline

import (
	"context"
	"fmt"
	"time"
)

type ProcessType string

const (
	Serial   ProcessType = "Serial"
	Parallel ProcessType = "Parallel"
)

type InitFunc func() map[string]any

type Pipeline struct {
	ctx    context.Context
	init   InitFunc
	stages []*stage
	opts   *Options
}

func New(ctx context.Context, init InitFunc, opts ...Option) *Pipeline {
	defaultOpts := &Options{
		timeout: 200 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(defaultOpts)
	}
	return &Pipeline{
		ctx:  ctx,
		init: init,
	}
}

func (p *Pipeline) Add(s *stage) *Pipeline {
	p.stages = append(p.stages, s)
	return p
}

func (p *Pipeline) Process(t ProcessType) error {
	if t == Serial {
	} else if t == Parallel {
	} else {
		return fmt.Errorf("not support process type: %s", string(t))
	}
	for key, value := range p.init() {
		p.ctx = context.WithValue(p.ctx, key, value)
	}
	return nil
}
