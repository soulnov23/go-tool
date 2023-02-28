// Package main 应用程序
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/soulnov23/go-tool/internal"
	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/net"
	"github.com/soulnov23/go-tool/pkg/utils"
)

var DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGPIPE, syscall.SIGTERM, syscall.SIGSEGV}
var DefaultUserCustomSIG = []os.Signal{syscall.SIGUSR1, syscall.SIGUSR2}

func main() {
	defer func() {
		if err := recover(); err != nil {
			buffer := make([]byte, 10*1024)
			runtime.Stack(buffer, false)
			fmt.Printf("[PANIC] %v\n%s\n", err, utils.Byte2String(buffer))
		}
	}()
	appConfig, err := internal.GetAppConfig()
	if err != nil {
		fmt.Print("[ERROR] " + utils.GetCaller(1) + " " + err.Error())
		return
	}
	frameLog, err := log.NewZapLog(appConfig.FrameLog)
	if err != nil {
		fmt.Print("[ERROR] " + utils.GetCaller(1) + " " + err.Error())
		return
	}
	defer frameLog.Sync()
	sugaredCallLog, err := log.NewZapLog(appConfig.CallLog)
	if err != nil {
		fmt.Print("[ERROR] " + utils.GetCaller(1) + " " + err.Error())
		return
	}
	callLog := sugaredCallLog.Desugar()
	defer callLog.Sync()
	runLog, err := log.NewZapLog(appConfig.RunLog)
	if err != nil {
		fmt.Print("[ERROR] " + utils.GetCaller(1) + " " + err.Error())
		return
	}
	defer runLog.Sync()

	frameLog.Debugf("go-tool start")
	eventLoop, err := net.NewEventLoop(frameLog, net.WithLoopSize(runtime.NumCPU()))
	if err != nil {
		fmt.Print("[ERROR] " + utils.GetCaller(1) + " " + err.Error())
		return
	}
	for _, serverConfig := range appConfig.Server {
		if serverConfig.Protocol == "rpc" {
			err := eventLoop.Listen(serverConfig.Network, serverConfig.Address, &internal.RPCServer{CallLog: callLog, RunLog: runLog})
			if err != nil {
				fmt.Print("[ERROR] " + utils.GetCaller(1) + " " + err.Error())
				return
			}
		} else if serverConfig.Protocol == "http" {
			err := eventLoop.Listen(serverConfig.Network, serverConfig.Address, &internal.HTTPServer{CallLog: callLog, RunLog: runLog})
			if err != nil {
				fmt.Print("[ERROR] " + utils.GetCaller(1) + " " + err.Error())
				return
			}
		} else {
			fmt.Print("[ERROR] " + utils.GetCaller(1) + " protocol " + serverConfig.Protocol + " not support")
		}
	}
	eventLoop.Wait()

	frameLog.Debugf("go-tool version: %s", os.Getenv("SERVER_VERSION"))

	signalClose := make(chan os.Signal, 1)
	signal.Notify(signalClose, DefaultServerCloseSIG...)
	signalUser := make(chan os.Signal, 1)
	signal.Notify(signalUser, DefaultUserCustomSIG...)
	select {
	case sig := <-signalClose:
		frameLog.Debugf("signal close: %s", sig.String())
		eventLoop.Close()
	case sig := <-signalUser:
		frameLog.Debugf("signal user: %s", sig.String())
		eventLoop.Trigger()
	}
	frameLog.Debugf("go-tool stop")
}
