package internal

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/soulnov23/go-tool/pkg/log"
	convert "github.com/soulnov23/go-tool/pkg/strconv"
)

type AppConfig struct {
	Server   ServerConfig  `yaml:"server"`
	Client   ClientConfig  `yaml:"client"`
	FrameLog log.LogConfig `yaml:"frame_log"`
	RunLog   log.LogConfig `yaml:"run_log"`
}

type ServerConfig struct {
	Debug    DebugConfig     `yaml:"debug"`
	Services []ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	Debug    DebugConfig `yaml:"debug"`
	Name     string      `yaml:"name"`
	Network  string      `yaml:"network"`
	Address  string      `yaml:"address"`
	Protocol string      `yaml:"protocol"`
	Timeout  string      `yaml:"timeout"`
}

type DebugConfig struct {
	Address      string `yaml:"address"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
}

type ClientConfig struct {
	Services []ServiceConfig `yaml:"services"`
}

func GetAppConfig(path string) (*AppConfig, error) {
	buffer, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("read config file " + path + ": " + err.Error())
	}
	buffer = convert.StringToBytes(os.ExpandEnv(convert.BytesToString(buffer)))
	appConfig := &AppConfig{}
	if err = yaml.Unmarshal(buffer, appConfig); err != nil {
		return nil, errors.New("unmarshal " + path + ": " + err.Error())
	}
	return appConfig, nil
}
