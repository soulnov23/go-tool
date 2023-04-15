// Package main 应用程序
package main

import (
	"flag"
	"fmt"
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

var (
	goVersion       string
	gitBranch       string
	gitCommitID     string
	gitCommitTime   string
	gitCommitAuthor string

	DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGPIPE, syscall.SIGTERM, syscall.SIGSEGV}
	DefaultHotRestartSIG  = []os.Signal{syscall.SIGUSR1}
	DefaultTriggerSIG     = []os.Signal{syscall.SIGUSR2}
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			buffer := make([]byte, 10*1024)
			runtime.Stack(buffer, false)
			fmt.Printf("[PANIC] %v\n%s\n", err, utils.Byte2String(buffer))
		}
	}()

	// 定义需要解析的命令行参数
	var version bool
	var path string
	flag.BoolVar(&version, "version", false, "show server version")
	flag.StringVar(&path, "conf", "./go_tool.yaml", "server config file path")
	// 开始解析命令行
	flag.Parse()
	// 命令行参数都不匹配，打印help
	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}
	if version {
		fmt.Printf("go version: %s\n", goVersion)
		fmt.Printf("git branch: %s\n", gitBranch)
		fmt.Printf("git commit id: %s\n", gitCommitID)
		fmt.Printf("git commit time: %s\n", gitCommitTime)
		fmt.Printf("git commit author: %s\n", gitCommitAuthor)
		return
	}

	appConfig, err := internal.GetAppConfig(path)
	if err != nil {
		fmt.Printf("get app config: %s\n" + err.Error())
		return
	}

	frameLog, err := log.NewZapLog(appConfig.FrameLog)
	if err != nil {
		fmt.Printf("new frame log: %s\n" + err.Error())
		return
	}
	frameLog = frameLog.With(zap.String("name", "frame"))
	defer frameLog.Sync()

	runLog, err := log.NewZapLog(appConfig.RunLog)
	if err != nil {
		fmt.Printf("new run log: %s\n" + err.Error())
		return
	}
	runLog = runLog.With(zap.String("name", "run"))
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
			err := eventLoop.Start(serverConfig.Network, serverConfig.Address, &internal.RPCServer{FrameLog: frameLog, RunLog: runLog})
			if err != nil {
				frameLog.ErrorFields("event loop start rpc", zap.Error(err))
				return
			}
		} else if serverConfig.Protocol == "http" {
			err := eventLoop.Start(serverConfig.Network, serverConfig.Address, &internal.HTTPServer{FrameLog: frameLog, RunLog: runLog})
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
