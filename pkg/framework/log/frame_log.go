package log

import (
	"fmt"

	"github.com/soulnov23/go-tool/pkg/framework/plugin"
	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const pluginName = "frame_log"

var DefaultLogger log.Logger

func init() {
	plugin.Register(pluginName, &FrameLogPlugin{})
}

type FrameLogPlugin struct{}

func (p *FrameLogPlugin) Name() string {
	return pluginName
}

func (p *FrameLogPlugin) Setup(node yaml.Node) error {
	config := &log.Config{}
	if err := node.Decode(config); err != nil {
		return fmt.Errorf("plugin name[%s] invalid config: %v", pluginName, err)
	}
	logger, err := log.New(config)
	if err != nil {
		return fmt.Errorf("plugin name[%s] new logger: %v", pluginName, err)
	}
	DefaultLogger = logger.With(zap.String("name", "frame"))
	return nil
}
