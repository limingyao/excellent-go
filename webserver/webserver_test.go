package webserver_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/limingyao/excellent-go/encoding"
	pb "github.com/limingyao/excellent-go/internal/proto"
	"github.com/limingyao/excellent-go/webserver"
	"github.com/limingyao/excellent-go/webserver/interceptors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type echoService struct {
	pb.UnimplementedEchoServiceServer
}

func (echoService) Echo(ctx context.Context, req *pb.Message) (*pb.Message, error) {
	if rand.Intn(10) < 6 {
		return &pb.Message{Value: fmt.Sprintf("[%s]", req.Value)}, nil
	}
	return nil, status.Errorf(codes.InvalidArgument, "[%s]", req.Value)
}

type protoMarshaller struct {
	runtime.ProtoMarshaller
}

// ContentType always returns "".
func (*protoMarshaller) ContentType(interface{}) string {
	return encoding.MIMEJSON
}

func (*protoMarshaller) Marshal(value interface{}) ([]byte, error) {
	message, _ := value.(proto.Message)
	return protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(message)
}

func TestNewServer(t *testing.T) {
	srv := webserver.NewServer(
		webserver.WithPort(8080),
		webserver.WithHealthz(),
		webserver.WithPProf(),
		webserver.WithReflection(),
		webserver.WithPrometheus(),
		webserver.WithServerOptions([]grpc.ServerOption{
			grpc.MaxRecvMsgSize(webserver.ServerMaxReceiveMessageSize),
			grpc.MaxSendMsgSize(webserver.ServerMaxSendMessageSize),
			grpc.ChainUnaryInterceptor(
				interceptors.UnaryServerInterceptorOfRecovery(),
				interceptors.UnaryServerInterceptorOfContext(),
				interceptors.UnaryServerInterceptorOfSessionId(),
				interceptors.UnaryServerInterceptorOfDebug(),
			),
		}...),
		webserver.WithDialOptions([]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(webserver.ClientMaxReceiveMessageSize),
				grpc.MaxCallSendMsgSize(webserver.ClientMaxSendMessageSize),
			),
			grpc.WithChainUnaryInterceptor(
				interceptors.UnaryClientInterceptorOfMetadata(),
				interceptors.UnaryClientInterceptorOfSessionId(),
				interceptors.UnaryClientInterceptorOfDebug(),
			),
		}...),
		webserver.WithGatewayOptions([]runtime.ServeMuxOption{
			// request: protobuf, response: json
			runtime.WithMarshalerOption(encoding.MIMEPROTOBUF, &protoMarshaller{}),
			// request: json, response: json
			runtime.WithMarshalerOption(encoding.MIMEJSON, &runtime.JSONPb{
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames: true,
				},
			}),
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
				MarshalOptions: protojson.MarshalOptions{
					EmitUnpopulated: true,
					UseProtoNames:   true,
					Multiline:       true,
					Indent:          "  ",
				},
			}),
			// headers to metadata
			runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
				switch key {
				case "X-User-Id":
					return key, true
				default:
					return runtime.DefaultHeaderMatcher(key)
				}
			}),
			// metadata to headers
			runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
				switch key {
				case "X-User-Id":
					return key, true
				default:
					return runtime.DefaultHeaderMatcher(key)
				}
			}),
			// mutate response, headers
			runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
				w.Header().Set("X-Test", "true")
				if v, ok := resp.(*pb.Message); ok {
					v.Value = fmt.Sprintf("{%s}", v.Value)
				}
				return nil
			}),
			// error handler
			runtime.WithErrorHandler(func(
				ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error,
			) {
				log.WithError(err).Error()
				runtime.DefaultHTTPErrorHandler(ctx, mux, m, w, r, err)
			}),
			// routing error handler
			runtime.WithRoutingErrorHandler(func(
				ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, httpStatus int,
			) {
				if httpStatus != http.StatusMethodNotAllowed {
					runtime.DefaultRoutingErrorHandler(ctx, mux, m, w, r, httpStatus)
					return
				}

				err := &runtime.HTTPStatusError{
					HTTPStatus: httpStatus,
					Err:        status.Error(codes.Unimplemented, http.StatusText(httpStatus)),
				}
				runtime.DefaultHTTPErrorHandler(ctx, mux, m, w, r, err)
			}),
		}...),
	)

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
