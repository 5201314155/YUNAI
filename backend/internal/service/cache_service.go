package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
)

// CacheService 缓存管理服务接口
type CacheService interface {
	// 基础缓存操作
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	SetMultiple(ctx context.Context, items map[string]interface{}, expiration time.Duration) error
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	DeleteMultiple(ctx context.Context, keys []string) error

	// 缓存策略管理
	SetWithStrategy(ctx context.Context, key string, value interface{}, strategy *CacheStrategy) error
	GetCacheStrategy(ctx context.Context, pattern string) (*CacheStrategy, error)
	UpdateCacheStrategy(ctx context.Context, strategy *CacheStrategy) error

	// 缓存预热
	WarmupCache(ctx context.Context, patterns []string) error
	ScheduleWarmup(ctx context.Context, schedule *WarmupSchedule) error

	// 缓存失效
	InvalidatePattern(ctx context.Context, pattern string) error
	InvalidateByTags(ctx context.Context, tags []string) error
	SetTags(ctx context.Context, key string, tags []string) error

	// 分布式缓存同步
	SyncCache(ctx context.Context, key string) error
	BroadcastInvalidation(ctx context.Context, key string) error

	// 监控统计
	GetCacheStats(ctx context.Context) (*CacheStats, error)
	GetHitRatio(ctx context.Context, pattern string) (float64, error)
	GetCacheSize(ctx context.Context) (int64, error)

	// 缓存清理
	CleanupExpired(ctx context.Context) error
	FlushCache(ctx context.Context, pattern string) error
}

// cacheService 缓存管理服务实现
type cacheService struct {
	redis  *redis.Client
	logger *logrus.Logger

	// 缓存配置
	config *CacheConfig

	// 策略管理
	strategies map[string]*CacheStrategy
	strategyMu sync.RWMutex

	// 统计信息
	stats *CacheStatsCollector
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DefaultTTL      time.Duration `json:"default_ttl"`
	MaxKeyLength    int           `json:"max_key_length"`
	KeyPrefix       string        `json:"key_prefix"`
	EnableStats     bool          `json:"enable_stats"`
	EnableWarmup    bool          `json:"enable_warmup"`
	WarmupBatchSize int           `json:"warmup_batch_size"`
	SyncEnabled     bool          `json:"sync_enabled"`
	SyncChannel     string        `json:"sync_channel"`
}

// CacheStrategy 缓存策略
type CacheStrategy struct {
	Pattern     string        `json:"pattern"`
	TTL         time.Duration `json:"ttl"`
	MaxSize     int64         `json:"max_size"`
	EvictPolicy string        `json:"evict_policy"` // LRU, LFU, FIFO
	Tags        []string      `json:"tags"`
	WarmupFunc  string        `json:"warmup_func"`
	Enabled     bool          `json:"enabled"`
}

// WarmupSchedule 预热调度
type WarmupSchedule struct {
	ID         uuid.UUID  `json:"id"`
	Pattern    string     `json:"pattern"`
	CronExpr   string     `json:"cron_expr"`
	WarmupFunc string     `json:"warmup_func"`
	Enabled    bool       `json:"enabled"`
	LastRun    *time.Time `json:"last_run"`
	NextRun    *time.Time `json:"next_run"`
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalKeys     int64   `json:"total_keys"`
	TotalMemory   int64   `json:"total_memory"`
	HitCount      int64   `json:"hit_count"`
	MissCount     int64   `json:"miss_count"`
	HitRatio      float64 `json:"hit_ratio"`
	EvictionCount int64   `json:"eviction_count"`
	ExpiredCount  int64   `json:"expired_count"`
}

// CacheStatsCollector 缓存统计收集器
type CacheStatsCollector struct {
	hitCount      int64
	missCount     int64
	evictionCount int64
	expiredCount  int64
	mu            sync.RWMutex
}

// NewCacheService 创建缓存管理服务
func NewCacheService(
	redis *redis.Client,
	logger *logrus.Logger,
) CacheService {
	config := &CacheConfig{
		DefaultTTL:      time.Hour * 24,
		MaxKeyLength:    250,
		KeyPrefix:       "yunai:",
		EnableStats:     true,
		EnableWarmup:    true,
		WarmupBatchSize: 100,
		SyncEnabled:     true,
		SyncChannel:     "cache_sync",
	}

	service := &cacheService{
		redis:      redis,
		logger:     logger,
		config:     config,
		strategies: make(map[string]*CacheStrategy),
		stats:      &CacheStatsCollector{},
	}

	// 初始化默认策略
	service.initDefaultStrategies()

	// 启动缓存同步监听
	if config.SyncEnabled {
		go service.startSyncListener()
	}

	// 启动统计收集
	if config.EnableStats {
		go service.startStatsCollection()
	}

	return service
}

