package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EnhancedEmbeddingService 增强的嵌入服务
type EnhancedEmbeddingService struct {
	apiKey     string
	baseURL    string
	client     *http.Client
	logger     *logrus.Logger
	cache      *EmbeddingCache
	config     *EmbeddingConfig
	modelStats map[string]*ModelStats
	mu         sync.RWMutex
}

// EmbeddingConfig 嵌入配置
type EmbeddingConfig struct {
	PrimaryModel   string                  `json:"primary_model"`
	FallbackModels []string                `json:"fallback_models"`
	BatchSize      int                     `json:"batch_size"`
	MaxRetries     int                     `json:"max_retries"`
	CacheEnabled   bool                    `json:"cache_enabled"`
	CacheTTL       time.Duration           `json:"cache_ttl"`
	ModelConfigs   map[string]*ModelConfig `json:"model_configs"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	Dimensions     int `json:"dimensions"`
	MaxInputTokens int `json:"max_input_tokens"`
	BatchSize      int `json:"batch_size"`
}

// ModelStats 模型统计
type ModelStats struct {
	RequestCount int64         `json:"request_count"`
	SuccessCount int64         `json:"success_count"`
	ErrorCount   int64         `json:"error_count"`
	AvgLatency   time.Duration `json:"avg_latency"`
	LastUsed     time.Time     `json:"last_used"`
}

// EmbeddingCache 嵌入缓存
type EmbeddingCache struct {
	cache map[string]*CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Embedding []float64 `json:"embedding"`
	CreatedAt time.Time `json:"created_at"`
	Model     string    `json:"model"`
}

// NewEnhancedEmbeddingService 创建增强的嵌入服务
func NewEnhancedEmbeddingService(apiKey, baseURL string, logger *logrus.Logger) *EnhancedEmbeddingService {
	config := &EmbeddingConfig{
		PrimaryModel:   "BAAI/bge-large-zh-v1.5", // 基于测试结果的最佳模型
		FallbackModels: []string{"BAAI/bge-m3", "netease-youdao/bce-embedding-base_v1"},
		BatchSize:      32,
		MaxRetries:     3,
		CacheEnabled:   true,
		CacheTTL:       24 * time.Hour,
		ModelConfigs: map[string]*ModelConfig{
			"BAAI/bge-large-zh-v1.5": {
				Dimensions:     1024,
				MaxInputTokens: 512,
				BatchSize:      32,
			},
			"BAAI/bge-m3": {
				Dimensions:     1024,
				MaxInputTokens: 8192,
				BatchSize:      16,
			},
			"netease-youdao/bce-embedding-base_v1": {
				Dimensions:     768,
				MaxInputTokens: 512,
				BatchSize:      32,
			},
		},
	}

	cache := &EmbeddingCache{
		cache: make(map[string]*CacheEntry),
		ttl:   config.CacheTTL,
	}

	return &EnhancedEmbeddingService{
		apiKey:     apiKey,
		baseURL:    baseURL,
		client:     &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
		cache:      cache,
		config:     config,
		modelStats: make(map[string]*ModelStats),
	}
}

// GetEmbedding 获取单个文本的嵌入向量（带缓存和故障转移）
func (s *EnhancedEmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	return s.GetEmbeddingWithModel(ctx, text, s.config.PrimaryModel)
}

// GetEmbeddingWithModel 使用指定模型获取嵌入向量
func (s *EnhancedEmbeddingService) GetEmbeddingWithModel(ctx context.Context, text, model string) ([]float64, error) {
	// 检查缓存
	if s.config.CacheEnabled {
		if cached := s.getCachedEmbedding(text, model); cached != nil {
			s.logger.Debug("使用缓存的嵌入向量")
			return cached, nil
		}
	}

	// 尝试主模型
	embedding, err := s.getEmbeddingFromAPI(ctx, text, model)
	if err == nil {
		// 缓存结果
		if s.config.CacheEnabled {
			s.cacheEmbedding(text, model, embedding)
		}
		s.updateModelStats(model, true, time.Since(time.Now()))
		return embedding, nil
	}

	s.logger.WithError(err).Warnf("主模型 %s 失败，尝试故障转移", model)
	s.updateModelStats(model, false, 0)

	// 尝试故障转移模型
	for _, fallbackModel := range s.config.FallbackModels {
		s.logger.Infof("尝试故障转移模型: %s", fallbackModel)

		embedding, err := s.getEmbeddingFromAPI(ctx, text, fallbackModel)
		if err == nil {
			// 缓存结果
			if s.config.CacheEnabled {
				s.cacheEmbedding(text, fallbackModel, embedding)
			}
			s.updateModelStats(fallbackModel, true, time.Since(time.Now()))
			return embedding, nil
		}

		s.logger.WithError(err).Warnf("故障转移模型 %s 也失败", fallbackModel)
		s.updateModelStats(fallbackModel, false, 0)
	}

	return nil, fmt.Errorf("所有嵌入模型都失败了")
}

// GetEmbeddings 批量获取嵌入向量
func (s *EnhancedEmbeddingService) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return s.GetEmbeddingsWithModel(ctx, texts, s.config.PrimaryModel)
}

// GetEmbeddingsWithModel 使用指定模型批量获取嵌入向量
func (s *EnhancedEmbeddingService) GetEmbeddingsWithModel(ctx context.Context, texts []string, model string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("没有提供文本")
	}

	modelConfig := s.config.ModelConfigs[model]
	if modelConfig == nil {
		return nil, fmt.Errorf("未知模型: %s", model)
	}

	batchSize := modelConfig.BatchSize
	var allEmbeddings [][]float64

	// 分批处理
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		batchEmbeddings, err := s.getBatchEmbeddingsFromAPI(ctx, batch, model)
		if err != nil {
			// 尝试故障转移
			for _, fallbackModel := range s.config.FallbackModels {
				batchEmbeddings, err = s.getBatchEmbeddingsFromAPI(ctx, batch, fallbackModel)
				if err == nil {
					break
				}
			}

			if err != nil {
				return nil, fmt.Errorf("批量获取嵌入失败: %w", err)
			}
		}

		allEmbeddings = append(allEmbeddings, batchEmbeddings...)
	}

	return allEmbeddings, nil
}

// getEmbeddingFromAPI 从API获取单个嵌入向量
func (s *EnhancedEmbeddingService) getEmbeddingFromAPI(ctx context.Context, text, model string) ([]float64, error) {
	embeddings, err := s.getBatchEmbeddingsFromAPI(ctx, []string{text}, model)
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("没有返回嵌入向量")
	}

	return embeddings[0], nil
}

// getBatchEmbeddingsFromAPI 从API批量获取嵌入向量
func (s *EnhancedEmbeddingService) getBatchEmbeddingsFromAPI(ctx context.Context, texts []string, model string) ([][]float64, error) {
	startTime := time.Now()

	request := &EmbeddingRequest{
		Model: model,
		Input: texts,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API错误 %d: %s", resp.StatusCode, string(body))
	}

	var response EmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	var embeddings [][]float64
	for _, data := range response.Data {
		embeddings = append(embeddings, data.Embedding)
	}

	// 记录性能指标
	latency := time.Since(startTime)
	s.logger.WithFields(logrus.Fields{
		"model":   model,
		"texts":   len(texts),
		"latency": latency,
		"tokens":  response.Usage.TotalTokens,
	}).Debug("嵌入向量获取成功")

	return embeddings, nil
}

// getCachedEmbedding 获取缓存的嵌入向量
func (s *EnhancedEmbeddingService) getCachedEmbedding(text, model string) []float64 {
	key := s.getCacheKey(text, model)

	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()

	entry, exists := s.cache.cache[key]
	if !exists {
		return nil
	}

	// 检查是否过期
	if time.Since(entry.CreatedAt) > s.cache.ttl {
		// 异步删除过期条目
		go func() {
			s.cache.mu.Lock()
			delete(s.cache.cache, key)
			s.cache.mu.Unlock()
		}()
		return nil
	}

	return entry.Embedding
}

// cacheEmbedding 缓存嵌入向量
func (s *EnhancedEmbeddingService) cacheEmbedding(text, model string, embedding []float64) {
	key := s.getCacheKey(text, model)

	entry := &CacheEntry{
		Embedding: embedding,
		CreatedAt: time.Now(),
		Model:     model,
	}

	s.cache.mu.Lock()
	s.cache.cache[key] = entry
	s.cache.mu.Unlock()
}

// getCacheKey 生成缓存键
func (s *EnhancedEmbeddingService) getCacheKey(text, model string) string {
	hash := md5.Sum([]byte(text + "|" + model))
	return hex.EncodeToString(hash[:])
}

// updateModelStats 更新模型统计
func (s *EnhancedEmbeddingService) updateModelStats(model string, success bool, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.modelStats[model]
	if !exists {
		stats = &ModelStats{}
		s.modelStats[model] = stats
	}

	stats.RequestCount++
	stats.LastUsed = time.Now()

	if success {
		stats.SuccessCount++
		// 计算平均延迟
		if stats.AvgLatency == 0 {
			stats.AvgLatency = latency
		} else {
			stats.AvgLatency = (stats.AvgLatency + latency) / 2
		}
	} else {
		stats.ErrorCount++
	}
}

// GetModelStats 获取模型统计
func (s *EnhancedEmbeddingService) GetModelStats() map[string]*ModelStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本
	result := make(map[string]*ModelStats)
	for k, v := range s.modelStats {
		result[k] = &ModelStats{
			RequestCount: v.RequestCount,
			SuccessCount: v.SuccessCount,
			ErrorCount:   v.ErrorCount,
			AvgLatency:   v.AvgLatency,
			LastUsed:     v.LastUsed,
		}
	}

	return result
}

// CosineSimilarity 计算余弦相似度
func (s *EnhancedEmbeddingService) CosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	var dotProduct, norm1, norm2 float64

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0.0 || norm2 == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// SemanticSearch 语义搜索（增强版）
func (s *EnhancedEmbeddingService) SemanticSearch(ctx context.Context, query string, documents []string, topK int, threshold float64) ([]MatchResult, error) {
	if len(documents) == 0 {
		return nil, fmt.Errorf("没有提供文档")
	}

	// 获取查询嵌入
	queryEmbedding, err := s.GetEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("获取查询嵌入失败: %w", err)
	}

	// 获取文档嵌入
	docEmbeddings, err := s.GetEmbeddings(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("获取文档嵌入失败: %w", err)
	}

	var results []MatchResult
	for i, docEmbedding := range docEmbeddings {
		similarity := s.CosineSimilarity(queryEmbedding, docEmbedding)
		if similarity >= threshold {
			results = append(results, MatchResult{
				Text:       documents[i],
				Similarity: similarity,
				Index:      i,
			})
		}
	}

	// 按相似度排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 返回前K个结果
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

// 注意：MatchResult, EmbeddingRequest, EmbeddingResponse 类型定义在原始的 embedding_service.go 中

// AnalyzeMomentsSimilarity 分析朋友圈内容相似度
func (s *EnhancedEmbeddingService) AnalyzeMomentsSimilarity(ctx context.Context, moments []string, threshold float64) ([]SimilarityGroup, error) {
	if len(moments) < 2 {
		return nil, fmt.Errorf("至少需要2条朋友圈内容")
	}

	// 获取所有朋友圈的嵌入向量
	embeddings, err := s.GetEmbeddings(ctx, moments)
	if err != nil {
		return nil, fmt.Errorf("获取嵌入向量失败: %w", err)
	}

	var groups []SimilarityGroup
	used := make([]bool, len(moments))

	for i := 0; i < len(moments); i++ {
		if used[i] {
			continue
		}

		group := SimilarityGroup{
			Representative: moments[i],
			Members:        []SimilarityMember{{Index: i, Text: moments[i], Similarity: 1.0}},
		}
		used[i] = true

		for j := i + 1; j < len(moments); j++ {
			if used[j] {
				continue
			}

			similarity := s.CosineSimilarity(embeddings[i], embeddings[j])
			if similarity >= threshold {
				group.Members = append(group.Members, SimilarityMember{
					Index:      j,
					Text:       moments[j],
					Similarity: similarity,
				})
				used[j] = true
			}
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// SimilarityGroup 相似度组
type SimilarityGroup struct {
	Representative string             `json:"representative"`
	Members        []SimilarityMember `json:"members"`
}

// SimilarityMember 相似度成员
type SimilarityMember struct {
	Index      int     `json:"index"`
	Text       string  `json:"text"`
	Similarity float64 `json:"similarity"`
}

// RecommendMomentsContent 基于历史内容推荐朋友圈内容
func (s *EnhancedEmbeddingService) RecommendMomentsContent(ctx context.Context, userHistory []string, candidateContents []string, topK int) ([]MatchResult, error) {
	if len(userHistory) == 0 {
		return nil, fmt.Errorf("用户历史为空")
	}
	if len(candidateContents) == 0 {
		return nil, fmt.Errorf("候选内容为空")
	}

	// 计算用户历史内容的平均嵌入向量
	historyEmbeddings, err := s.GetEmbeddings(ctx, userHistory)
	if err != nil {
		return nil, fmt.Errorf("获取历史嵌入失败: %w", err)
	}

	// 计算平均向量
	avgEmbedding := make([]float64, len(historyEmbeddings[0]))
	for _, embedding := range historyEmbeddings {
		for i, val := range embedding {
			avgEmbedding[i] += val
		}
	}
	for i := range avgEmbedding {
		avgEmbedding[i] /= float64(len(historyEmbeddings))
	}

	// 获取候选内容的嵌入向量
	candidateEmbeddings, err := s.GetEmbeddings(ctx, candidateContents)
	if err != nil {
		return nil, fmt.Errorf("获取候选嵌入失败: %w", err)
	}

	var results []MatchResult
	for i, candidateEmbedding := range candidateEmbeddings {
		similarity := s.CosineSimilarity(avgEmbedding, candidateEmbedding)
		results = append(results, MatchResult{
			Text:       candidateContents[i],
			Similarity: similarity,
			Index:      i,
		})
	}

	// 按相似度排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 返回前K个结果
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

// ClearCache 清理缓存
func (s *EnhancedEmbeddingService) ClearCache() {
	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	s.cache.cache = make(map[string]*CacheEntry)
	s.logger.Info("嵌入向量缓存已清理")
}

// GetCacheStats 获取缓存统计
func (s *EnhancedEmbeddingService) GetCacheStats() EmbeddingCacheStats {
	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()

	totalEntries := len(s.cache.cache)
	expiredEntries := 0

	now := time.Now()
	for _, entry := range s.cache.cache {
		if now.Sub(entry.CreatedAt) > s.cache.ttl {
			expiredEntries++
		}
	}

	return EmbeddingCacheStats{
		TotalEntries:   totalEntries,
		ExpiredEntries: expiredEntries,
		ValidEntries:   totalEntries - expiredEntries,
		TTL:            s.cache.ttl,
	}
}

// EmbeddingCacheStats 嵌入缓存统计
type EmbeddingCacheStats struct {
	TotalEntries   int           `json:"total_entries"`
	ExpiredEntries int           `json:"expired_entries"`
	ValidEntries   int           `json:"valid_entries"`
	TTL            time.Duration `json:"ttl"`
}

// HealthCheck 健康检查
func (s *EnhancedEmbeddingService) HealthCheck(ctx context.Context) error {
	// 测试主模型
	_, err := s.GetEmbedding(ctx, "健康检查测试")
	if err != nil {
		return fmt.Errorf("主模型健康检查失败: %w", err)
	}

	s.logger.Info("嵌入服务健康检查通过")
	return nil
}
