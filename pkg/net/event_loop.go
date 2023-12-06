package net

import (
	"runtime"

	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
)

type EventLoop struct {
	log.Logger
	opts   *Options
	epolls []*Epoll
}

func NewEventLoop(log log.Logger, opts ...Option) (*EventLoop, error) {
	eventLoop := &EventLoop{
		Logger: log,
		opts: &Options{
			loopSize:  runtime.NumCPU(),
			eventSize: 10 * 1024,
			backlog:   1024,
		},
	}
	for _, o := range opts {
		o(eventLoop.opts)
	}
	log.DebugFields("new event loop", zap.Int("loop_size", eventLoop.opts.loopSize), zap.Int("event_size", eventLoop.opts.eventSize), zap.Int("backlog", eventLoop.opts.backlog))

	for i := 0; i < eventLoop.opts.loopSize; i++ {
		epoll, err := NewEpoll(log, eventLoop.opts.eventSize)
		if err != nil {
			return nil, err
		}
		eventLoop.epolls = append(eventLoop.epolls, epoll)
	}
	return eventLoop, nil
}

func (loop *EventLoop) Start(network string, address string, operator Operator) error {
	for _, epoll := range loop.epolls {
		err := epoll.Listen(network, address, loop.opts.backlog, operator)
		if err != nil {
			return err
		}
	}
	return nil
}

func (loop *EventLoop) Wait() {
	for _, epoll := range loop.epolls {
		go func(epoll *Epoll) {
			err := epoll.Wait()
			if err != nil {
				loop.Logger.ErrorFields("event loop wait", zap.Error(err))
			}
		}(epoll)
	}
}

func (loop *EventLoop) Trigger() error {
	for _, epoll := range loop.epolls {
		err := epoll.Trigger()
		if err != nil {
			return err
		}
	}
	return nil
}

func (loop *EventLoop) Close() error {
	for _, epoll := range loop.epolls {
		err := epoll.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