// Set 设置缓存
func (s *cacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := s.buildKey(key)

	// 序列化值
	data, err := s.serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// 设置缓存
	if err := s.redis.Set(ctx, fullKey, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"key":        key,
		"expiration": expiration,
	}).Debug("Cache set")

	return nil
}

// Get 获取缓存
func (s *cacheService) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := s.buildKey(key)

	// 获取缓存
	data, err := s.redis.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			s.recordMiss()
			return domain.ErrCacheNotFound
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	// 反序列化值
	if err := s.deserialize(data, dest); err != nil {
		return fmt.Errorf("failed to deserialize value: %w", err)
	}

	s.recordHit()

	s.logger.WithField("key", key).Debug("Cache hit")
	return nil
}

// Delete 删除缓存
func (s *cacheService) Delete(ctx context.Context, key string) error {
	fullKey := s.buildKey(key)

	if err := s.redis.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	s.logger.WithField("key", key).Debug("Cache deleted")
	return nil
}

// Exists 检查缓存是否存在
func (s *cacheService) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := s.buildKey(key)

	count, err := s.redis.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return count > 0, nil
}

// SetMultiple 批量设置缓存
func (s *cacheService) SetMultiple(ctx context.Context, items map[string]interface{}, expiration time.Duration) error {
	pipe := s.redis.Pipeline()

	for key, value := range items {
		fullKey := s.buildKey(key)
		data, err := s.serialize(value)
		if err != nil {
			return fmt.Errorf("failed to serialize value for key %s: %w", key, err)
		}
		pipe.Set(ctx, fullKey, data, expiration)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute batch set: %w", err)
	}

	s.logger.WithField("count", len(items)).Debug("Batch cache set")
	return nil
}

// GetMultiple 批量获取缓存
func (s *cacheService) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 构建完整键名
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = s.buildKey(key)
	}

	// 批量获取
	values, err := s.redis.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple cache: %w", err)
	}

	// 构建结果
	result := make(map[string]interface{})
	hitCount := 0

	for i, value := range values {
		if value != nil {
			var dest interface{}
			if err := s.deserialize(value.(string), &dest); err != nil {
				s.logger.WithError(err).WithField("key", keys[i]).Warn("Failed to deserialize cache value")
				continue
			}
			result[keys[i]] = dest
			hitCount++
		}
	}

	// 记录统计
	s.recordHits(hitCount)
	s.recordMisses(len(keys) - hitCount)

	s.logger.WithFields(logrus.Fields{
		"requested": len(keys),
		"found":     len(result),
	}).Debug("Batch cache get")

	return result, nil
}

// DeleteMultiple 批量删除缓存
func (s *cacheService) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// 构建完整键名
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = s.buildKey(key)
	}

	if err := s.redis.Del(ctx, fullKeys...).Err(); err != nil {
		return fmt.Errorf("failed to delete multiple cache: %w", err)
	}

	s.logger.WithField("count", len(keys)).Debug("Batch cache delete")
	return nil
}

// SetWithStrategy 使用策略设置缓存
func (s *cacheService) SetWithStrategy(ctx context.Context, key string, value interface{}, strategy *CacheStrategy) error {
	if strategy == nil || !strategy.Enabled {
		return s.Set(ctx, key, value, s.config.DefaultTTL)
	}

	// 检查大小限制
	if strategy.MaxSize > 0 {
		data, err := s.serialize(value)
		if err != nil {
			return fmt.Errorf("failed to serialize value: %w", err)
		}
		if int64(len(data)) > strategy.MaxSize {
			return fmt.Errorf("value size exceeds strategy limit")
		}
	}

	// 设置标签
	if len(strategy.Tags) > 0 {
		if err := s.SetTags(ctx, key, strategy.Tags); err != nil {
			s.logger.WithError(err).Warn("Failed to set cache tags")
		}
	}

	return s.Set(ctx, key, value, strategy.TTL)
}

