package webserver_test

import (
	"testing"

	"github.com/limingyao/excellent-go/metrics/prometheus"
	"github.com/limingyao/excellent-go/webserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewServer(t *testing.T) {
	srv := webserver.NewServer(
		webserver.WithHealthz(),
		webserver.WithPProf(),
		webserver.WithDialOptions([]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}...),
	)
	srv.RegisterHttpHandler("/metrics", prometheus.Handler())
	if err := srv.Serve(); err != nil {
		t.Error(err)
	}
}
