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

func (p *FrameLogPlugin) Setup(node *yaml.Node) error {
	if node == nil {
		return fmt.Errorf("plugin name[%s] nil config", pluginName)
	}
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

func With(fields ...zap.Field) log.Logger {
	return DefaultLogger.With(fields...)
}

func Debug(args ...any) {
	DefaultLogger.Debug(args...)
}

func Debugf(formatter string, args ...any) {
	DefaultLogger.Debugf(formatter, args...)
}

func DebugFields(msg string, fields ...zap.Field) {
	DefaultLogger.DebugFields(msg, fields...)
}

func Info(args ...any) {
	DefaultLogger.Info(args...)
}

func Infof(formatter string, args ...any) {
	DefaultLogger.Infof(formatter, args...)
}

func InfoFields(msg string, fields ...zap.Field) {
	DefaultLogger.InfoFields(msg, fields...)
}

func Warn(args ...any) {
	DefaultLogger.Warn(args...)
}

func Warnf(formatter string, args ...any) {
	DefaultLogger.Warnf(formatter, args...)
}

func WarnFields(msg string, fields ...zap.Field) {
	DefaultLogger.WarnFields(msg, fields...)
}

func Error(args ...any) {
	DefaultLogger.Error(args...)
}

func Errorf(formatter string, args ...any) {
	DefaultLogger.Errorf(formatter, args...)
}

func ErrorFields(msg string, fields ...zap.Field) {
	DefaultLogger.ErrorFields(msg, fields...)
}

func Fatal(args ...any) {
	DefaultLogger.Fatal(args...)
}

func Fatalf(formatter string, args ...any) {
	DefaultLogger.Fatalf(formatter, args...)
}

func FatalFields(msg string, fields ...zap.Field) {
	DefaultLogger.FatalFields(msg, fields...)
}

func Sync() error {
	return DefaultLogger.Sync()
}
