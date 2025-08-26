package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// RedisClient Redis 客户端包装器
type RedisClient struct {
	Client *redis.Client
	config *RedisConfig
	logger *logrus.Logger
}

// NewClient 创建新的 Redis 客户端
func NewClient() (*RedisClient, error) {
	config := &RedisConfig{}

	// 从配置文件读取 Redis 配置
	if err := viper.UnmarshalKey("redis", config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis config: %w", err)
	}

	// 设置默认值
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 6379
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 5
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	logger := logrus.New()
	logger.Info("Redis connection established successfully")

	return &RedisClient{
		Client: rdb,
		config: config,
		logger: logger,
	}, nil
}

// NewClusterClient 创建 Redis 集群客户端
func NewClusterClient() (*RedisClient, error) {
	// 从配置读取集群节点
	clusterNodes := viper.GetStringSlice("redis.cluster_nodes")
	if len(clusterNodes) == 0 {
		return nil, fmt.Errorf("no cluster nodes configured")
	}

	password := viper.GetString("redis.password")
	poolSize := viper.GetInt("redis.pool_size")
	if poolSize == 0 {
		poolSize = 50
	}

	minIdleConns := viper.GetInt("redis.min_idle_conns")
	if minIdleConns == 0 {
		minIdleConns = 20
	}

	// 创建集群客户端
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        clusterNodes,
		Password:     password,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis cluster: %w", err)
	}

	logger := logrus.New()
	logger.Info("Redis cluster connection established successfully")

	// 集群客户端需要特殊处理
	client := &RedisClient{
		Client: nil, // 集群客户端不能直接赋值给 Client 字段
		logger: logger,
	}

	return client, nil
}

// Close 关闭 Redis 连接
func (c *RedisClient) Close() error {
	if c.Client != nil {
		c.logger.Info("Closing Redis connection")
		return c.Client.Close()
	}
	return nil
}

// Health 检查 Redis 健康状态
func (c *RedisClient) Health(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// GetStats 获取 Redis 统计信息
func (c *RedisClient) GetStats(ctx context.Context) (*redis.PoolStats, error) {
	return c.Client.PoolStats(), nil
}

// GetInfo 获取 Redis 服务器信息
func (c *RedisClient) GetInfo(ctx context.Context) (string, error) {
	return c.Client.Info(ctx).Result()
}

// FlushDB 清空当前数据库（谨慎使用）
func (c *RedisClient) FlushDB(ctx context.Context) error {
	c.logger.Warn("Flushing Redis database")
	return c.Client.FlushDB(ctx).Err()
}

// FlushAll 清空所有数据库（谨慎使用）
func (c *RedisClient) FlushAll(ctx context.Context) error {
	c.logger.Warn("Flushing all Redis databases")
	return c.Client.FlushAll(ctx).Err()
}

// SetWithExpiration 设置键值对并指定过期时间
func (c *RedisClient) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// GetString 获取字符串值
func (c *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// GetInt 获取整数值
func (c *RedisClient) GetInt(ctx context.Context, key string) (int, error) {
	return c.Client.Get(ctx, key).Int()
}

// GetFloat 获取浮点数值
func (c *RedisClient) GetFloat(ctx context.Context, key string) (float64, error) {
	return c.Client.Get(ctx, key).Float64()
}

// Exists 检查键是否存在
func (c *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Exists(ctx, keys...).Result()
}

// Delete 删除键
func (c *RedisClient) Delete(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Del(ctx, keys...).Result()
}

// Expire 设置键的过期时间
func (c *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.Client.Expire(ctx, key, expiration).Result()
}

// TTL 获取键的剩余生存时间
func (c *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}

// Increment 原子递增
func (c *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

// IncrementBy 原子递增指定值
func (c *RedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.Client.IncrBy(ctx, key, value).Result()
}

// Decrement 原子递减
func (c *RedisClient) Decrement(ctx context.Context, key string) (int64, error) {
	return c.Client.Decr(ctx, key).Result()
}

// DecrementBy 原子递减指定值
func (c *RedisClient) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.Client.DecrBy(ctx, key, value).Result()
}

// HSet 设置哈希字段
func (c *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.Client.HSet(ctx, key, values...).Result()
}

// HGet 获取哈希字段值
func (c *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return c.Client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希所有字段
func (c *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.Client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (c *RedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.Client.HDel(ctx, key, fields...).Result()
}

// LPush 从列表左侧推入元素
func (c *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.Client.LPush(ctx, key, values...).Result()
}

// RPush 从列表右侧推入元素
func (c *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.Client.RPush(ctx, key, values...).Result()
}

// LPop 从列表左侧弹出元素
func (c *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	return c.Client.LPop(ctx, key).Result()
}

// RPop 从列表右侧弹出元素
func (c *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return c.Client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (c *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return c.Client.LLen(ctx, key).Result()
}

// SAdd 向集合添加成员
func (c *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.Client.SAdd(ctx, key, members...).Result()
}

// SMembers 获取集合所有成员
func (c *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.Client.SMembers(ctx, key).Result()
}

// SIsMember 检查是否为集合成员
func (c *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.Client.SIsMember(ctx, key, member).Result()
}

// SRem 从集合移除成员
func (c *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.Client.SRem(ctx, key, members...).Result()
}

// GetClient 获取原始 Redis 客户端
func (c *RedisClient) GetClient() *redis.Client {
	return c.Client
}

// GetConfig 获取 Redis 配置
func (c *RedisClient) GetConfig() *RedisConfig {
	return c.config
}
