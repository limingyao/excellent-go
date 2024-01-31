package webserver

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	ServerMaxReceiveMessageSize = 4 * 1024 * 1024 * 1024 // 4GB
	ServerMaxSendMessageSize    = math.MaxInt32

	ClientMaxReceiveMessageSize = 4 * 1024 * 1024 * 1024 // 4GB
	ClientMaxSendMessageSize    = math.MaxInt32
)

type Webserver struct {
	ctx    context.Context
	cancel context.CancelFunc

	ip         string
	port       int
	httpMux    *http.ServeMux
	gatewayMux *runtime.ServeMux
	grpcSrv    *grpc.Server

	gatewayOptions       []runtime.ServeMuxOption // gateway option
	dialOptions          []grpc.DialOption        // gateway dial grpc option
	serverOptions        []grpc.ServerOption      // grpc option
	handlersFromEndpoint []HandlerFromEndpoint    // gateway handlers

	enableHealthz    bool
	healthzPath      string
	enableReflection bool
	enablePProf      bool
	pprofPath        string
}

func NewServer(opts ...ServerOption) *Webserver {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Webserver{
		ctx:         ctx,
		cancel:      cancel,
		healthzPath: "/healthz",
		pprofPath:   "/debug/pprof",
	}
	for _, opt := range opts {
		opt(s)
	}
	s.httpMux = http.NewServeMux()
	s.gatewayMux = runtime.NewServeMux(s.gatewayOptions...)
	s.grpcSrv = grpc.NewServer(s.serverOptions...)
	return s
}

func (s *Webserver) Serve() error {
	// httpMux 执行最长前缀匹配，注册路径最后必须以/结尾才会触发，否则都交由/路径处理
	// 所有未匹配到的路径最终都会交给/路径处理
	s.httpMux.Handle("/", s.gatewayMux)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			s.grpcSrv.ServeHTTP(w, r)
		} else {
			s.httpMux.ServeHTTP(w, r)
		}
	})

	// 监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		log.WithError(err).Fatal("listen failed")
	}
	defer s.cancel()

	// 获取分配的端口
	if s.port == 0 {
		items := strings.Split(lis.Addr().String(), ":")
		if len(items) > 1 {
			s.port, err = strconv.Atoi(items[len(items)-1])
			if err != nil {
				log.WithError(err).Errorf("convert %s to int fail", items[len(items)-1])
			}
		}
	}
	log.Infof("webserver started on %s", lis.Addr().String())

	if s.enableHealthz {
		if err := s.registerHealthServer(); err != nil {
			log.WithError(err).Error()
		}
	}
	if s.enableReflection {
		s.registerReflectionServer()
	}
	if s.enablePProf {
		s.registerPProf()
	}
	s.registerGatewayHandler()

	return http.Serve(lis, h2c.NewHandler(handler, &http2.Server{}))
}

func (s *Webserver) ServeTLS() error {
	// TODO ...
	return nil
}

func (s *Webserver) registerHealthServer() error {
	registerHealthServer(s.grpcSrv)

	cc, err := grpc.Dial(fmt.Sprintf("passthrough:///%s:%d", s.ip, s.port), s.dialOptions...)
	if err != nil {
		log.WithError(err).Errorf("dail fail")
		return err
	}

	runtime.WithHealthEndpointAt(newHealthClient(cc), s.healthzPath)(s.gatewayMux)

	SetServerStatus(ServingStatus)

	return nil
}

func (s *Webserver) registerReflectionServer() {
	s.RegisterGrpcServer(func(srv *grpc.Server) {
		reflection.Register(srv)
	})
}

func (s *Webserver) registerPProf() {
	s.httpMux.HandleFunc(fmt.Sprintf("%s/", s.pprofPath), pprof.Index)
	s.httpMux.HandleFunc(fmt.Sprintf("%s/cmdline", s.pprofPath), pprof.Cmdline)
	s.httpMux.HandleFunc(fmt.Sprintf("%s/profile", s.pprofPath), pprof.Profile)
	s.httpMux.HandleFunc(fmt.Sprintf("%s/symbol", s.pprofPath), pprof.Symbol)
	s.httpMux.HandleFunc(fmt.Sprintf("%s/trace", s.pprofPath), pprof.Trace)
}

func (s *Webserver) RegisterHttpHandler(pattern string, handler http.Handler) {
	s.httpMux.Handle(pattern, handler)
}

type HandlerFromEndpoint func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error

func (s *Webserver) RegisterGatewayHandlerFromEndpoint(
	endpoint string, opts []grpc.DialOption, handlerFromEndpoint HandlerFromEndpoint,
) {
	if err := handlerFromEndpoint(s.ctx, s.gatewayMux, endpoint, opts); err != nil {
		log.Fatal(err)
	}
}

func (s *Webserver) RegisterGatewayHandlerWithDefault(handlerFromEndpoint HandlerFromEndpoint) {
	s.handlersFromEndpoint = append(s.handlersFromEndpoint, handlerFromEndpoint)
}

func (s *Webserver) registerGatewayHandler() {
	endpoint := fmt.Sprintf("passthrough:///%s:%d", s.ip, s.port)
	for i := range s.handlersFromEndpoint {
		s.RegisterGatewayHandlerFromEndpoint(endpoint, s.dialOptions, s.handlersFromEndpoint[i])
	}
}

func (s *Webserver) RegisterGrpcServer(fn func(srv *grpc.Server)) {
	fn(s.grpcSrv)
}

func (s *Webserver) HttpMux() *http.ServeMux {
	return s.httpMux
}

func (s *Webserver) GatewayMux() *runtime.ServeMux {
	return s.gatewayMux
}

func (s *Webserver) GrpcServer() *grpc.Server {
	return s.grpcSrv
}

func (s *Webserver) Stop() {
	s.cancel()
	s.grpcSrv.Stop()
}
