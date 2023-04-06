// Package main 应用程序
package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/soulnov23/go-tool/internal"
	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/net"
	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

var DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGPIPE, syscall.SIGTERM, syscall.SIGSEGV}
var DefaultHotRestartSIG = []os.Signal{syscall.SIGUSR1}
var DefaultTriggerSIG = []os.Signal{syscall.SIGUSR2}

func main() {
	defer func() {
		if err := recover(); err != nil {
			buffer := make([]byte, 10*1024)
			runtime.Stack(buffer, false)
			log.Info("[PANIC] %v\n%s", err, utils.Byte2String(buffer))
		}
	}()

	appConfig, err := internal.GetAppConfig()
	if err != nil {
		log.Error("get app config: " + err.Error())
		return
	}

	frameLog, err := log.NewZapLog(appConfig.FrameLog)
	if err != nil {
		log.Error("new frame log: " + err.Error())
		return
	}
	frameLog = frameLog.With(zap.String("name", "frame"), zap.String("version", os.Getenv("GO_TOOL_VERSION")))
	defer frameLog.Sync()

	callLog, err := log.NewZapLog(appConfig.CallLog)
	if err != nil {
		log.Error("new call log: " + err.Error())
		return
	}
	defer callLog.Sync()

	runLog, err := log.NewZapLog(appConfig.RunLog)
	if err != nil {
		log.Error("new run log: " + err.Error())
		return
	}
	defer runLog.Sync()

	maxprocs.Set(maxprocs.Logger(frameLog.Debugf))

	frameLog.DebugFields("go-tool start...")
	loopSize := runtime.NumCPU()
	eventLoop, err := net.NewEventLoop(frameLog, net.WithLoopSize(loopSize))
	if err != nil {
		frameLog.ErrorFields("new event loop", zap.Error(err), zap.Int("loop_size", loopSize))
		return
	}
	for _, serverConfig := range appConfig.Server {
		if serverConfig.Protocol == "rpc" {
			err := eventLoop.Start(serverConfig.Network, serverConfig.Address, &internal.RPCServer{FrameLog: frameLog, CallLog: callLog, RunLog: runLog})
			if err != nil {
				frameLog.ErrorFields("event loop start rpc", zap.Error(err))
				return
			}
		} else if serverConfig.Protocol == "http" {
			err := eventLoop.Start(serverConfig.Network, serverConfig.Address, &internal.HTTPServer{FrameLog: frameLog, CallLog: callLog, RunLog: runLog})
			if err != nil {
				frameLog.ErrorFields("event loop start http", zap.Error(err))
				return
			}
		} else {
			frameLog.ErrorFields("protocol not support", zap.String("protocol", serverConfig.Protocol))
			return
		}
	}
	eventLoop.Wait()

	signalClose := make(chan os.Signal, 1)
	signal.Notify(signalClose, DefaultServerCloseSIG...)
	signalHotRestart := make(chan os.Signal, 1)
	signal.Notify(signalHotRestart, DefaultHotRestartSIG...)
	signalTrigger := make(chan os.Signal, 1)
	signal.Notify(signalTrigger, DefaultTriggerSIG...)
	select {
	case sig := <-signalClose:
		frameLog.DebugFields("signal close", zap.String("sig", sig.String()))
		eventLoop.Close()
	case sig := <-signalHotRestart:
		frameLog.DebugFields("signal hot restart", zap.String("sig", sig.String()))
	case sig := <-signalTrigger:
		frameLog.DebugFields("signal trigger", zap.String("sig", sig.String()))
		eventLoop.Trigger()
	}
	frameLog.DebugFields("go-tool closed")
}
