package log

import (
	"fmt"

	"github.com/soulnov23/go-tool/pkg/framework/plugin"
	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const pluginName = "frame_log"

var defaultLogger log.Logger

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
	defaultLogger = logger.With(zap.String("name", "frame"))
	return nil
}

func With(fields ...zap.Field) log.Logger {
	return defaultLogger.With(fields...)
}

func Debug(args ...any) {
	defaultLogger.Debug(args...)
}

func Debugf(formatter string, args ...any) {
	defaultLogger.Debugf(formatter, args...)
}

func DebugFields(msg string, fields ...zap.Field) {
	defaultLogger.DebugFields(msg, fields...)
}

func Info(args ...any) {
	defaultLogger.Info(args...)
}

func Infof(formatter string, args ...any) {
	defaultLogger.Infof(formatter, args...)
}

func InfoFields(msg string, fields ...zap.Field) {
	defaultLogger.InfoFields(msg, fields...)
}

func Warn(args ...any) {
	defaultLogger.Warn(args...)
}

func Warnf(formatter string, args ...any) {
	defaultLogger.Warnf(formatter, args...)
}

func WarnFields(msg string, fields ...zap.Field) {
	defaultLogger.WarnFields(msg, fields...)
}

func Error(args ...any) {
	defaultLogger.Error(args...)
}

func Errorf(formatter string, args ...any) {
	defaultLogger.Errorf(formatter, args...)
}

func ErrorFields(msg string, fields ...zap.Field) {
	defaultLogger.ErrorFields(msg, fields...)
}

func Fatal(args ...any) {
	defaultLogger.Fatal(args...)
}

func Fatalf(formatter string, args ...any) {
	defaultLogger.Fatalf(formatter, args...)
}

func FatalFields(msg string, fields ...zap.Field) {
	defaultLogger.FatalFields(msg, fields...)
}

func Sync() error {
	return defaultLogger.Sync()
}
