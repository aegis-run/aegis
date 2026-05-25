package servers

import (
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer = health.Server

func NewHealthServer() *HealthServer {
	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	return hs
}

// type HealthServer struct {
// 	health.UnimplementedHealthServer
// }

// var _ health.HealthServer = (*HealthServer)(nil)

// func NewHealthServer() *HealthServer {
// 	return &HealthServer{}
// }

// func (s *HealthServer) Check(
// 	_ context.Context,
// 	_ *health.HealthCheckRequest,
// ) (*health.HealthCheckResponse, error) {
// 	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_SERVING}, nil
// }

// func (s *HealthServer) Watch(
// 	_ *health.HealthCheckRequest,
// 	_ grpc.ServerStreamingServer[health.HealthCheckResponse],
// ) error {
// 	return status.Error(codes.Unimplemented, "unimplemented streaming endpoint")
// }
