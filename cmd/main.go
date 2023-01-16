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
	appLog, err := log.NewZapLog(appConfig.AppLog)
	if err != nil {
		panic(rt.GetCaller() + " " + err.Error())
	}
	defer appLog.Sync()

	eventLoop, err := net.NewEventLoop(frameLog, &App{appLog: appLog}, net.WithLoopSize(runtime.NumCPU()))
	if err != nil {
		panic(rt.GetCaller() + "\t" + err.Error())
	}
	for _, serverConfig := range appConfig.Server {
		err := eventLoop.Listen(serverConfig.Network, serverConfig.Ip+":"+serverConfig.Port)
		if err != nil {
			panic(rt.GetCaller() + "\t" + err.Error())
		}
	}
	eventLoop.Wait()

	appLog.Debugf("go-tool version: %s", os.Getenv("SERVER_VERSION"))

	signalClose := make(chan os.Signal, 1)
	signal.Notify(signalClose, DefaultServerCloseSIG...)
	signalUser := make(chan os.Signal, 1)
	signal.Notify(signalUser, DefaultUserCustomSIG...)
	select {
	case sig := <-signalClose:
		appLog.Debugf("signal: %s", sig.String())
		eventLoop.Close()
	case sig := <-signalUser:
		appLog.Debugf("signal: %s", sig.String())
		eventLoop.Trigger()
	}
}
