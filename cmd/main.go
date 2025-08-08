package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/google/feitian/internal/conf"
	"github.com/google/feitian/internal/server"
	"github.com/google/feitian/internal/storage"
	"github.com/google/feitian/pkg/common/client"
	"github.com/google/feitian/pkg/common/config"
	logs "github.com/google/feitian/pkg/common/log"
	"github.com/rs/zerolog/log"
)

var configFilename string
var configDirs string

func init() {
	const (
		defaultConfigFilename = "config"
		defaultConfigDirs     = "./,./configs/"
	)
	flag.StringVar(&configFilename, "c", defaultConfigFilename, "Name of the config file, without extension")
	flag.StringVar(&configFilename, "dev_config", defaultConfigFilename, "Name of the config file, without extension")
	flag.StringVar(&configDirs, "cPath", defaultConfigDirs, "Directories to search for config file, separated by ','")
}

func main() {
	flag.Parse()

	var appConfig conf.Config
	if err := config.InitConfiguration(configFilename, strings.Split(configDirs, ","), &appConfig); err != nil {
		panic(err)
	}
	if b, err := json.MarshalIndent(appConfig, "", "  "); err == nil {
		fmt.Println(string(b))
	}
	fmt.Println("Config loaded successfully!")

	// Logger
	logs.InitLog(appConfig.LoggerConfiguration)

	// Postgres
	pg, err := client.PostgresClient(appConfig.PostgresConfiguration, nil)
	if err != nil {
		log.Error().Msg("Failed to connect to postgres")
		panic(err)
	}

	// Redis
	rc, err := client.RedisClient(appConfig.RedisConfiguration)
	if err != nil {
		log.Error().Msg("Failed to connect to redis")
		panic(err)
	}
	ping := rc.Ping(context.Background())
	if err := ping.Err(); err != nil {
		log.Error().Msg("Redis ping failed")
		panic(err)
	}
	log.Info().Msgf("Redis ping: %s", ping.Val())

	st := storage.NewStorage(rc, pg)
	log.Info().Msg("Storage initialized")

	s := server.NewServer(st, appConfig)
	if err := s.Run(); err != nil {
		log.Error().Msgf("Failed to start server %s", err)
	}
}
