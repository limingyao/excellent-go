package mux

import (
	"context"
	"crypto/tls"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/limingyao/excellent-go/encoding/proto"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
	"sync"
)

type ServeMux struct {
	http.Server

	opts options
	once sync.Once

	// grpc server
	grpcServer *grpc.Server
	// grpc gateway server
	grpcGatewayServer *runtime.ServeMux
}

func Serve(l net.Listener, handler http.Handler, opts ...Option) error {
	srv := &ServeMux{}
	srv.Server.Handler = handler
	srv.init(opts...)
	return srv.Serve(l)
}

func ServeTLS(l net.Listener, handler http.Handler, certFile, keyFile string, opts ...Option) error {
	srv := &ServeMux{}
	srv.Server.Handler = handler
	srv.init(opts...)
	return srv.ServeTLS(l, certFile, keyFile)
}

func ListenAndServe(addr string, handler http.Handler, opts ...Option) error {
	srv := ServeMux{
		Server: http.Server{
			Addr: addr,
		},
	}
	srv.Server.Handler = handler
	srv.init(opts...)
	return srv.ListenAndServe()
}

func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler, opts ...Option) error {
	srv := ServeMux{
		Server: http.Server{
			Addr: addr,
		},
	}
	srv.Server.Handler = handler
	srv.init(opts...)
	return srv.ListenAndServeTLS(certFile, keyFile)
}

func NewServeMux(addr string, opts ...Option) *ServeMux {
	return NewServeMuxTLS(addr, nil, opts...)
}

func NewServeMuxTLS(addr string, tlsConfig *tls.Config, opts ...Option) *ServeMux {
	srv := &ServeMux{
		Server: http.Server{
			Addr:      addr,
			TLSConfig: tlsConfig,
		},
	}
	srv.init(opts...)
	return srv
}

func (x *ServeMux) init(opts ...Option) {
	x.once.Do(func() {
		defaultOpts := defaultOptions
		for _, o := range opts {
			o.apply(&defaultOpts)
		}
		x.opts = defaultOpts

		if x.TLSConfig != nil {
			// todo
		} else {

		}

		x.grpcServer = grpc.NewServer()                                                                                       // todo opt ...
		x.grpcGatewayServer = runtime.NewServeMux(runtime.WithMarshalerOption("application/proto", &proto.ProtoMarshaller{})) // todo opt ...
		x.Server.Handler = h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				x.grpcServer.ServeHTTP(w, r)
			} else {
				x.grpcGatewayServer.ServeHTTP(w, r)
			}
		}), &http2.Server{})
	})
}

func (x *ServeMux) Serve(l net.Listener) error {
	return x.Server.Serve(l)
}

func (x *ServeMux) ServeTLS(l net.Listener, certFile, keyFile string) error {
	return x.Server.ServeTLS(l, certFile, keyFile)
}

func (x *ServeMux) ListenAndServe() error {
	return x.Server.ListenAndServe()
}

func (x *ServeMux) ListenAndServeTLS(certFile, keyFile string) error {
	return x.Server.ListenAndServeTLS(certFile, keyFile)
}

// grpc next

type GRPCService interface {
	Register(srv *grpc.Server)
}

func (x *ServeMux) RegisterGRPCService(srv GRPCService) {
	srv.Register(x.grpcServer)
}

type GRPCServiceFunc func(srv *grpc.Server)

func (f GRPCServiceFunc) Register(srv *grpc.Server) {
	f(srv)
}

func (x *ServeMux) RegisterGRPCServiceFunc(srv func(srv *grpc.Server)) {
	x.RegisterGRPCService(GRPCServiceFunc(srv))
}

// grpc gateway next

type GRPCGatewayHandler interface {
	Register(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error
}

func (x *ServeMux) RegisterGRPCGatewayHandler(ctx context.Context, handler GRPCGatewayHandler) error {
	return handler.Register(ctx, x.grpcGatewayServer, x.Server.Addr, x.opts.ClientDialOpts())
}

type GRPCGatewayHandlerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error

func (f GRPCGatewayHandlerFunc) Register(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return f(ctx, mux, endpoint, opts)
}

func (x *ServeMux) RegisterGRPCGatewayHandlerFunc(ctx context.Context, handler func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error) error {
	return x.RegisterGRPCGatewayHandler(ctx, GRPCGatewayHandlerFunc(handler))
}
