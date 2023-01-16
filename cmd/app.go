// Package main 应用程序
package main

import (
	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/net"
)

type App struct {
	appLog log.Logger
}

func (app *App) OnAccept(conn *net.TcpConn) {
	app.appLog.Debug("TODO")
}

func (app *App) OnClose(conn *net.TcpConn) {
	app.appLog.Debug("TODO")
}

func (app *App) OnRead(conn *net.TcpConn) {
	app.appLog.Debug("TODO")
}
