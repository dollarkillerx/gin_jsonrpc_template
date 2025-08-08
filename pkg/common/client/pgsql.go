package client

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/feitian/pkg/common/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func PostgresClient(conf config.PostgresConfiguration, gormConfig *gorm.Config) (*gorm.DB, error) {
	if conf.TimeZone == "" {
		conf.TimeZone = "Asia/Tokyo"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d TimeZone=%s", conf.Host, conf.User, conf.Password, conf.DBName, conf.Port, conf.TimeZone)
	if !conf.SSLMode {
		dsn += " sslmode=disable"
	}
	if gormConfig == nil {
		gormConfig = &gorm.Config{
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  logger.Info,
					IgnoreRecordNotFoundError: true,
					Colorful:                  true,
				},
			),
		}
	}
	return gorm.Open(postgres.Open(dsn), gormConfig)
}
