package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
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
