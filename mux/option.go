package mux

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type options struct {
	grpcOptions struct {
		opts            []grpc.ServerOption
		chainUnaryInts  []grpc.UnaryServerInterceptor
		chainStreamInts []grpc.StreamServerInterceptor
	}

	grpcGatewayOptions []runtime.ServeMuxOption
}

var (
	defaultOptions = options{}
)

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fo *funcOption) apply(o *options) {
	fo.f(o)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func (o *options) ClientDialOpts() []grpc.DialOption {
	return nil
}

//func ChainUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
//	return newFuncServerOption(func(o *serverOptions) {
//		o.grpcOptions.chainUnaryInts = append(o.grpcOptions.chainUnaryInts, interceptors...)
//	})
//}
//
//func ChainStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) ServerOption {
//	return newFuncServerOption(func(o *serverOptions) {
//		o.grpcOptions.chainStreamInts = append(o.grpcOptions.chainStreamInts, interceptors...)
//	})
//}
