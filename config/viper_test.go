package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/limingyao/excellent-go/config"
)

type testConfig struct {
	App    string        `yaml:"app"`
	MaxAge time.Duration `yaml:"max_age"`
	Addrs  []string      `yaml:"addrs"`
	Scores []int         `yaml:"scores"`
}

func (c testConfig) Init() error {
	return nil
}

var configBytes = []byte(`
app: test
max_age: 3s # time.Duration
addrs: shanghai,beijing # []string
scores: 1,2,3
`)

func TestInit(t *testing.T) {
	if err := os.WriteFile("config.yaml", configBytes, 0644); err != nil {
		t.Error(err)
		return
	}
	go func() {
		for cfg := range config.Watch("./config.yaml", func() config.Configuration {
			return &testConfig{}
		}) {
			t.Log(cfg)
		}
	}()
	time.Sleep(20 * time.Second)
}
