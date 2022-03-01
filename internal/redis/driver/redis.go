package redis

import (
	"cdr-pusher/common/log"
	IR "cdr-pusher/internal/redis"
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Redis struct {
	Client *redis.Client
	config Config
}

type Config struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	PoolTimeout  int
	IdleTimeout  int
	ReadTimeout  int
	WriteTimeout int
}

func NewRedis(config Config) (IR.IRedis, error) {
	r := &Redis{
		config: config,
	}
	if err := r.Connect(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Redis) GetClient() *redis.Client {
	return r.Client
}

func (r *Redis) Connect() error {
	Client := redis.NewClient(&redis.Options{
		Addr:         r.config.Addr,
		Password:     r.config.Password,
		DB:           r.config.DB,
		PoolSize:     r.config.PoolSize,
		PoolTimeout:  time.Duration(r.config.PoolTimeout) * time.Second,
		IdleTimeout:  time.Duration(r.config.IdleTimeout) * time.Second,
		ReadTimeout:  time.Duration(r.config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(r.config.WriteTimeout) * time.Second,
	})
	str, err := Client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Redis", "Connect", err.Error())
		return err
	}
	log.Info("Redis", "Connect", str)
	r.Client = Client
	return nil
}

func (r *Redis) Ping() error {
	_, err := r.Client.Ping(ctx).Result()
	if err != nil {
		log.Error("Redis", "Ping", err.Error())
	}
	return err
}

func (r *Redis) Set(key string, value interface{}) (string, error) {
	ret, err := r.Client.Set(ctx, key, value, 0).Result()
	if err != nil {
		log.Error("Redis", "Set", err.Error())
	}
	return ret, err
}

//Set - Set a value with key to Redis DB
func (r *Redis) SetTTL(key string, value interface{}, t time.Duration) (string, error) {
	ret, err := r.Client.Set(ctx, key, value, t).Result()
	if err != nil {
		log.Error("Redis", "SetTTL", err.Error())
	}
	return ret, err
}

func (r *Redis) Get(key string) (string, error) {
	ret, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		log.Error("Redis", "Get", err.Error())
	}
	return ret, err
}

func (r *Redis) IsExisted(key string) (bool, error) {
	res, err := r.Client.Exists(ctx, key).Result()
	if res == 0 || err != nil {
		if err != nil {
			log.Error("Redis", "IsExisted", err.Error())
		}
		return false, err
	}
	return true, nil
}

func (r *Redis) IsHExisted(list, key string) (bool, error) {
	res, err := r.Client.HExists(ctx, list, key).Result()
	if res == false || err != nil {
		if err != nil {
			log.Error("Redis", "IsExisted", err.Error())
		}
		return false, err
	}
	return true, nil
}

func (r *Redis) HGet(list, key string) (string, error) {
	ret, err := r.Client.HGet(ctx, list, key).Result()
	if err != nil {
		log.Error("Redis", "Get", err.Error())
	}
	return ret, err
}

func (r *Redis) HGetAll(list string) (map[string]string, error) {
	ret, err := r.Client.HGetAll(ctx, list).Result()
	if err != nil {
		log.Error("Redis", "HGetAll", err.Error())
	}
	return ret, err
}

func (r *Redis) HSet(key string, values []interface{}) (int64, error) {
	ret, err := r.Client.HSet(ctx, key, values...).Result()
	if err != nil {
		log.Error("Redis", "HSet", err.Error())
	}
	return ret, err
}

func (r *Redis) Del(key []string) error {
	err := r.Client.Del(ctx, key...).Err()
	if err != nil {
		log.Error("Redis", "Del", err.Error())
	}
	return err
}

func (r *Redis) HMSet(key string, values ...interface{}) error {
	ret, err := r.Client.HMSet(ctx, key, values...).Result()
	if err != nil {
		log.Error("Redis", "HMSet", err.Error())
		return err
	}
	if !ret {
		err = errors.New("HashMap Set failed")
		log.Error("Redis", "HMSet", err.Error())
	}
	return err
}

func (r *Redis) HMDel(key string, fields ...string) error {
	err := r.Client.HDel(ctx, key, fields...).Err()
	if err != nil {
		log.Error("Redis", "HMDel", err.Error())
	}
	return err
}

func (r *Redis) FLUSHALL() interface{} {
	ret := r.Client.FlushAll(ctx)
	return ret
}

func (r *Redis) HMGet(key string, fields ...string) ([]interface{}, error) {
	ret, err := r.Client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		log.Error("Redis", "HMGet", err.Error())
		return ret, err
	}
	return ret, err
}

func (r *Redis) HDel(key string, fields ...string) error {
	err := r.Client.HDel(ctx, key, fields...).Err()
	if err != nil {
		log.Error("Redis", "HDel", err.Error())
	}
	return err
}

func (r *Redis) GetKeysPattern(pattern string) ([]string, error) {
	ret, err := r.Client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Error("Redis", "HMGet", err.Error())
		return ret, err
	}
	return ret, err
}
