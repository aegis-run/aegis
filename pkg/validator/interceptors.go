package validator

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/aegis-run/aegis/pkg/telemetry"
)

// UnaryServerInterceptor returns a new unary server interceptor that validates incoming
// messages. If the request is invalid, clients may access a structured representation of
// the validation failure as an error detail.
func UnaryServerInterceptor(validator protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		if err := validateFunc(ctx, req, validator); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that validates
// incoming messages. If the request is invalid, clients may access a structured
// representation of the validation failure as an error detail.
func StreamServerInterceptor(validator protovalidate.Validator) grpc.StreamServerInterceptor {
	return func(
		srv any,
		stream grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		return handler(srv, &wrappedServerStream{
			ServerStream: stream,
			validator:    validator,
		})
	}
}

// wrappedServerStream is a thin wrapper around grpc.ServerStream that allows modifying context.
type wrappedServerStream struct {
	grpc.ServerStream

	validator protovalidate.Validator
}

func (w *wrappedServerStream) RecvMsg(m any) error {
	if err := w.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	return validateFunc(w.Context(), m, w.validator)
}

func validateFunc(ctx context.Context, m any, validator protovalidate.Validator) (err error) {
	ctx, span := telemetry.Start(ctx, "validator.validate")
	defer telemetry.End(span, &err)

	msg, ok := m.(proto.Message)
	if !ok {
		return status.Errorf(codes.Internal, "unsupported message type: %T", m)
	}

	telemetry.Attr(ctx,
		attribute.String("validation.message_type", string(msg.ProtoReflect().Descriptor().FullName())),
	)

	err = validator.Validate(msg)
	if err == nil {
		telemetry.Attr(ctx,
			attribute.String("validation.status", "success"),
		)
		return nil
	}

	telemetry.Attr(ctx,
		attribute.String("validation.status", "failed"),
		attribute.Bool("validation.invalid", true),
	)

	if valErr, ok := errors.AsType[*protovalidate.ValidationError](err); ok {
		// Message is invalid.
		telemetry.Attr(ctx,
			attribute.Int("validation.violations", len(valErr.Violations)),
		)

		st := status.New(codes.InvalidArgument, err.Error())

		ds, detErr := st.WithDetails(valErr.ToProto())
		if detErr != nil {
			return st.Err()
		}

		return ds.Err()
	}

	// CEL expression doesn't compile or type-check.
	return status.Error(codes.Internal, err.Error())
}
