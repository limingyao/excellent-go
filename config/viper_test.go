package config_test

import (
	"testing"
	"time"

	"github.com/limingyao/excellent-go/config"
)

type testConfig struct {
	App    string        `yaml:"app"`
	MaxAge time.Duration `yaml:"max_age"`
	Addrs  []string      `yaml:"addrs"`
}

func (c testConfig) Init() error {
	return nil
}

func TestInit(t *testing.T) {
	go func() {
		for cfg := range config.Watch("./config.yaml", func() config.Configuration {
			return &testConfig{}
		}) {
			t.Log(cfg)
		}
	}()
	time.Sleep(20 * time.Second)
}
