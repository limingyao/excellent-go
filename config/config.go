package config

import (
	"errors"

	"golang.org/x/net/context"
)

type EventType uint8

const (
	EventTypeNil EventType = 0
	EventTypePut EventType = 1
	EventTypeDel EventType = 2
)

type WatchResponse struct {
	Data  []byte
	Event EventType
}

type WatchChan <-chan WatchResponse

type Config interface {
	Load() error
	Reload() error
	Bytes() []byte
	Unmarshal(interface{}) error
	Watch(ctx context.Context) WatchChan
}

type config struct {
	p       DataProvider
	key     string
	decoder Codec

	rawData       []byte
	unmarshalData interface{}
}

func (c *config) Load() error {
	data, err := c.p.Read(c.key)
	if err != nil {
		return err
	}

	c.rawData = data
	c.unmarshalData = map[string]interface{}{}
	if err := c.decoder.Unmarshal(c.rawData, &c.unmarshalData); err != nil {
		return err
	}

	return nil
}

func (c *config) Reload() error {
	data, err := c.p.Read(c.key)
	if err != nil {
		return err
	}

	unmarshalData := map[string]interface{}{}
	if err := c.decoder.Unmarshal(data, &unmarshalData); err != nil {
		return err
	}

	c.rawData = data
	c.unmarshalData = unmarshalData

	return nil
}

func (c config) Bytes() []byte {
	return c.rawData
}

func (c config) Unmarshal(out interface{}) error {
	return c.decoder.Unmarshal(c.rawData, out)
}

func (c config) Watch(ctx context.Context) WatchChan {
	sendCh := make(chan WatchResponse, 1)

	go func() {
		for ch := range c.p.Watch(ctx, c.key) {
			sendCh <- WatchResponse{
				Data:  ch.Data,
				Event: EventTypePut,
			}
		}
		close(sendCh)
	}()

	return sendCh
}

type ConfigLoader struct {
}

func (c *ConfigLoader) Load(key string, opts ...Option) (Config, error) {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	cfg := &config{
		p:       GetProvider(defaultOpts.providerName),
		key:     key,
		decoder: GetCodec(defaultOptions.codecName),
	}

	if cfg.decoder == nil {
		return nil, errors.New("decoder not exist")
	}
	if cfg.p == nil {
		return nil, errors.New("provider not exist")
	}

	if err := cfg.Load(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func newConfigLoader() *ConfigLoader {
	return &ConfigLoader{}
}

var defaultConfigLoader = newConfigLoader()

func Load(key string, opts ...Option) (Config, error) {
	return defaultConfigLoader.Load(key, opts...)
}
