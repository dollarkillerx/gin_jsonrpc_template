package storage

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Storage struct {
	redis *redis.Client
	db    *gorm.DB
}

func NewStorage(redisConn *redis.Client, db *gorm.DB) *Storage {
	return &Storage{redis: redisConn, db: db}
}

func (s *Storage) GetRedis() *redis.Client { return s.redis }
func (s *Storage) GetDB() *gorm.DB { return s.db }
