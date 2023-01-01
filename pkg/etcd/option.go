package etcd

import (
	"crypto/tls"
	"log"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
)

type options struct {
	dialTimeout time.Duration

	username string
	password string

	tls *tls.Config
}

var (
	defaultOptions = options{
		dialTimeout: 5 * time.Second,
	}
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

func WithAuth(username, password string) Option {
	return newFuncOption(func(o *options) {
		o.username = username
		o.password = password
	})
}

func WithSSL(keyFile, certFile, trustedCAFile string) Option {
	return newFuncOption(func(o *options) {
		tlsInfo := transport.TLSInfo{
			CertFile:      certFile,
			KeyFile:       keyFile,
			TrustedCAFile: trustedCAFile,
		}

		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			log.Fatal(err)
		}

		o.tls = tlsConfig
	})
}
