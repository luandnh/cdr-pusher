package redis

import (
	"time"

	"github.com/go-redis/redis/v8"
)

type IRedis interface {
	GetClient() *redis.Client
	Connect() error
	Ping() error
	Set(key string, value interface{}) (string, error)
	SetTTL(key string, value interface{}, t time.Duration) (string, error)
	Get(key string) (string, error)
	IsExisted(key string) (bool, error)
	IsHExisted(list, key string) (bool, error)
	HGet(list, key string) (string, error)
	HGetAll(list string) (map[string]string, error)
	HSet(key string, values []interface{}) (int64, error)
	HMGet(key string, fields ...string) ([]interface{}, error)
	HMSet(key string, values ...interface{}) error
	HMDel(key string, fields ...string) error
	FLUSHALL() interface{}
	Del(key []string) error
	HDel(key string, fields ...string) error
	GetKeysPattern(pattern string) ([]string, error)
}

var Redis IRedis
