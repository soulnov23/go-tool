package pprof

import (
	"net/http"
	"net/http/pprof"
)

const defaultAddress = "0.0.0.0:6060"

// "Profile"是指性能分析数据，"Profiler"是指生成和处理这些数据的工具
type ProfileProfiler struct {
	opts *Options
}

func New(opts ...Option) *ProfileProfiler {
	pprof := &ProfileProfiler{
		opts: &Options{
			Address:      defaultAddress,
			ReadTimeout:  0, // not limited
			WriteTimeout: 0, // not limited
			IdleTimeout:  0, // not limited
		},
	}
	for _, opt := range opts {
		opt(pprof.opts)
	}
	return pprof
}

// 创建mux自定义处理函数，避免与pprof的默认http.DefaultServeMux冲突
// mux := http.NewServeMux()
// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
// http.ListenAndServe("ip:port", mux)
func (pp *ProfileProfiler) Serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	pprofServer := &http.Server{
		Addr:         pp.opts.Address,
		Handler:      mux,
		ReadTimeout:  pp.opts.ReadTimeout,
		WriteTimeout: pp.opts.WriteTimeout,
		IdleTimeout:  pp.opts.IdleTimeout,
	}
	return pprofServer.ListenAndServe()
}
