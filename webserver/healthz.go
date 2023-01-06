package webserver

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	pb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthStatus pb.HealthCheckResponse_ServingStatus

const (
	UnknownStatus        = HealthStatus(pb.HealthCheckResponse_UNKNOWN)
	ServingStatus        = HealthStatus(pb.HealthCheckResponse_SERVING)
	NotServingStatus     = HealthStatus(pb.HealthCheckResponse_NOT_SERVING)
	ServiceUnknownStatus = HealthStatus(pb.HealthCheckResponse_SERVICE_UNKNOWN)
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
