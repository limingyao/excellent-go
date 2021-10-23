package mux

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/limingyao/excellent-go/test"
	"github.com/limingyao/excellent-go/transport/http"
	"google.golang.org/grpc"
	"log"
	"testing"
)

type server struct {
	test.UnimplementedGreeterServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) SayHello(ctx context.Context, in *test.HelloRequest) (*test.HelloReply, error) {
	return &test.HelloReply{Message: in.Name + " world"}, nil
}

func TestNewServeMux(t *testing.T) {
	srv := NewServeMux("0.0.0.0:8020")
	srv.RegisterGRPCServiceFunc(func(srv *grpc.Server) {
		test.RegisterGreeterServer(srv, NewServer())
	})
	err := srv.RegisterGRPCGatewayHandlerFunc(context.Background(),
		func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
			return test.RegisterGreeterHandlerFromEndpoint(ctx, mux, endpoint, []grpc.DialOption{grpc.WithInsecure()})
		})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(srv.ListenAndServe())
}

func TestNewServeMux_Client(t *testing.T) {
	httpClient, _ := http.New()

	req := &test.HelloRequest{Name: "hello world"}
	rsp := &test.HelloReply{}

	// http + json
	_, code, err := httpClient.JSONPost(context.Background(), "http://localhost:8020/v1/example/echo", nil, req, rsp)
	log.Println(code, err, rsp)

	// http + proto
	_, code, err = httpClient.ProtoPost(context.Background(), "http://localhost:8020/v1/example/echo",
		map[string]string{"Content-Type": "application/proto"}, req, rsp)
	log.Println(code, err, rsp)

	// grpc
	conn, err := grpc.Dial("127.0.0.1:8020", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("grpc.Dial err: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	client := test.NewGreeterClient(conn)
	rsp, err = client.SayHello(context.Background(), req)
	log.Println(rsp, err)
}
