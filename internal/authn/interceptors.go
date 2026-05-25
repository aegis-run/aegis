package authn

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aegis-run/aegis/pkg/telemetry"
)

func authFunc(a Authenticator) auth.AuthFunc {
	return func(ctx context.Context) (_ context.Context, err error) {
		ctx, span := telemetry.Start(ctx, "authn.authenticate")
		defer telemetry.End(span, &err)

		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			telemetry.Attr(ctx, attribute.Bool("auth.authorized", false))
			return nil, err
		}

		id, err := a.Authenticate(ctx, token)
		if err != nil {
			telemetry.Attr(ctx, attribute.Bool("auth.authorized", false))
			err = status.Error(codes.Unauthenticated, "invalid api key")
			return nil, err
		}

		enrichSpan(ctx, id)
		return ContextWithIdentity(ctx, id), nil
	}
}

func enrichSpan(ctx context.Context, id *Identity) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	telemetry.Attr(ctx,
		attribute.String("auth.identity", id.ID),
		attribute.Bool("auth.authorized", true),
	)
}

var exemptMethods = map[string]struct{}{
	"/grpc.health.v1.Health/Check":                                   {},
	"/grpc.health.v1.Health/Watch":                                   {},
	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": {},
	"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo":      {},
}

func matchFunc(_ context.Context, callMeta interceptors.CallMeta) bool {
	_, exempt := exemptMethods[callMeta.FullMethod()]
	return !exempt
}

func UnaryServerInterceptor(a Authenticator) grpc.UnaryServerInterceptor {
	return selector.UnaryServerInterceptor(
		auth.UnaryServerInterceptor(authFunc(a)),
		selector.MatchFunc(matchFunc),
	)
}

func StreamServerInterceptor(a Authenticator) grpc.StreamServerInterceptor {
	return selector.StreamServerInterceptor(
		auth.StreamServerInterceptor(authFunc(a)),
		selector.MatchFunc(matchFunc),
	)
}
