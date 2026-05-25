package aegis

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/aegis-run/aegis/internal"
	"github.com/aegis-run/aegis/internal/aegis/config"
	"github.com/aegis-run/aegis/internal/authn"
	"github.com/aegis-run/aegis/internal/datalayer"
	"github.com/aegis-run/aegis/internal/schema"
	"github.com/aegis-run/aegis/internal/servers"
	"github.com/aegis-run/aegis/pkg/db"
	"github.com/aegis-run/aegis/pkg/must"
	"github.com/aegis-run/aegis/pkg/runtime"
	"github.com/aegis-run/aegis/pkg/telemetry"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
	"github.com/aegis-run/aegis/pkg/validator"
)

func Run(ctx context.Context, cfg *config.Config) (err error) {
	rt := runtime.New()
	defer rt.Recover()

	if err := configureTelemetry(ctx, rt, cfg); err != nil {
		return err
	}

	logger.Info("aegis.starting",
		"id", internal.Identifier,
		"version", internal.Version,
	)

	deps := configureDeps(ctx, rt, cfg)

	if err := configureServers(ctx, rt, deps); err != nil {
		return err
	}

	if err := configurePprof(ctx, rt, &cfg.Profiler); err != nil {
		return err
	}

	if err := rt.Run(ctx, runtime.WithTimeout(time.Minute)); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("aegis.stopped")

	return nil
}

func configureDeps(
	ctx context.Context,
	rt *runtime.Runtime,
	cfg *config.Config,
) *servers.Dependencies {
	database := must.NotError(db.New(ctx, &cfg.Datastore))
	rt.Defer(database.Close)

	dl := must.NotError(datalayer.New(database))

	deps := &servers.Dependencies{
		Config:    &cfg.Server,
		Validator: must.NotError(validator.New()),
		Schema:    schema.NewAPI(dl.Schema),
	}

	if cfg.Authn.Enabled {
		deps.Authn = must.NotError(authn.New(ctx, &cfg.Authn))
	}

	return deps
}

func configureTelemetry(ctx context.Context, rt *runtime.Runtime, cfg *config.Config) error {
	shutdown, err := telemetry.ConfigureLogs(ctx, &cfg.Logs)
	if err != nil {
		return fmt.Errorf("unable to configure logs: %w", err)
	}
	rt.DeferFunc(shutdown)

	shutdown, err = telemetry.ConfigureTracing(ctx, &cfg.Tracing)
	if err != nil {
		return fmt.Errorf("unable to configure tracing: %w", err)
	}
	rt.DeferFunc(shutdown)

	shutdown, err = telemetry.ConfigureMetrics(ctx, &cfg.Metrics)
	if err != nil {
		return fmt.Errorf("unable to configure metrics: %w", err)
	}
	rt.DeferFunc(shutdown)

	return nil
}

func configureServers(
	ctx context.Context,
	rt *runtime.Runtime,
	deps *servers.Dependencies,
) error {
	grpc, err := servers.NewGRPC(deps)
	if err != nil {
		return err
	}
	rt.DeferFunc(grpc.Shutdown)

	ln, err := newListener(ctx, deps.Config.GRPC.Port)
	if err != nil {
		return err
	}
	rt.Defer(ln.Close)

	rt.Go(func(ctx context.Context) error {
		return grpc.Serve(ctx, ln)
	})

	if deps.Config.HTTP.Enabled {
		http, err := servers.NewHTTP(deps.Config)
		if err != nil {
			return err
		}
		rt.DeferFunc(http.Shutdown)

		ln, err := newListener(ctx, deps.Config.HTTP.Port)
		if err != nil {
			return err
		}
		rt.Defer(ln.Close)

		rt.Go(func(ctx context.Context) error {
			return http.Serve(ctx, ln)
		})
	}

	return nil
}

func configurePprof(
	ctx context.Context,
	rt *runtime.Runtime,
	cfg *servers.PprofConfig,
) error {
	if cfg.Enabled {
		pprof, err := servers.NewPprof(cfg)
		if err != nil {
			return err
		}
		rt.DeferFunc(pprof.Shutdown)

		ln, err := newListener(ctx, cfg.Port)
		if err != nil {
			return err
		}
		rt.Defer(ln.Close)

		rt.Go(func(ctx context.Context) error {
			return pprof.Serve(ctx, ln)
		})
	}

	return nil
}

func newListener(ctx context.Context, port string) (net.Listener, error) {
	ln, err := new(net.ListenConfig{}).Listen(ctx, "tcp", ":"+port)
	if err != nil {
		logger.Error("tcp.listen_failed",
			"error", err,
			"port", port,
		)
		return nil, fmt.Errorf("unable to listen on port %s: %w", port, err)
	}

	return ln, nil
}
