package docker

import (
	"encoding/base64"
	"encoding/json"

	"github.com/docker/docker/api/types"
)

type options struct {
	// registryAuth for docker image pull
	registryAuth string

	// compress for docker build
	compress bool
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

func WithRegistryAuth(username, password string) Option {
	return newFuncOption(func(o *options) {
		if username == "" && password == "" {
			return
		}
		authConfig := types.AuthConfig{
			Username: username,
			Password: password,
		}
		encodedJSON, _ := json.Marshal(authConfig)
		o.registryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
	})
}

func WithCompress() Option {
	return newFuncOption(func(o *options) {
		o.compress = true
	})
}
