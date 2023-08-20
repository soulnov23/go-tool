// Package main 应用程序
package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/soulnov23/go-tool/internal"
	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/net"
	convert "github.com/soulnov23/go-tool/pkg/strconv"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

var (
	goVersion       string
	gitBranch       string
	gitCommitID     string
	gitCommitTime   string
	gitCommitAuthor string

	DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV}
	DefaultHotRestartSIG  = []os.Signal{syscall.SIGUSR1}
	DefaultTriggerSIG     = []os.Signal{syscall.SIGUSR2}
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("[PANIC] %v\n%s\n", err, convert.BytesToString(debug.Stack()))
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

	frameLog, err := log.New(appConfig.FrameLog)
	if err != nil {
		fmt.Printf("new frame log: %s\n" + err.Error())
		return
	}
	frameLog = frameLog.With(zap.String("name", "frame"))
	defer frameLog.Sync()

	runLog, err := log.New(appConfig.RunLog)
	if err != nil {
		fmt.Printf("new run log: %s\n" + err.Error())
		return
	}
	runLog = runLog.With(zap.String("name", "run"))
	defer runLog.Sync()

	maxprocs.Set(maxprocs.Logger(frameLog.Debugf))

	frameLog.DebugFields("go-tool start...")

	/*
		创建mux自定义处理函数，避免与pprof的默认http.DefaultServeMux冲突
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
		http.ListenAndServe("ip:port", mux)
	*/
	addr := "127.0.0.1:9999"
	readTimeout := 0
	writeTimeout := 0
	idleTimeout := 0
	if appConfig.Server.Debug.Address != "" {
		addr = appConfig.Server.Debug.Address
	}
	if appConfig.Server.Debug.ReadTimeout > 0 {
		readTimeout = appConfig.Server.Debug.ReadTimeout
	}
	if appConfig.Server.Debug.WriteTimeout > 0 {
		writeTimeout = appConfig.Server.Debug.WriteTimeout
	}
	if appConfig.Server.Debug.IdleTimeout > 0 {
		idleTimeout = appConfig.Server.Debug.IdleTimeout
	}
	debugServer := &http.Server{
		Addr:         addr,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  time.Duration(readTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(writeTimeout) * time.Millisecond,
		IdleTimeout:  time.Duration(idleTimeout) * time.Millisecond,
	}
	go func() {
		if err := debugServer.ListenAndServe(); err != nil {
			frameLog.FatalFields("new debug", zap.Reflect("debug_server", debugServer), zap.Error(err))
		}
	}()

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
