package config

import (
	"os"
	"reflect"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration interface {
	Init() error
}

func Watch(filepath string, fn func() Configuration) <-chan Configuration {
	ch := make(chan Configuration, 1)

	config := fn()
	v := reflect.ValueOf(config)
	if !v.IsValid() || v.Kind() != reflect.Ptr {
		log.Fatalf("parameter %T must be a pointer", config)
	}

	viper.SetConfigFile(filepath)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Fatalf("read config %s fail", filepath)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("config %s changed", e.Name)
		config := fn()
		if err := unmarshal(config); err == nil {
			select {
			case ch <- config:
			case <-time.After(time.Second):
			}
		}
	})
	viper.WatchConfig()

	if err := unmarshal(config); err != nil {
		log.WithError(err).Fatal()
	}
	ch <- config
	return ch
}

func unmarshal(config Configuration) error {
	for _, key := range viper.AllKeys() {
		val := viper.GetString(key)
		newVal := os.ExpandEnv(val)
		if newVal != val {
			log.Infof("key: %s, replace %s -> [%s]", key, val, newVal)
			viper.Set(key, newVal)
		}
	}

	if err := viper.Unmarshal(config, DecoderOptions()...); err != nil {
		log.WithError(err).Errorf("decode config fail")
	}
	if err := config.Init(); err != nil {
		log.WithError(err).Errorf("init config fail")
	}

	log.Infof("loaded config: %+v", config)
	return nil
}
