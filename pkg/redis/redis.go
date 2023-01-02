package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

func New(addrs []string, opts ...UniversalOptions) redis.UniversalClient {
	// 根据配置返回
	//   redis.NewUniversalClient()
	// Redis 哨兵模式连接
	//   redis.NewFailoverClient()
	// 单机连接
	//   redis.NewClient()
	// Redis 集群连接
	//   redis.NewClusterClient()
	// 连接 Redis 哨兵服务器
	//   redis.NewSentinelClient()
	// 将只读命令路由到从 Redis 服务器
	//   redis.NewFailoverClusterClient()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	defaultOpts := &redis.UniversalOptions{
		Addrs: addrs,
	}
	for _, opt := range opts {
		opt(defaultOpts)
	}

	client := redis.NewUniversalClient(defaultOpts)
	if err := client.Ping(ctx).Err(); err != nil {
		log.WithError(err).Fatalf("ping redis fail")
	}

	return client
}
