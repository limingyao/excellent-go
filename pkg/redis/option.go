package redis

import (
	"time"

	"github.com/go-redis/redis/v8"
)

type UniversalOptions func(*redis.UniversalOptions)

func WithDB(db int) UniversalOptions {
	return func(opt *redis.UniversalOptions) {
		opt.DB = db
	}
}

func WithMasterName(masterName string) UniversalOptions {
	return func(opt *redis.UniversalOptions) {
		opt.MasterName = masterName
	}
}

func WithPassword(password string) UniversalOptions {
	return func(opt *redis.UniversalOptions) {
		opt.Password = password
		opt.SentinelPassword = password
	}
}

func WithAuth(username, password string) UniversalOptions {
	return func(opt *redis.UniversalOptions) {
		opt.Username = username
		opt.Password = password
		opt.SentinelUsername = username
		opt.SentinelPassword = password
	}
}

func WithPoolSize(poolSize int) UniversalOptions {
	return func(opt *redis.UniversalOptions) {
		opt.PoolSize = poolSize
	}
}

func WithMinIdleConns(idleConns int) UniversalOptions {
	return func(opt *redis.UniversalOptions) {
		opt.MinIdleConns = idleConns
	}
}

func WithMaxConnAge(connAge time.Duration) UniversalOptions {
	return func(opt *redis.UniversalOptions) {
		opt.MaxConnAge = connAge
	}
}
