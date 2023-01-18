package internal

import (
	"errors"
	"flag"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/SoulNov23/go-tool/pkg/env"
	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/unsafe"
)

var DefaultConfPath = "./go_tool.yaml"

type AppConfig struct {
	Server   []*ServerConfig `yaml:"server"`
	Client   interface{}     `yaml:"client"`
	FrameLog log.LogConfig   `yaml:"frame_log"`
	CallLog  log.LogConfig   `yaml:"call_log"`
	RunLog   log.LogConfig   `yaml:"run_log"`
}

type ServerConfig struct {
	Name     string `yaml:"name"`
	Ip       string `yaml:"ip"`
	Port     string `yaml:"port"`
	Network  string `yaml:"network"`
	Protocol string `yaml:"protocol"`
	Timeout  string `yaml:"timeout"`
}

func GetAppConfig() (*AppConfig, error) {
	// 定义需要解析的命令行参数
	var confPath string
	flag.StringVar(&confPath, "conf", DefaultConfPath, "server config path")
	// 开始解析命令行
	flag.Parse()
	buffer, err := os.ReadFile(confPath)
	if err != nil {
		return nil, errors.New("os.ReadFile: " + err.Error())
	}
	buffer = unsafe.String2Byte(env.ExpandEnv(unsafe.Byte2String(buffer)))
	appConfig := &AppConfig{}
	if err = yaml.Unmarshal(buffer, appConfig); err != nil {
		return nil, errors.New("unmarshal " + DefaultConfPath + ": " + err.Error())
	}
	return appConfig, nil
}
