package webserver

import (
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/limingyao/excellent-go/metrics/prometheus"
	"google.golang.org/grpc"
)

type ServerOption func(*Webserver)

func WithAddr(ip string, port int) ServerOption {
	return func(s *Webserver) {
		s.ip = ip
		s.port = port
	}
}

func WithPort(port int) ServerOption {
	return func(s *Webserver) {
		s.port = port
	}
}

func WithGatewayOptions(opts ...runtime.ServeMuxOption) ServerOption {
	return func(s *Webserver) {
		s.gatewayOptions = opts
	}
}

func WithDialOptions(opts ...grpc.DialOption) ServerOption {
	return func(s *Webserver) {
		s.dialOptions = opts
	}
}

func WithServerOptions(opts ...grpc.ServerOption) ServerOption {
	return func(s *Webserver) {
		s.serverOptions = opts
	}
}

func WithHealthz() ServerOption {
	return func(s *Webserver) {
		s.enableHealthz = true
	}
}

func WithHealthzPath(path string) ServerOption {
	return func(s *Webserver) {
		s.enableHealthz = true
		s.healthzPath = strings.TrimSuffix(path, "/")
	}
}

func WithReflection() ServerOption {
	return func(s *Webserver) {
		s.enableReflection = true
	}
}

func WithPProf() ServerOption {
	return func(s *Webserver) {
		s.enablePProf = true
	}
}

func WithPProfPath(path string) ServerOption {
	return func(s *Webserver) {
		s.enablePProf = true
		s.pprofPath = strings.TrimSuffix(path, "/")
	}
}

func WithPrometheus(opts ...prometheus.Option) ServerOption {
	return func(s *Webserver) {
		s.enablePrometheus = true
		s.prometheusOptions = opts
	}
}

func WithPrometheusPath(path string, opts ...prometheus.Option) ServerOption {
	return func(s *Webserver) {
		s.enablePrometheus = true
		s.prometheusOptions = opts
		s.prometheusPath = strings.TrimSuffix(path, "/")
	}
}
