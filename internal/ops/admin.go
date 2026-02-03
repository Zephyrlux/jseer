package ops

import (
	"expvar"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"go.uber.org/zap"
)

func StartAdminServer(addr string, enablePprof bool, logger *zap.Logger) {
	if addr == "" {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/debug/vars", expvar.Handler())
	if enablePprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !isNetClosed(err) {
			if logger != nil {
				logger.Warn("admin server stopped", zap.Error(err))
			}
		}
	}()
	if logger != nil {
		logger.Info("admin server listening", zap.String("addr", addr))
	}
}

func isNetClosed(err error) bool {
	if err == nil {
		return false
	}
	if err == http.ErrServerClosed {
		return true
	}
	if ne, ok := err.(*net.OpError); ok && !ne.Temporary() {
		return true
	}
	return false
}
