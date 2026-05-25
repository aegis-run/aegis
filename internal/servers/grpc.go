package servers

import (
	"context"
	"net"

	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/aegis-run/aegis/internal/authn"
	"github.com/aegis-run/aegis/pkg/telemetry"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
	"github.com/aegis-run/aegis/pkg/validator"
	schemav1 "github.com/aegis-run/aegis/proto/aegis/schema/v1"
)

type GRPC struct {
	base
	srv    *grpc.Server
	health *HealthServer
}

type Dependencies struct {
	Authn     authn.Authenticator
	Config    *Config
	Validator validator.Validator
	Schema    schemav1.SchemaServer
}

func NewGRPC(deps *Dependencies) (*GRPC, error) {
	opts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(unaryInterceptors(deps)...),
		grpc.ChainStreamInterceptor(streamInterceptors(deps)...),
	}

	if deps.Config.GRPC.TLS.Enabled {
		c, err := credentials.NewServerTLSFromFile(
			deps.Config.GRPC.TLS.CertPath,
			deps.Config.GRPC.TLS.KeyPath,
		)
		if err != nil {
			logger.Error("grpc.server.tls_load_failed",
				"error", err,
				"cert_path", deps.Config.GRPC.TLS.CertPath,
				"key_path", deps.Config.GRPC.TLS.KeyPath,
			)
			return nil, err
		}

		opts = append(opts, grpc.Creds(c))
	}

	srv := grpc.NewServer(opts...)
	hs := NewHealthServer()
	health.RegisterHealthServer(srv, hs)
	if deps.Schema != nil {
		schemav1.RegisterSchemaServer(srv, deps.Schema)
	}
	reflection.Register(srv)

	return &GRPC{
		base:   base{name: "grpc.server"},
		srv:    srv,
		health: hs,
	}, nil
}

func unaryInterceptors(deps *Dependencies) []grpc.UnaryServerInterceptor {
	interceptors := []grpc.UnaryServerInterceptor{
		recovery.UnaryServerInterceptor(),
		telemetry.UnaryServerInterceptor(),
	}

	if deps.Authn != nil {
		interceptors = append(interceptors, authn.UnaryServerInterceptor(deps.Authn))
	}

	interceptors = append(interceptors,
		logger.UnaryServerInterceptor(
			logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		),
		validator.UnaryServerInterceptor(deps.Validator),
	)

	return interceptors
}

func streamInterceptors(deps *Dependencies) []grpc.StreamServerInterceptor {
	interceptors := []grpc.StreamServerInterceptor{
		recovery.StreamServerInterceptor(),
		telemetry.StreamServerInterceptor(),
	}

	if deps.Authn != nil {
		interceptors = append(interceptors, authn.StreamServerInterceptor(deps.Authn))
	}

	interceptors = append(interceptors,
		logger.StreamServerInterceptor(
			logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		),
		validator.StreamServerInterceptor(deps.Validator),
	)

	return interceptors
}

func (srv *GRPC) Serve(_ context.Context, ln net.Listener) error {
	if !srv.listen(ln) {
		return nil
	}

	srv.health.SetServingStatus("", health.HealthCheckResponse_SERVING)
	err := srv.srv.Serve(ln)
	return srv.isServeErr(err)
}

func (srv *GRPC) Shutdown(ctx context.Context) error {
	srv.shutdown()
	srv.health.SetServingStatus("", health.HealthCheckResponse_NOT_SERVING)

	stopped := make(chan struct{})
	go func() {
		srv.srv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		logger.Warn("grpc.server.graceful_shutdown_timeout_force_stop")
		srv.srv.Stop()
		return ctx.Err()
	}
}
