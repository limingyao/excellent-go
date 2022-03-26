package docker

import (
	"encoding/base64"
	"encoding/json"

	"github.com/docker/docker/api/types"
)

type options struct {
	registryAuth string
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
		authConfig := types.AuthConfig{
			Username: username,
			Password: password,
		}
		encodedJSON, _ := json.Marshal(authConfig)
		o.registryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
	})
}
