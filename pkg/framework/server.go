package framework

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"errors"

	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/pprof"
	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV}
	DefaultHotRestartSIG  = []os.Signal{syscall.SIGUSR1}
	DefaultTriggerSIG     = []os.Signal{syscall.SIGUSR2}
)

type Server struct {
	updateGOMAXPROCSInterval time.Duration
	maxCloseWaitTime         time.Duration // max waiting time when closing server
	log.Logger
	*pprof.ProfileProfiler

	services map[string]*service // k=service_name,v=Service
}

func New(configPath string) *Server {
	config, err := loadConfig(configPath)
	if err != nil {
		panic(err)
	}

	if config.Server == nil {
		panic("server is empty")
	}

	if config.Server.Log == nil {
		logger = log.DefaultLogger
	} else {
		log.DefaultLogger, err = log.New(config.Server.Log)
		if err != nil {
			panic(errors.NewInternalServerError(InvalidConfig, "new server log: %v", err))
		}
	}
	logger = logger.With(zap.String("name", "frame"))

	server := &Server{
		updateGOMAXPROCSInterval: config.Server.UpdateGOMAXPROCSInterval,
		maxCloseWaitTime:         config.Server.MaxCloseWaitTime,
		Logger:                   logger,
	}

	if config.ProfileProfiler != nil {
		opts := []pprof.Option{
			pprof.WithAddress(config.ProfileProfiler.Address),
			pprof.WithReadTimeout(config.ProfileProfiler.ReadTimeout),
			pprof.WithWriteTimeout(config.ProfileProfiler.WriteTimeout),
			pprof.WithIdleTimeout(config.ProfileProfiler.IdleTimeout),
		}
		server.ProfileProfiler = pprof.New(opts...)
	}

	for _, serviceConfig := range config.Server.Services {
		s := newService(serviceConfig.Name, serviceConfig.Address, serviceConfig.Network, serviceConfig.Protocol, serviceConfig.Timeout)
		server.services[serviceConfig.Name] = s
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
	defer s.Logger.Sync()

	close := utils.UpdateGOMAXPROCS(s.Logger.Debugf, s.updateGOMAXPROCSInterval)

	if s.ProfileProfiler != nil {
		go func() {
			if err := s.ProfileProfiler.Serve(); err != nil {
				s.Logger.FatalFields("pprof Serve failed", zap.Reflect("pprof_server", s.ProfileProfiler), zap.Error(err))
			}
		}()
	}

	signalClose := make(chan os.Signal, 1)
	signal.Notify(signalClose, DefaultServerCloseSIG...)
	signalHotRestart := make(chan os.Signal, 1)
	signal.Notify(signalHotRestart, DefaultHotRestartSIG...)
	signalTrigger := make(chan os.Signal, 1)
	signal.Notify(signalTrigger, DefaultTriggerSIG...)
	select {
	case sig := <-signalClose:
		s.Logger.DebugFields("signal close", zap.String("sig", sig.String()))
		eventLoop.Close()
	case sig := <-signalHotRestart:
		frameLog.DebugFields("signal hot restart", zap.String("sig", sig.String()))
	case sig := <-signalTrigger:
		frameLog.DebugFields("signal trigger", zap.String("sig", sig.String()))
		eventLoop.Trigger()
	}
}
