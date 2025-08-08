package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// PostgresConfiguration configuration for Postgres database connection
// Mimics the layout from the reference project

type PostgresConfiguration struct {
	Host     string `mapstructure:"Host"`
	Port     int    `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
	DBName   string `mapstructure:"DBName"`
	SSLMode  bool   `mapstructure:"SSLMode"`
	TimeZone string `mapstructure:"TimeZone"`
}

// ServiceConfiguration configuration for service

type ServiceConfiguration struct {
	Port  string `mapstructure:"Port"`
	Debug bool   `mapstructure:"Debug"`
}

// RedisConfiguration configuration for Redis

type RedisConfiguration struct {
	Addr     string `mapstructure:"Addr"`
	Db       int    `mapstructure:"Db"`
	Password string `mapstructure:"Password"`
}

// LoggerConfig configuration for logger

type LoggerConfig struct {
	Filename string `mapstructure:"Filename"`
	MaxSize  int    `mapstructure:"MaxSize"` // MB
}

// InitConfiguration reads configuration from files and env vars
// - configName without extension (e.g., "config")
// - configPaths are searched in order (e.g., ./, ./configs/)

func InitConfiguration(configName string, configPaths []string, config interface{}) error {
	vp := viper.New()
	vp.SetConfigName(configName)
	vp.AutomaticEnv()

	for _, p := range configPaths {
		vp.AddConfigPath(p)
	}

	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return errors.WithStack(err)
		}
	}

	if err := vp.Unmarshal(config); err != nil {
		return errors.WithStack(err)
	}

	for _, key := range vp.AllKeys() {
		if err := vp.BindEnv(key); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
