package conf

import "github.com/google/feitian/pkg/common/config"

type Config struct {
	ServiceConfiguration  config.ServiceConfiguration  `mapstructure:"ServiceConfiguration"`
	PostgresConfiguration config.PostgresConfiguration `mapstructure:"PostgresConfiguration"`
	RedisConfiguration    config.RedisConfiguration    `mapstructure:"RedisConfiguration"`
	LoggerConfiguration   config.LoggerConfig          `mapstructure:"LoggerConfiguration"`
}
