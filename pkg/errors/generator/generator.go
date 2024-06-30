package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/soulnov23/go-tool/pkg/errors"
	"github.com/soulnov23/go-tool/pkg/utils"
	"gopkg.in/yaml.v3"
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

	// 读取yaml文件
	var configs []*errors.Error
	for _, file := range strings.Split(*source, ",") {
		buffer, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("read yaml file %s: %v", file, err)
			return
		}
		var tmp []*errors.Error
		if err = yaml.Unmarshal(buffer, &tmp); err != nil {
			fmt.Printf("unmarshal yaml file %s: %v", file, err)
			return
		}
		for _, config := range tmp {
			if config.Validate() != nil {
				fmt.Printf("validate yaml file %s: %v", file, err)
				return
			}
		}
		configs = append(configs, tmp...)
	}

	// 生成go代码
	file, err := os.Create(*destination)
	if err != nil {
		fmt.Printf("create go file %s: %v", *destination, err)
		return
	}
	tpl, err := template.New("generator error").Parse(templateErrors)
	if err != nil {
		fmt.Printf("parse template: %v", err)
		return
	}
	if err := tpl.Execute(file, map[string]any{
		"source":  *source,
		"package": *pkg,
		"configs": configs,
	}); err != nil {
		fmt.Printf("execute template: %v", err)
		return
	}
}
