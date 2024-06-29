package main

import (
	"flag"
	"fmt"
	"runtime/debug"

	"github.com/soulnov23/go-tool/pkg/utils"
)

var (
	source      = flag.String("source", "", "Specify a list of yaml files. eg:-source=./a.yaml,./b.yaml")
	destination = flag.String("destination", "", "Specify a golang file. eg:-destination=./errors.go")
	pkg         = flag.String("package", "", "Specify a package name. eg:-package=errors")
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("[PANIC] %v\n%s\n", err, utils.BytesToString(debug.Stack()))
		}
	}()

	// 开始解析命令行
	flag.Parse()
	// 命令行参数都不匹配，打印help
	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}
	// 必须参数为空，打印help
	if *source == "" || *destination == "" || *pkg == "" {
		flag.Usage()
		return
	}
}
