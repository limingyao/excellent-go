package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// DecoderOptions 配置文件解码参数
func DecoderOptions() []viper.DecoderConfigOption {
	return []viper.DecoderConfigOption{func(config *mapstructure.DecoderConfig) {
		config.TagName = "yaml"
	}}
}