// GetCacheStrategy 获取缓存策略
func (s *cacheService) GetCacheStrategy(ctx context.Context, pattern string) (*CacheStrategy, error) {
	s.strategyMu.RLock()
	defer s.strategyMu.RUnlock()

	strategy, exists := s.strategies[pattern]
	if !exists {
		return nil, fmt.Errorf("strategy not found for pattern: %s", pattern)
	}

	return strategy, nil
}

// UpdateCacheStrategy 更新缓存策略
func (s *cacheService) UpdateCacheStrategy(ctx context.Context, strategy *CacheStrategy) error {
	s.strategyMu.Lock()
	defer s.strategyMu.Unlock()

	s.strategies[strategy.Pattern] = strategy

	s.logger.WithFields(logrus.Fields{
		"pattern": strategy.Pattern,
		"ttl":     strategy.TTL,
		"enabled": strategy.Enabled,
	}).Info("Cache strategy updated")

	return nil
}

// WarmupCache 预热缓存
func (s *cacheService) WarmupCache(ctx context.Context, patterns []string) error {
	if !s.config.EnableWarmup {
		return fmt.Errorf("cache warmup is disabled")
	}

	for _, pattern := range patterns {
		if err := s.warmupPattern(ctx, pattern); err != nil {
			s.logger.WithError(err).WithField("pattern", pattern).Error("Failed to warmup cache pattern")
		}
	}

	return nil
}

// ScheduleWarmup 调度预热
func (s *cacheService) ScheduleWarmup(ctx context.Context, schedule *WarmupSchedule) error {
	// 简化实现，实际应该使用cron调度器
	schedule.ID = uuid.New()

	s.logger.WithFields(logrus.Fields{
		"schedule_id": schedule.ID,
		"pattern":     schedule.Pattern,
		"cron":        schedule.CronExpr,
	}).Info("Cache warmup scheduled")

	return nil
}

// InvalidatePattern 按模式失效缓存
func (s *cacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	fullPattern := s.buildKey(pattern)

	// 获取匹配的键
	keys, err := s.redis.Keys(ctx, fullPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// 批量删除
	if err := s.redis.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"pattern": pattern,
		"count":   len(keys),
	}).Info("Cache invalidated by pattern")

	return nil
}

// InvalidateByTags 按标签失效缓存
func (s *cacheService) InvalidateByTags(ctx context.Context, tags []string) error {
	for _, tag := range tags {
		tagKey := s.buildTagKey(tag)

		// 获取标签关联的键
		keys, err := s.redis.SMembers(ctx, tagKey).Result()
		if err != nil {
			s.logger.WithError(err).WithField("tag", tag).Error("Failed to get keys for tag")
			continue
		}

		if len(keys) > 0 {
			// 删除关联的缓存
			if err := s.redis.Del(ctx, keys...).Err(); err != nil {
				s.logger.WithError(err).WithField("tag", tag).Error("Failed to delete tagged keys")
			}

			// 删除标签集合
			s.redis.Del(ctx, tagKey)
		}
	}

	s.logger.WithField("tags", tags).Info("Cache invalidated by tags")
	return nil
}

// SetTags 设置缓存标签
func (s *cacheService) SetTags(ctx context.Context, key string, tags []string) error {
	fullKey := s.buildKey(key)

	for _, tag := range tags {
		tagKey := s.buildTagKey(tag)
		if err := s.redis.SAdd(ctx, tagKey, fullKey).Err(); err != nil {
			return fmt.Errorf("failed to add key to tag set: %w", err)
		}
	}

	return nil
}

// SyncCache 同步缓存
func (s *cacheService) SyncCache(ctx context.Context, key string) error {
	if !s.config.SyncEnabled {
		return nil
	}

	// 发布同步消息
	message := map[string]interface{}{
		"action": "sync",
		"key":    key,
		"time":   time.Now(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal sync message: %w", err)
	}

	if err := s.redis.Publish(ctx, s.config.SyncChannel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish sync message: %w", err)
	}

	return nil
}

// BroadcastInvalidation 广播失效消息
func (s *cacheService) BroadcastInvalidation(ctx context.Context, key string) error {
	if !s.config.SyncEnabled {
		return nil
	}

	// 发布失效消息
	message := map[string]interface{}{
		"action": "invalidate",
		"key":    key,
		"time":   time.Now(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal invalidation message: %w", err)
	}

	if err := s.redis.Publish(ctx, s.config.SyncChannel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish invalidation message: %w", err)
	}

	return nil
}

// GetCacheStats 获取缓存统计
func (s *cacheService) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	info, err := s.redis.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis info: %w", err)
	}

	dbSize, err := s.redis.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get db size: %w", err)
	}

	s.stats.mu.RLock()
	hitCount := s.stats.hitCount
	missCount := s.stats.missCount
	evictionCount := s.stats.evictionCount
	expiredCount := s.stats.expiredCount
	s.stats.mu.RUnlock()

	totalRequests := hitCount + missCount
	var hitRatio float64
	if totalRequests > 0 {
		hitRatio = float64(hitCount) / float64(totalRequests)
	}

	stats := &CacheStats{
		TotalKeys:     dbSize,
		TotalMemory:   s.parseMemoryUsage(info),
		HitCount:      hitCount,
		MissCount:     missCount,
		HitRatio:      hitRatio,
		EvictionCount: evictionCount,
		ExpiredCount:  expiredCount,
	}

	return stats, nil
}

