package config

import (
	"bytes"
	"os"
	"reflect"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration interface {
	Init() error
}

func Watch(filepath string, initializer func() Configuration, opts ...Option) <-chan Configuration {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	viper.SetConfigFile(filepath)
	viper.SetConfigType(defaultOpts.tagName)
	if defaultOpts.automaticEnv {
		viper.AutomaticEnv()
	}
	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Fatalf("read config %s fail", filepath)
	}

	configs := make(chan Configuration, 1)
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("config %s changed", e.Name)
		config := initializer()
		if err := unmarshal(config, defaultOpts); err == nil {
			select {
			case configs <- config:
			case <-time.After(time.Second):
			}
		}
	})
	viper.WatchConfig()

	config := initializer()
	if err := unmarshal(config, defaultOpts); err != nil {
		log.WithError(err).Fatal()
	}
	configs <- config

	return configs
}

func unmarshal(config Configuration, opt options) error {
	v := reflect.ValueOf(config)
	if !v.IsValid() || v.Kind() != reflect.Ptr {
		log.Fatalf("parameter %T must be a pointer", config)
	}

	var decoderOpts []viper.DecoderConfigOption
	decoderOpts = append(decoderOpts, func(config *mapstructure.DecoderConfig) {
		config.TagName = opt.tagName
	})

	for _, key := range viper.AllKeys() {
		val := viper.GetString(key)
		newVal := os.ExpandEnv(val)
		if newVal != val {
			log.Infof("key: %s, replace %s -> [%s]", key, val, newVal)
			viper.Set(key, newVal)
		}
	}

	if err := viper.Unmarshal(config, decoderOpts...); err != nil {
		log.WithError(err).Errorf("decode config fail")
	}
	if err := config.Init(); err != nil {
		log.WithError(err).Errorf("init config fail")
	}

	log.Infof("loaded config: %+v", config)
	return nil
}

func Unmarshal(buffer []byte, config Configuration, opts ...Option) error {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	viper.SetConfigType(defaultOpts.tagName)
	if defaultOpts.automaticEnv {
		viper.AutomaticEnv()
	}
	if err := viper.ReadConfig(bytes.NewReader(buffer)); err != nil {
		log.WithError(err).Fatal("read config buffer fail")
	}
	if err := unmarshal(config, defaultOpts); err != nil {
		log.WithError(err).Fatal()
	}

	return nil
}

func UnmarshalFile(filepath string, config Configuration, opts ...Option) error {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	viper.SetConfigFile(filepath)
	viper.SetConfigType(defaultOpts.tagName)
	if defaultOpts.automaticEnv {
		viper.AutomaticEnv()
	}
	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Fatalf("read config %s fail", filepath)
	}
	if err := unmarshal(config, defaultOpts); err != nil {
		log.WithError(err).Fatal()
	}

	return nil
}
