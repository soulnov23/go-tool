package plugin

import (
	"fmt"
	"reflect"
	"sync"

	"errors"

	"gopkg.in/yaml.v3"
)

var (
	plugins = map[string]Plugin{}
	mutex   = sync.Mutex{}
)

type Plugin interface {
	Name() string
	Setup(node *yaml.Node) error
}

func Register(name string, plugin Plugin) error {
	value := reflect.ValueOf(plugin)
	if plugin == nil || value.Kind() == reflect.Pointer && value.IsNil() {
		return errors.New("register nil plugin")
	}
	if name == "" {
		return errors.New("register empty name of plugin")
	}
	mutex.Lock()
	defer mutex.Unlock()
	plugins[name] = plugin
	return nil
}

type Config map[string]*yaml.Node

func (c Config) Setup() error {
	for name, cfg := range c {
		plugin := plugins[name]
		if plugin == nil {
			return fmt.Errorf("plugin name[%s] not found", name)
		}
		if err := plugin.Setup(cfg); err != nil {
			return err
		}
	}
	return nil
}
