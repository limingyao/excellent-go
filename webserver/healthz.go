package webserver

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	pb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthStatus pb.HealthCheckResponse_ServingStatus

const (
	UNKNOWN         = HealthStatus(pb.HealthCheckResponse_UNKNOWN)
	SERVING         = HealthStatus(pb.HealthCheckResponse_SERVING)
	NOT_SERVING     = HealthStatus(pb.HealthCheckResponse_NOT_SERVING)
	SERVICE_UNKNOWN = HealthStatus(pb.HealthCheckResponse_SERVICE_UNKNOWN)
)

// https://github.com/grpc-ecosystem/grpc-health-probe
var healthServer *health.Server

func newHealthClient(cc grpc.ClientConnInterface) pb.HealthClient {
	return pb.NewHealthClient(cc)
}

func registerHealthServer(svr grpc.ServiceRegistrar) {
	healthServer = health.NewServer()
	pb.RegisterHealthServer(svr, healthServer)
}

func SetServerStatus(status HealthStatus) {
	if healthServer == nil {
		return
	}
	healthServer.SetServingStatus("", pb.HealthCheckResponse_ServingStatus(status))
}