// GetHitRatio 获取命中率
func (s *cacheService) GetHitRatio(ctx context.Context, pattern string) (float64, error) {
	// 简化实现，实际应该按模式统计
	stats, err := s.GetCacheStats(ctx)
	if err != nil {
		return 0, err
	}

	return stats.HitRatio, nil
}

// GetCacheSize 获取缓存大小
func (s *cacheService) GetCacheSize(ctx context.Context) (int64, error) {
	info, err := s.redis.Info(ctx, "memory").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get redis info: %w", err)
	}

	return s.parseMemoryUsage(info), nil
}

// CleanupExpired 清理过期缓存
func (s *cacheService) CleanupExpired(ctx context.Context) error {
	// Redis会自动清理过期键，这里主要是清理标签关联
	tagPattern := s.buildTagKey("*")
	tagKeys, err := s.redis.Keys(ctx, tagPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get tag keys: %w", err)
	}

	cleanedCount := 0
	for _, tagKey := range tagKeys {
		// 检查标签集合中的键是否还存在
		members, err := s.redis.SMembers(ctx, tagKey).Result()
		if err != nil {
			continue
		}

		var validMembers []string
		for _, member := range members {
			exists, err := s.redis.Exists(ctx, member).Result()
			if err != nil || exists == 0 {
				// 键不存在，从标签集合中移除
				s.redis.SRem(ctx, tagKey, member)
				cleanedCount++
			} else {
				validMembers = append(validMembers, member)
			}
		}

		// 如果标签集合为空，删除标签键
		if len(validMembers) == 0 {
			s.redis.Del(ctx, tagKey)
		}
	}

	s.recordExpired(cleanedCount)

	s.logger.WithField("cleaned_count", cleanedCount).Info("Expired cache cleaned up")
	return nil
}

// FlushCache 清空缓存
func (s *cacheService) FlushCache(ctx context.Context, pattern string) error {
	if pattern == "" || pattern == "*" {
		// 清空所有缓存
		if err := s.redis.FlushDB(ctx).Err(); err != nil {
			return fmt.Errorf("failed to flush cache: %w", err)
		}
		s.logger.Info("All cache flushed")
	} else {
		// 按模式清空
		if err := s.InvalidatePattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to flush cache by pattern: %w", err)
		}
	}

	return nil
}

// 辅助方法

// buildKey 构建完整的缓存键
func (s *cacheService) buildKey(key string) string {
	if len(key) > s.config.MaxKeyLength {
		key = key[:s.config.MaxKeyLength]
	}
	return s.config.KeyPrefix + key
}

// buildTagKey 构建标签键
func (s *cacheService) buildTagKey(tag string) string {
	return s.config.KeyPrefix + "tag:" + tag
}

