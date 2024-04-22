package framework

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/soulnov23/go-tool/pkg/framework/log"
	"github.com/soulnov23/go-tool/pkg/pprof"
	"github.com/soulnov23/go-tool/pkg/utils"
)

var (
	DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV}
	DefaultHotRestartSIG  = []os.Signal{syscall.SIGUSR1}
	DefaultTriggerSIG     = []os.Signal{syscall.SIGUSR2}
)

type Server struct {
	updateGOMAXPROCSInterval time.Duration
	maxCloseWaitTime         time.Duration // max waiting time when closing server
	*pprof.ProfileProfiler

	services map[string]*service // k=service_name,v=Service
}

func New(configPath string) *Server {
	config, err := loadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("loadConfig: %v", err))
	}

	if config.Server == nil {
		panic("server is empty")
	}

	if config.Plugins == nil {
		panic("server is empty")
	}
	if err := config.Plugins.Setup(); err != nil {
		panic(fmt.Sprintf("config.Plugins.Setup: %v", err))
	}

	server := &Server{
		updateGOMAXPROCSInterval: time.Duration(config.Server.UpdateGOMAXPROCSInterval) * time.Millisecond,
		maxCloseWaitTime:         time.Duration(config.Server.MaxCloseWaitTime) * time.Millisecond,
		services:                 make(map[string]*service),
	}

	if config.ProfileProfiler != nil {
		opts := []pprof.Option{
			pprof.WithAddress(config.ProfileProfiler.Address),
			pprof.WithReadTimeout(time.Duration(config.ProfileProfiler.ReadTimeout) * time.Millisecond),
			pprof.WithWriteTimeout(time.Duration(config.ProfileProfiler.WriteTimeout) * time.Millisecond),
			pprof.WithIdleTimeout(time.Duration(config.ProfileProfiler.IdleTimeout) * time.Millisecond),
		}
		server.ProfileProfiler = pprof.New(opts...)
	}

	for _, serviceConfig := range config.Server.Services {
		server.services[serviceConfig.Name] = newService(serviceConfig.Name, serviceConfig.Address, serviceConfig.Network, serviceConfig.Protocol, time.Duration(serviceConfig.Timeout)*time.Millisecond)
	}

	return server
}

func loadConfig(configPath string) (*Config, error) {
	buffer, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config path: %v", err)
	}
	buffer = utils.StringToBytes(os.ExpandEnv(utils.BytesToString(buffer)))
	config := &Config{}
	if err = yaml.Unmarshal(buffer, config); err != nil {
		return nil, fmt.Errorf("yaml unmarshal config: %v", err)
	}
	return config, nil
}

func (s *Server) Register(serviceName string, rpcName string, handler Handler) error {
	service, ok := s.services[serviceName]
	if !ok {
		return fmt.Errorf("not found service: %s", serviceName)
	}
	return service.register(rpcName, handler)
}

func (s *Server) Serve() error {
	defer log.DefaultLogger.Sync()

	utils.UpdateGOMAXPROCS(log.DefaultLogger.Infof, s.updateGOMAXPROCSInterval)

	if s.ProfileProfiler != nil {
		go func() {
			if err := s.ProfileProfiler.Serve(); err != nil {
				log.DefaultLogger.FatalFields("pprof Serve failed", zap.Reflect("pprof_server", s.ProfileProfiler), zap.Error(err))
			}
		}()
	}

	for n, s := range s.services {
		go func(name string, service *service) {
			if err := service.serve(); err != nil {
				log.DefaultLogger.FatalFields("service serve", zap.String("service_name", name), zap.Error(err))
			}
			service.close()
		}(n, s)
	}

	signalClose := make(chan os.Signal, 1)
	signal.Notify(signalClose, DefaultServerCloseSIG...)
	signalHotRestart := make(chan os.Signal, 1)
	signal.Notify(signalHotRestart, DefaultHotRestartSIG...)
	signalTrigger := make(chan os.Signal, 1)
	signal.Notify(signalTrigger, DefaultTriggerSIG...)
	select {
	case sig := <-signalClose:
		log.DefaultLogger.InfoFields("signal close", zap.String("sig", sig.String()))
	case sig := <-signalHotRestart:
		log.DefaultLogger.InfoFields("signal hot restart", zap.String("sig", sig.String()))
	case sig := <-signalTrigger:
		log.DefaultLogger.InfoFields("signal trigger", zap.String("sig", sig.String()))
	}
	return nil
}
