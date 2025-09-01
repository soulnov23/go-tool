// Package main 应用程序
package main

import (
	"flag"
	"fmt"
	"runtime/debug"

	"github.com/soulnov23/go-tool/pkg/framework"
	"github.com/soulnov23/go-tool/pkg/utils"
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
	framework.New(path).Serve()
}
