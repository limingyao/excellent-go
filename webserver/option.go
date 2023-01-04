package webserver

import (
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type serverOption func(*webserver)

func WithAddr(ip string, port int) serverOption {
	return func(s *webserver) {
		s.ip = ip
		s.port = port
	}
}

func WithPort(port int) serverOption {
	return func(s *webserver) {
		s.port = port
	}
}

func WithGatewayOptions(opts ...runtime.ServeMuxOption) serverOption {
	return func(s *webserver) {
		s.gatewayOptions = opts
	}
}

func WithDialOptions(opts ...grpc.DialOption) serverOption {
	return func(s *webserver) {
		s.dialOptions = opts
	}
}

func WithServerOptions(opts ...grpc.ServerOption) serverOption {
	return func(s *webserver) {
		s.serverOptions = opts
	}
}

func WithHealthz() serverOption {
	return func(s *webserver) {
		s.enableHealthz = true
	}
}

func WithHealthzPath(path string) serverOption {
	return func(s *webserver) {
		s.enableHealthz = true
		s.healthzPath = strings.TrimLeft(path, "/")
	}
}

func WithReflection() serverOption {
	return func(s *webserver) {
		s.enableReflection = true
	}
}

func WithPProf() serverOption {
	return func(s *webserver) {
		s.enablePProf = true
	}
}

func WithPProfPath(path string) serverOption {
	return func(s *webserver) {
		s.enablePProf = true
		s.pprofPath = strings.TrimLeft(path, "/")
	}
}
