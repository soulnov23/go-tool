package framework

import (
	"time"

	"github.com/soulnov23/go-tool/pkg/framework/plugin"
)

const (
	defaultConfigPath               = "../conf/go_tool.yaml"
	defaultUpdateGOMAXPROCSInterval = 60000
)

type Config struct {
	ProfileProfiler *struct {
		Address      string        `yaml:"address"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
		IdleTimeout  time.Duration `yaml:"idle_timeout"`
	} `yaml:"pprof"`

	Server *struct {
		UpdateGOMAXPROCSInterval time.Duration `yaml:"update_gomaxprocs_interval"`
		MaxCloseWaitTime         time.Duration `yaml:"max_close_wait_time"`

		Services []*struct {
			Name     string        `yaml:"name"`
			Address  string        `yaml:"address"`
			Network  string        `yaml:"network"`
			Protocol string        `yaml:"protocol"`
			Timeout  time.Duration `yaml:"timeout"`
		} `yaml:"services"`
	} `yaml:"server"`

	Client *struct {
		Services []*struct {
			Name     string        `yaml:"name"`
			Address  string        `yaml:"address"`
			Network  string        `yaml:"network"`
			Protocol string        `yaml:"protocol"`
			Timeout  time.Duration `yaml:"timeout"`
		} `yaml:"services"`
	} `yaml:"client"`

	Plugins plugin.Config `yaml:"plugins"`
}
