package servers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aegis-run/aegis/pkg/telemetry"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

type HTTP struct {
	base

	srv  *http.Server
	conn *grpc.ClientConn
}

func NewHTTP(cfg *Config) (*HTTP, error) {
	opts := []grpc.DialOption{}
	if cfg.HTTP.TLS.Enabled {
		c, err := credentials.NewClientTLSFromFile(cfg.HTTP.TLS.CertPath, cfg.NameOverride)
		if err != nil {
			logger.Error("http.server.tls_load_failed",
				"error", err,
				"cert_path", cfg.HTTP.TLS.CertPath,
			)
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(c))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(net.JoinHostPort(cfg.HTTP.GRPCTargetHost, cfg.GRPC.Port), opts...)
	if err != nil {
		return nil, err
	}

	healthClient := health.NewHealthClient(conn)
	muxOpts := []runtime.ServeMuxOption{
		runtime.WithHealthzEndpoint(healthClient),

		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	}

	mux := runtime.NewServeMux(muxOpts...)

	httpMux := http.NewServeMux()
	httpMux.Handle("/", mux)
	httpMux.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.HTTP.Port),
		Handler:           injectWideEvent(httpMux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &HTTP{
		base: base{name: "http.server"},
		srv:  srv,
		conn: conn,
	}, nil
}

func (srv *HTTP) Serve(_ context.Context, ln net.Listener) error {
	if !srv.listen(ln) {
		return nil
	}

	err := srv.srv.Serve(ln)
	return srv.isServeErr(err)
}

func (srv *HTTP) Shutdown(ctx context.Context) error {
	srv.shutdown()

	if err := srv.conn.Close(); err != nil {
		logger.Error("http.server.closing_http_conn_failed",
			"error", err,
		)
		return err
	}

	if err := srv.srv.Shutdown(ctx); err != nil {
		logger.Error("http.shutdown_failed",
			"error", err,
		)
		return fmt.Errorf("http server shutdown failed: %w", err)
	}
	return nil
}

func injectWideEvent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wideCtx := telemetry.InjectMainSpan(r.Context(), trace.SpanFromContext(r.Context()))
		next.ServeHTTP(w, r.WithContext(wideCtx))
	})
}
