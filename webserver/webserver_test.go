package webserver_test

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/limingyao/excellent-go/internal/proto"
	"github.com/limingyao/excellent-go/metrics/prometheus"
	"github.com/limingyao/excellent-go/webserver"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type echoService struct {
	pb.UnimplementedEchoServiceServer
}

func (echoService) Echo(ctx context.Context, req *pb.Message) (*pb.Message, error) {
	if rand.Intn(10) < 6 {
		return &pb.Message{Value: req.Value}, nil
	}
	return nil, errors.New(req.Value)
}

func TestNewServer(t *testing.T) {
	srv := webserver.NewServer(
		webserver.WithPort(8080),
		webserver.WithHealthz(),
		webserver.WithPProf(),
		webserver.WithReflection(),
		webserver.WithServerOptions([]grpc.ServerOption{
			grpc.MaxRecvMsgSize(webserver.ServerMaxReceiveMessageSize),
			grpc.MaxSendMsgSize(webserver.ServerMaxSendMessageSize),
		}...),
		webserver.WithDialOptions([]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}...),
		webserver.WithGatewayOptions([]runtime.ServeMuxOption{
			runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
				log.Info("WithForwardResponseOption")
				w.WriteHeader(http.StatusNotFound)
				return nil
			}),
			runtime.WithErrorHandler(func(
				ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler,
				w http.ResponseWriter, r *http.Request, err error,
			) {
				log.Info("WithErrorHandler")
				w.WriteHeader(http.StatusMethodNotAllowed)
			}),
			runtime.WithRoutingErrorHandler(func(
				ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler,
				w http.ResponseWriter, r *http.Request, httpStatus int,
			) {
				log.Info("WithRoutingErrorHandler")
				if httpStatus != http.StatusMethodNotAllowed {
					runtime.DefaultRoutingErrorHandler(ctx, mux, marshaler, w, r, httpStatus)
					return
				}

				// Use HTTPStatusError to customize the DefaultHTTPErrorHandler status code
				err := &runtime.HTTPStatusError{
					HTTPStatus: httpStatus,
					Err:        status.Error(codes.Unimplemented, http.StatusText(httpStatus)),
				}
				runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			}),
		}...),
	)
	srv.RegisterHttpHandler("/metrics", prometheus.Handler())
	// TODO add to webserver WithPrometheus ?
	// TODO support ?
	// https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/#pretty-print-json-responses-when-queried-with-pretty

	srv.RegisterGrpcServer(func(srv *grpc.Server) {
		pb.RegisterEchoServiceServer(srv, &echoService{})
	})
	srv.RegisterGatewayHandlerWithDefault(pb.RegisterEchoServiceHandlerFromEndpoint)

	if err := srv.Serve(); err != nil {
		t.Error(err)
	}
}
