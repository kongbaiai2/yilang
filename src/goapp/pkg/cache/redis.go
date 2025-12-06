package cache

import (
	"log"
	"strconv"
	"time"

	"github.com/kongbaiai2/yilang/goapp/internal/global"

	"github.com/gomodule/redigo/redis"
)

var (
	REDIS_NO_TIMEOUT = 0
)

type RedisConf struct {
	RedisAddr   string `yaml:"redis_addr"`
	RedisPasswd string `yaml:"redis_passwd"`
}

func newPool(redisAddr string, passwd string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     80,
		MaxActive:   12000, // max number of connections
		IdleTimeout: 6 * time.Second,
		Wait:        false,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp",
				redisAddr,
				redis.DialConnectTimeout(time.Duration(3000)*time.Millisecond),
				redis.DialReadTimeout(time.Duration(1000)*time.Millisecond),
				redis.DialWriteTimeout(time.Duration(1000)*time.Millisecond))
			if err != nil {
				return nil, redis.Error("connect time out")
			}
			if _, err := c.Do("AUTH", passwd); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
	}
}

func Redis() *redis.Pool {
	var pool = newPool(global.CONFIG.Redis.Addr, global.CONFIG.Redis.Password)

	return pool
}

func SetKeyToRedis(pool *redis.Pool, key, value string, timeout int) (err error) {
	c := pool.Get()
	defer c.Close()

	if timeout == REDIS_NO_TIMEOUT {
		_, err = c.Do("SET", key, value)
	} else {
		_, err = c.Do("SET", key, value, "EX", strconv.Itoa(timeout))
	}

	if err != nil {
		log.Printf("[ERROR] set key:%s to redis failed, err:%s", key, err)
		return err
	}

	return nil
}

func IfKeyExistInRedis(pool *redis.Pool, key string) bool {
	c := pool.Get()
	defer c.Close()

	isKeyExit, err := redis.Bool(c.Do("EXISTS", key))
	if err != nil {
		log.Printf("[ERROR] get key:%s from redis failed, err:%s", key, err)
		return false
	}

	return isKeyExit
}

func GetKeyMatchInRedis(pool *redis.Pool, key string) ([]string, error) {
	c := pool.Get()
	defer c.Close()

	keys, err := redis.Strings(c.Do("KEYS", key))
	if err != nil {
		log.Printf("[ERROR] get key:%s from redis failed, err:%s", key, err)
		return keys, err
	}

	return keys, nil
}

func DelKeyFromRedis(pool *redis.Pool, key string) error {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	if err != nil {
		log.Printf("[ERROR] delete key:%s from redis failed, err:%s", key, err)
		return err
	}

	return nil
}

func GetValueFromRedis(pool *redis.Pool, key string) ([]byte, error) {
	c := pool.Get()
	defer c.Close()

	data, err := redis.Bytes(c.Do("GET", key))

	if err != nil {
		log.Printf("[ERROR] get key:%s to redis failed, err:%s", key, err)
		return data, err
	}

	return data, nil
}
