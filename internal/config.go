package internal

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/utils"
)

type AppConfig struct {
	Server   []*ServiceConfig `yaml:"server"`
	Client   []*ServiceConfig `yaml:"client"`
	FrameLog log.LogConfig    `yaml:"frame_log"`
	RunLog   log.LogConfig    `yaml:"run_log"`
}

type ServiceConfig struct {
	Name     string `yaml:"name"`
	Network  string `yaml:"network"`
	Address  string `yaml:"address"`
	Protocol string `yaml:"protocol"`
	Timeout  string `yaml:"timeout"`
}

func GetAppConfig(path string) (*AppConfig, error) {
	buffer, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("read config file " + path + ": " + err.Error())
	}
	buffer = utils.String2Byte(os.ExpandEnv(utils.Byte2String(buffer)))
	appConfig := &AppConfig{}
	if err = yaml.Unmarshal(buffer, appConfig); err != nil {
		return nil, errors.New("unmarshal " + path + ": " + err.Error())
	}
	return appConfig, nil
}
