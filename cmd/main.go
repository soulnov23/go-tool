// Package main 应用程序
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/SoulNov23/go-tool/internal"
	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/net"
	rt "github.com/SoulNov23/go-tool/pkg/runtime"
	"github.com/SoulNov23/go-tool/pkg/unsafe"
)

var DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGPIPE, syscall.SIGTERM, syscall.SIGSEGV}
var DefaultUserCustomSIG = []os.Signal{syscall.SIGUSR1, syscall.SIGUSR2}

func main() {
	defer func() {
		if err := recover(); err != nil {
			buffer := make([]byte, 10*1024)
			runtime.Stack(buffer, false)
			fmt.Printf("[PANIC]%v\n%s\n", err, unsafe.Byte2String(buffer))
		}
	}()
	appConfig, err := internal.GetAppConfig()
	if err != nil {
		panic(rt.GetCaller() + " " + err.Error())
	}
	frameLog, err := log.NewZapLog(appConfig.FrameLog)
	if err != nil {
		panic(rt.GetCaller() + " " + err.Error())
	}
	defer frameLog.Sync()
	callLog, err := log.NewZapLog(appConfig.CallLog)
	if err != nil {
		panic(rt.GetCaller() + " " + err.Error())
	}
	defer callLog.Sync()
	runLog, err := log.NewZapLog(appConfig.RunLog)
	if err != nil {
		panic(rt.GetCaller() + " " + err.Error())
	}
	defer runLog.Sync()

	runLog.Debugf("go-tool start")
	eventLoop, err := net.NewEventLoop(frameLog, net.WithLoopSize(1 /*runtime.NumCPU()*/))
	if err != nil {
		panic(rt.GetCaller() + " " + err.Error())
	}
	for _, serverConfig := range appConfig.Server {
		if serverConfig.Protocol == "rpc" {
			err := eventLoop.Listen(serverConfig.Network, serverConfig.Address, &internal.RPCServer{CallLog: callLog, RunLog: runLog})
			if err != nil {
				panic(rt.GetCaller() + " " + err.Error())
			}
		} else if serverConfig.Protocol == "http" {
			err := eventLoop.Listen(serverConfig.Network, serverConfig.Address, &internal.HTTPServer{CallLog: callLog, RunLog: runLog})
			if err != nil {
				panic(rt.GetCaller() + " " + err.Error())
			}
		} else {
			panic(rt.GetCaller() + " protocol " + serverConfig.Protocol + " not support")
		}
	}
	eventLoop.Wait()

	runLog.Debugf("go-tool version: %s", os.Getenv("SERVER_VERSION"))

	signalClose := make(chan os.Signal, 1)
	signal.Notify(signalClose, DefaultServerCloseSIG...)
	signalUser := make(chan os.Signal, 1)
	signal.Notify(signalUser, DefaultUserCustomSIG...)
	select {
	case sig := <-signalClose:
		runLog.Debugf("signal close: %s", sig.String())
		eventLoop.Close()
	case sig := <-signalUser:
		runLog.Debugf("signal user: %s", sig.String())
		eventLoop.Trigger()
	}
	runLog.Debugf("go-tool stop")
}
