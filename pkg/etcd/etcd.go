package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func New(addrs []string, opts ...Option) (*clientv3.Client, error) {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	return clientv3.New(clientv3.Config{
		Endpoints:   addrs,
		DialTimeout: defaultOpts.dialTimeout,
		Username:    defaultOpts.username,
		Password:    defaultOpts.password,
		TLS:         defaultOpts.tls,
	})
}

func Elect(ctx context.Context, client *clientv3.Client, pfx, val string, processor func(context.Context)) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	session, err := concurrency.NewSession(client, concurrency.WithTTL(10))
	if err != nil {
		return err
	}
	defer session.Close()

	election := concurrency.NewElection(session, pfx)
	if err := election.Campaign(ctx, val); err != nil {
		return err
	}
	// 这里 ctx 不能使用外部传入的，避免不能 resign
	defer election.Resign(context.TODO())

	go func() {
		select {
		case <-session.Done():
			cancel() // session 过期，取消 ctx
		}
	}()

	// 需要在 processor 中处理 ctx
	//  select {
	//  case <-ctx.Done():
	//    return
	//  default:
	//	  // do something
	//  }
	processor(ctx)

	return nil
}
