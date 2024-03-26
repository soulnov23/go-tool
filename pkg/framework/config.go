package framework

import (
	"github.com/soulnov23/go-tool/pkg/framework/plugin"
)

const (
	defaultConfigPath               = "../conf/go_tool.yaml"
	defaultUpdateGOMAXPROCSInterval = 60000
)

type Config struct {
	ProfileProfiler *struct {
		Address      string `yaml:"address"`
		ReadTimeout  int64  `yaml:"read_timeout"`
		WriteTimeout int64  `yaml:"write_timeout"`
		IdleTimeout  int64  `yaml:"idle_timeout"`
	} `yaml:"pprof"`

	Server *struct {
		UpdateGOMAXPROCSInterval int64 `yaml:"update_gomaxprocs_interval"`
		MaxCloseWaitTime         int64 `yaml:"max_close_wait_time"`

		Services []*struct {
			Name     string `yaml:"name"`
			Address  string `yaml:"address"`
			Network  string `yaml:"network"`
			Protocol string `yaml:"protocol"`
			Timeout  int64  `yaml:"timeout"`
		} `yaml:"services"`
	} `yaml:"server"`

	Plugins plugin.Config `yaml:"plugins"`
}