// serialize 序列化值
func (s *cacheService) serialize(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

// deserialize 反序列化值
func (s *cacheService) deserialize(data string, dest interface{}) error {
	switch dest := dest.(type) {
	case *string:
		*dest = data
		return nil
	case *[]byte:
		*dest = []byte(data)
		return nil
	default:
		return json.Unmarshal([]byte(data), dest)
	}
}

// recordHit 记录命中
func (s *cacheService) recordHit() {
	if !s.config.EnableStats {
		return
	}
	s.stats.mu.Lock()
	s.stats.hitCount++
	s.stats.mu.Unlock()
}

// recordMiss 记录未命中
func (s *cacheService) recordMiss() {
	if !s.config.EnableStats {
		return
	}
	s.stats.mu.Lock()
	s.stats.missCount++
	s.stats.mu.Unlock()
}

// recordHits 记录多次命中
func (s *cacheService) recordHits(count int) {
	if !s.config.EnableStats {
		return
	}
	s.stats.mu.Lock()
	s.stats.hitCount += int64(count)
	s.stats.mu.Unlock()
}

// recordMisses 记录多次未命中
func (s *cacheService) recordMisses(count int) {
	if !s.config.EnableStats {
		return
	}
	s.stats.mu.Lock()
	s.stats.missCount += int64(count)
	s.stats.mu.Unlock()
}

// recordEviction 记录驱逐
func (s *cacheService) recordEviction(count int) {
	if !s.config.EnableStats {
		return
	}
	s.stats.mu.Lock()
	s.stats.evictionCount += int64(count)
	s.stats.mu.Unlock()
}

// recordExpired 记录过期
func (s *cacheService) recordExpired(count int) {
	if !s.config.EnableStats {
		return
	}
	s.stats.mu.Lock()
	s.stats.expiredCount += int64(count)
	s.stats.mu.Unlock()
}

// parseMemoryUsage 解析内存使用量
func (s *cacheService) parseMemoryUsage(info string) int64 {
	// 简化实现，实际应该解析Redis INFO输出
	return 0
}

// initDefaultStrategies 初始化默认策略
func (s *cacheService) initDefaultStrategies() {
	// 用户缓存策略
	s.strategies["user:*"] = &CacheStrategy{
		Pattern:     "user:*",
		TTL:         time.Hour * 2,
		MaxSize:     1024 * 1024, // 1MB
		EvictPolicy: "LRU",
		Tags:        []string{"user"},
		Enabled:     true,
	}

	// 角色缓存策略
	s.strategies["character:*"] = &CacheStrategy{
		Pattern:     "character:*",
		TTL:         time.Hour * 6,
		MaxSize:     1024 * 1024 * 2, // 2MB
		EvictPolicy: "LRU",
		Tags:        []string{"character"},
		Enabled:     true,
	}

	// 聊天缓存策略
	s.strategies["chat:*"] = &CacheStrategy{
		Pattern:     "chat:*",
		TTL:         time.Minute * 30,
		MaxSize:     1024 * 512, // 512KB
		EvictPolicy: "FIFO",
		Tags:        []string{"chat"},
		Enabled:     true,
	}
}

// startSyncListener 启动同步监听
func (s *cacheService) startSyncListener() {
	pubsub := s.redis.Subscribe(context.Background(), s.config.SyncChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var syncMsg map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &syncMsg); err != nil {
			s.logger.WithError(err).Error("Failed to unmarshal sync message")
			continue
		}

		action, ok := syncMsg["action"].(string)
		if !ok {
			continue
		}

		key, ok := syncMsg["key"].(string)
		if !ok {
			continue
		}

		switch action {
		case "invalidate":
			s.Delete(context.Background(), key)
		case "sync":
			// 处理同步逻辑
			s.logger.WithField("key", key).Debug("Cache sync received")
		}
	}
}

// startStatsCollection 启动统计收集
func (s *cacheService) startStatsCollection() {
	ticker := time.NewTicker(time.Minute * 5) // 每5分钟收集一次统计
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()

		// 获取Redis统计信息
		info, err := s.redis.Info(ctx, "stats").Result()
		if err != nil {
			s.logger.WithError(err).Error("Failed to get Redis stats")
			continue
		}

		// 解析并更新统计信息
		s.updateStatsFromRedis(info)
	}
}

// updateStatsFromRedis 从Redis更新统计信息
func (s *cacheService) updateStatsFromRedis(info string) {
	// 简化实现，实际应该解析Redis INFO stats输出
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "keyspace_hits:") {
			// 解析命中数
		} else if strings.HasPrefix(line, "keyspace_misses:") {
			// 解析未命中数
		} else if strings.HasPrefix(line, "evicted_keys:") {
			// 解析驱逐数
		} else if strings.HasPrefix(line, "expired_keys:") {
			// 解析过期数
		}
	}
}

// warmupPattern 预热指定模式的缓存
func (s *cacheService) warmupPattern(ctx context.Context, pattern string) error {
	// 简化实现，实际应该根据模式调用相应的预热函数
	s.logger.WithField("pattern", pattern).Info("Cache warmup started")

	// 这里应该根据pattern调用相应的数据加载函数
	// 例如：如果pattern是"user:*"，则预加载热门用户数据

	return nil
}
