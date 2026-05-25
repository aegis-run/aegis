package logger

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
)

func logger() logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func UnaryServerInterceptor(opts ...logging.Option) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(logger(), opts...)
}

func StreamServerInterceptor(opts ...logging.Option) grpc.StreamServerInterceptor {
	return logging.StreamServerInterceptor(logger(), opts...)
}
