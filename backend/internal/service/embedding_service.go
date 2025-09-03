package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// EmbeddingService 嵌入模型服务
type EmbeddingService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// EmbeddingRequest 嵌入请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse 嵌入响应
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewEmbeddingService 创建嵌入服务
func NewEmbeddingService(apiKey, baseURL string) *EmbeddingService {
	return &EmbeddingService{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GetEmbedding 获取单个文本的嵌入向量
func (s *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := s.GetEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	
	return embeddings[0], nil
}

// GetEmbeddings 获取多个文本的嵌入向量
func (s *EmbeddingService) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	request := &EmbeddingRequest{
		Model: "BAAI/bge-large-zh-v1.5", // 使用中文嵌入模型
		Input: texts,
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}
	
	var response EmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}
	
	var embeddings [][]float64
	for _, data := range response.Data {
		embeddings = append(embeddings, data.Embedding)
	}
	
	return embeddings, nil
}

// CosineSimilarity 计算余弦相似度
func (s *EmbeddingService) CosineSimilarity(vec1, vec2 []float64) float64 {
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

// FindBestMatch 找到最佳匹配
func (s *EmbeddingService) FindBestMatch(ctx context.Context, query string, candidates []string, threshold float64) (string, float64, error) {
	if len(candidates) == 0 {
		return "", 0.0, fmt.Errorf("no candidates provided")
	}
	
	// 获取查询文本的嵌入
	queryEmbedding, err := s.GetEmbedding(ctx, query)
	if err != nil {
		return "", 0.0, fmt.Errorf("get query embedding failed: %w", err)
	}
	
	// 获取候选文本的嵌入
	candidateEmbeddings, err := s.GetEmbeddings(ctx, candidates)
	if err != nil {
		return "", 0.0, fmt.Errorf("get candidate embeddings failed: %w", err)
	}
	
	bestMatch := ""
	bestScore := 0.0
	
	for i, candidateEmbedding := range candidateEmbeddings {
		similarity := s.CosineSimilarity(queryEmbedding, candidateEmbedding)
		if similarity > bestScore && similarity >= threshold {
			bestScore = similarity
			bestMatch = candidates[i]
		}
	}
	
	return bestMatch, bestScore, nil
}

// FindTopMatches 找到前N个最佳匹配
func (s *EmbeddingService) FindTopMatches(ctx context.Context, query string, candidates []string, topN int, threshold float64) ([]MatchResult, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates provided")
	}
	
	// 获取查询文本的嵌入
	queryEmbedding, err := s.GetEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get query embedding failed: %w", err)
	}
	
	// 获取候选文本的嵌入
	candidateEmbeddings, err := s.GetEmbeddings(ctx, candidates)
	if err != nil {
		return nil, fmt.Errorf("get candidate embeddings failed: %w", err)
	}
	
	var results []MatchResult
	
	for i, candidateEmbedding := range candidateEmbeddings {
		similarity := s.CosineSimilarity(queryEmbedding, candidateEmbedding)
		if similarity >= threshold {
			results = append(results, MatchResult{
				Text:       candidates[i],
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
	
	// 返回前N个结果
	if topN > len(results) {
		topN = len(results)
	}
	
	return results[:topN], nil
}

// MatchResult 匹配结果
type MatchResult struct {
	Text       string  `json:"text"`
	Similarity float64 `json:"similarity"`
	Index      int     `json:"index"`
}

// SemanticSearch 语义搜索
func (s *EmbeddingService) SemanticSearch(ctx context.Context, query string, documents []string, topK int) ([]MatchResult, error) {
	return s.FindTopMatches(ctx, query, documents, topK, 0.5) // 默认阈值0.5
}

// ClusterTexts 文本聚类（简单版本）
func (s *EmbeddingService) ClusterTexts(ctx context.Context, texts []string, threshold float64) ([][]int, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}
	
	// 获取所有文本的嵌入
	embeddings, err := s.GetEmbeddings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("get embeddings failed: %w", err)
	}
	
	var clusters [][]int
	used := make([]bool, len(texts))
	
	for i := 0; i < len(texts); i++ {
		if used[i] {
			continue
		}
		
		cluster := []int{i}
		used[i] = true
		
		for j := i + 1; j < len(texts); j++ {
			if used[j] {
				continue
			}
			
			similarity := s.CosineSimilarity(embeddings[i], embeddings[j])
			if similarity >= threshold {
				cluster = append(cluster, j)
				used[j] = true
			}
		}
		
		clusters = append(clusters, cluster)
	}
	
	return clusters, nil
}

// AnalyzeTextSimilarity 分析文本相似性
func (s *EmbeddingService) AnalyzeTextSimilarity(ctx context.Context, text1, text2 string) (float64, error) {
	embeddings, err := s.GetEmbeddings(ctx, []string{text1, text2})
	if err != nil {
		return 0.0, fmt.Errorf("get embeddings failed: %w", err)
	}
	
	if len(embeddings) != 2 {
		return 0.0, fmt.Errorf("expected 2 embeddings, got %d", len(embeddings))
	}
	
	return s.CosineSimilarity(embeddings[0], embeddings[1]), nil
}
