package config_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/limingyao/excellent-go/config"
)

type hello struct {
	Name string `yaml:"name" json:"name"`
}

func TestLoad(t *testing.T) {
	if err := ioutil.WriteFile("config.yaml", []byte("name: hello"), 0644); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := config.Load("config.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	if err := c.Reload(); err != nil {
		t.Error(err)
		return
	}

	t.Log(string(c.Bytes()))

	h := &hello{}
	if err := c.Unmarshal(h); err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", h)

	go func() {
		for i := 0; ; i++ {
			select {
			case <-time.After(time.Second):
				if err := ioutil.WriteFile("config.yaml", []byte(fmt.Sprintf("name: hello-%d", i)), 0644); err != nil {
					t.Error(err)
					continue
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for w := range c.Watch(ctx) {
		t.Log(string(w.Data))
	}
}
