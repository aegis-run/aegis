package servers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

type Pprof struct {
	base
	srv *http.Server
}

func NewPprof(cfg *PprofConfig) (*Pprof, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: mux,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  15 * time.Second}

	return &Pprof{
		base: base{name: "pprof"},
		srv:  srv,
	}, nil
}

func (srv *Pprof) Serve(_ context.Context, ln net.Listener) error {
	if !srv.listen(ln) {
		return nil
	}

	err := srv.srv.Serve(ln)
	return srv.isServeErr(err)
}

func (srv *Pprof) Shutdown(ctx context.Context) error {
	srv.shutdown()
	if err := srv.srv.Shutdown(ctx); err != nil {
		logger.Error("pprof.shutdown_failed",
			"error", err,
		)
		return fmt.Errorf("pprof shutdown failed: %w", err)
	}
	return nil
}
