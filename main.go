// Package main 应用程序
package main

import (
	"flag"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/soulnov23/go-tool/internal"
	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/net"
	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/zap"
)

var (
	goVersion       string
	gitBranch       string
	gitCommitID     string
	gitCommitTime   string
	gitCommitAuthor string
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("[PANIC] %v\n%s\n", err, utils.BytesToString(debug.Stack()))
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

	runLog, err := log.New(appConfig.RunLog)
	if err != nil {
		fmt.Printf("new run log: %s\n" + err.Error())
		return
	}
	runLog = runLog.With(zap.String("name", "run"))
	defer runLog.Sync()

	frameLog.DebugFields("go-tool start...")

	loopSize := runtime.NumCPU()
	eventLoop, err := net.NewEventLoop(frameLog, net.WithLoopSize(loopSize))
	if err != nil {
		frameLog.FatalFields("new event loop", zap.Error(err), zap.Int("loop_size", loopSize))
		return
	}
	for _, serviceConfig := range appConfig.Server.Services {
		if serviceConfig.Protocol == "rpc" {
			err := eventLoop.Start(serviceConfig.Network, serviceConfig.Address, &internal.RPCServer{FrameLog: frameLog, RunLog: runLog})
			if err != nil {
				frameLog.FatalFields("event loop start rpc", zap.Error(err))
				return
			}
		} else if serviceConfig.Protocol == "http" {
			err := eventLoop.Start(serviceConfig.Network, serviceConfig.Address, &internal.HTTPServer{FrameLog: frameLog, RunLog: runLog})
			if err != nil {
				frameLog.FatalFields("event loop start http", zap.Error(err))
				return
			}
		} else {
			frameLog.FatalFields("protocol not support", zap.String("protocol", serviceConfig.Protocol))
			return
		}
	}
	eventLoop.Wait()

	frameLog.DebugFields("go-tool closed")
}
