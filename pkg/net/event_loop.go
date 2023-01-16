package net

import (
	"errors"
	"runtime"

	"github.com/SoulNov23/go-tool/pkg/log"
)

type EventLoop struct {
	log    log.Logger
	opts   *Options
	epolls []*Epoll
}

func NewEventLoop(log log.Logger, server Server, opts ...Option) (*EventLoop, error) {
	eventLoop := &EventLoop{
		log: log,
		opts: &Options{
			loopSize:  runtime.NumCPU(),
			eventSize: 10 * 1024,
			backlog:   1024,
		},
	}
	for _, o := range opts {
		o(eventLoop.opts)
	}
	log.Debugf("EventLoop loopSize: %d, eventSize: %d, backlog: %d", eventLoop.opts.loopSize, eventLoop.opts.eventSize, eventLoop.opts.backlog)

	for i := 0; i < eventLoop.opts.loopSize; i++ {
		epoll, err := NewEpoll(log, eventLoop.opts.eventSize, server)
		if err != nil {
			wrapErr := errors.New("net.NewEpoll: " + err.Error())
			log.Error(wrapErr)
			return nil, wrapErr
		}
		eventLoop.epolls = append(eventLoop.epolls, epoll)
	}
	return eventLoop, nil
}

func (loop *EventLoop) Listen(network string, address string) error {
	for _, epoll := range loop.epolls {
		err := epoll.Listen(network, address, loop.opts.backlog)
		if err != nil {
			wrapErr := errors.New("epoll.Listen: " + err.Error())
			loop.log.Error(wrapErr)
			return wrapErr
		}
	}
	return nil
}

func (loop *EventLoop) Wait() {
	for _, epoll := range loop.epolls {
		go func(epoll *Epoll) {
			err := epoll.Wait()
			if err != nil {
				wrapErr := errors.New("epoll.Wait: " + err.Error())
				loop.log.Error(wrapErr)
			}
		}(epoll)
	}
}

func (loop *EventLoop) Trigger() error {
	for _, epoll := range loop.epolls {
		err := epoll.Trigger()
		if err != nil {
			wrapErr := errors.New("epoll.Trigger: " + err.Error())
			loop.log.Error(wrapErr)
			return wrapErr
		}
	}
	return nil
}

func (loop *EventLoop) Close() {
	for _, epoll := range loop.epolls {
		epoll.Close()
	}
}
