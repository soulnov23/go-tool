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

	log.Debug("internal.GetAppConfig begin...")
	appConfig, err := internal.GetAppConfig()
	if err != nil {
		log.Error("internal.GetAppConfig: " + err.Error())
		return
	}
	log.Debug("internal.GetAppConfig success")

	log.Debug("log.NewZapLog frame log begin...")
	frameLog, err := log.NewZapLog(appConfig.FrameLog)
	if err != nil {
		log.Error("log.NewZapLog: " + err.Error())
		return
	}
	defer frameLog.Sync()
	log.Debug("log.NewZapLog frame log success")

	log.Debug("log.NewZapLog call log begin...")
	sugaredCallLog, err := log.NewZapLog(appConfig.CallLog)
	if err != nil {
		log.Error("log.NewZapLog: " + err.Error())
		return
	}
	callLog := sugaredCallLog.Desugar()
	defer callLog.Sync()
	log.Debug("log.NewZapLog call log success")

	log.Debug("log.NewZapLog run log begin...")
	runLog, err := log.NewZapLog(appConfig.RunLog)
	if err != nil {
		log.Error("log.NewZapLog: " + err.Error())
		return
	}
	defer runLog.Sync()
	log.Debug("log.NewZapLog run log success")

	maxprocs.Set(maxprocs.Logger(log.Debug))

	frameLog.Debugf("go-tool start")
	eventLoop, err := net.NewEventLoop(frameLog, net.WithLoopSize(runtime.NumCPU()))
	if err != nil {
		log.Error("net.NewEventLoop: " + err.Error())
		return
	}
	for _, serverConfig := range appConfig.Server {
		if serverConfig.Protocol == "rpc" {
			err := eventLoop.Listen(serverConfig.Network, serverConfig.Address, &internal.RPCServer{CallLog: callLog, RunLog: runLog})
			if err != nil {
				log.Error("eventLoop.Listen: " + err.Error())
				return
			}
		} else if serverConfig.Protocol == "http" {
			err := eventLoop.Listen(serverConfig.Network, serverConfig.Address, &internal.HTTPServer{CallLog: callLog, RunLog: runLog})
			if err != nil {
				log.Error("eventLoop.Listen: " + err.Error())
				return
			}
		} else {
			log.Error("protocol " + serverConfig.Protocol + " not support")
			return
		}
	}
	eventLoop.Wait()

	frameLog.Debugf("go-tool version: %s", os.Getenv("SERVER_VERSION"))

	signalClose := make(chan os.Signal, 1)
	signal.Notify(signalClose, DefaultServerCloseSIG...)
	signalHotRestart := make(chan os.Signal, 1)
	signal.Notify(signalHotRestart, DefaultHotRestartSIG...)
	signalTrigger := make(chan os.Signal, 1)
	signal.Notify(signalTrigger, DefaultTriggerSIG...)
	select {
	case sig := <-signalClose:
		frameLog.Debugf("signal close: %s", sig.String())
		eventLoop.Close()
	case sig := <-signalHotRestart:
		frameLog.Debugf("signal hot restart: %s", sig.String())
		// TODO
	case sig := <-signalTrigger:
		frameLog.Debugf("signal trigger: %s", sig.String())
		eventLoop.Trigger()
	}
	frameLog.Debugf("go-tool stop")
}
