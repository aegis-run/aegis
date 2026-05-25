package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func withMainSpan(ctx context.Context) context.Context {
	return InjectMainSpan(ctx, trace.SpanFromContext(ctx))
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		return handler(withMainSpan(ctx), req)
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		return handler(srv, &wrappedServerStream{
			ServerStream: ss,
			ctx:          withMainSpan(ss.Context()),
		})
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context { return w.ctx }
