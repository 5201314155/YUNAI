package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// SiliconFlow API请求结构
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SiliconFlow API响应结构
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message,omitempty"`
	Delta        Message `json:"delta,omitempty"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// 流式响应结构
type StreamResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

func main() {
	fmt.Println("🚀 YUNAI真实SiliconFlow模型测试")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// SiliconFlow API配置
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"
	baseURL := "https://api.siliconflow.cn/v1/chat/completions"

	// 获取支持流式的模型列表
	fmt.Println("📋 获取支持流式的模型...")
	streamingModels := getStreamingModels(db)

	if len(streamingModels) == 0 {
		fmt.Println("❌ 没有找到支持流式的模型")
		return
	}

	// 测试消息
	testMessages := []Message{
		{Role: "user", Content: "请简单介绍一下人工智能的发展历史，大概200字左右。"},
	}

	// 测试每个支持流式的模型
	for i, model := range streamingModels {
		fmt.Printf("\n🤖 测试模型 %d/%d: %s\n", i+1, len(streamingModels), model.DisplayName)
		fmt.Printf("   模型ID: %s\n", model.InternalKey)
		fmt.Printf("   支持流式: %t\n", model.SupportStreaming)
		fmt.Println("-------------------------------------------")

		// 1. 测试非流式输出
		fmt.Println("📝 测试非流式输出...")
		testNonStreaming(apiKey, baseURL, model.InternalKey, testMessages)

		// 2. 测试流式输出 (如果支持)
		if model.SupportStreaming {
			fmt.Println("\n🌊 测试流式输出...")
			testStreaming(apiKey, baseURL, model.InternalKey, testMessages)
		}

		// 记录测试结果到数据库
		recordTestResult(db, model.ID, true)

		// 间隔一下避免请求过快
		if i < len(streamingModels)-1 {
			fmt.Println("\n⏳ 等待3秒后测试下一个模型...")
			time.Sleep(3 * time.Second)
		}
	}

	fmt.Println("\n🎉 所有模型测试完成！")
}

// 模型信息结构
type ModelInfo struct {
	ID               uuid.UUID `db:"id"`
	InternalKey      string    `db:"internal_key"`
	DisplayName      string    `db:"display_name"`
	SupportStreaming bool      `db:"support_streaming"`
	ModelType        string    `db:"model_type"`
}

// 获取支持流式的模型
func getStreamingModels(db *sql.DB) []ModelInfo {
	query := `
		SELECT id, internal_key, display_name, support_streaming, model_type
		FROM ai_models 
		WHERE provider = 'siliconflow' 
		AND model_type = 'chat'
		AND is_active = true
		AND support_streaming = true
		ORDER BY weight DESC
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询模型失败: %v", err)
		return nil
	}
	defer rows.Close()

	var models []ModelInfo
	for rows.Next() {
		var model ModelInfo
		err := rows.Scan(&model.ID, &model.InternalKey, &model.DisplayName, &model.SupportStreaming, &model.ModelType)
		if err != nil {
			continue
		}
		models = append(models, model)
	}

	fmt.Printf("✅ 找到 %d 个支持流式的模型\n", len(models))
	return models
}

// 测试非流式输出
func testNonStreaming(apiKey, baseURL, modelKey string, messages []Message) {
	// 构建请求
	request := ChatRequest{
		Model:       modelKey,
		Messages:    messages,
		Stream:      false,
		Temperature: 0.7,
		MaxTokens:   500,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ 构建请求失败: %v\n", err)
		return
	}

	// 发送请求
	startTime := time.Now()
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ API请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("   响应: %s\n", string(body))
		return
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("   原始响应: %s\n", string(body))
		return
	}

	// 显示结果
	if len(chatResp.Choices) > 0 {
		content := chatResp.Choices[0].Message.Content
		fmt.Printf("✅ 非流式输出成功!\n")
		fmt.Printf("   响应时间: %v\n", duration)
		fmt.Printf("   Token使用: %d (输入) + %d (输出) = %d (总计)\n", 
			chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)
		fmt.Printf("   响应长度: %d 字符\n", len(content))
		fmt.Printf("   响应内容: %s\n", truncateString(content, 200))
	} else {
		fmt.Printf("❌ 响应中没有内容\n")
	}
}

// 测试流式输出
func testStreaming(apiKey, baseURL, modelKey string, messages []Message) {
	// 构建请求
	request := ChatRequest{
		Model:       modelKey,
		Messages:    messages,
		Stream:      true,
		Temperature: 0.7,
		MaxTokens:   500,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ 构建请求失败: %v\n", err)
		return
	}

	// 发送请求
	startTime := time.Now()
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ API请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("   响应: %s\n", string(body))
		return
	}

	// 读取流式响应
	fmt.Printf("🌊 流式输出开始:\n")
	fmt.Printf("-------------------------------------------\n")

	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	chunkCount := 0
	firstChunkTime := time.Time{}

	for scanner.Scan() {
		line := scanner.Text()
		
		// 跳过空行和非数据行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// 提取JSON数据
		jsonStr := strings.TrimPrefix(line, "data: ")
		if jsonStr == "[DONE]" {
			break
		}

		// 解析流式响应
		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(jsonStr), &streamResp); err != nil {
			continue
		}

		// 处理内容
		if len(streamResp.Choices) > 0 {
			delta := streamResp.Choices[0].Delta.Content
			if delta != "" {
				if chunkCount == 0 {
					firstChunkTime = time.Now()
				}
				chunkCount++
				fullContent.WriteString(delta)
				fmt.Print(delta) // 实时显示流式输出
			}
		}
	}

	totalDuration := time.Since(startTime)
	firstChunkDuration := firstChunkTime.Sub(startTime)

	fmt.Printf("\n-------------------------------------------\n")
	fmt.Printf("✅ 流式输出完成!\n")
	fmt.Printf("   总响应时间: %v\n", totalDuration)
	fmt.Printf("   首字符时间: %v\n", firstChunkDuration)
	fmt.Printf("   流式块数量: %d\n", chunkCount)
	fmt.Printf("   响应长度: %d 字符\n", fullContent.Len())
	fmt.Printf("   完整内容: %s\n", truncateString(fullContent.String(), 200))

	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ 读取流式响应出错: %v\n", err)
	}
}

// 记录测试结果到数据库
func recordTestResult(db *sql.DB, modelID uuid.UUID, success bool) {
	query := `
		INSERT INTO model_test_logs (model_id, test_type, success, tested_at)
		VALUES ($1, 'streaming_test', $2, NOW())
		ON CONFLICT DO NOTHING
	`
	
	// 如果表不存在就创建
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS model_test_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			model_id UUID NOT NULL,
			test_type VARCHAR(50) NOT NULL,
			success BOOLEAN NOT NULL,
			tested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	
	db.Exec(createTableQuery)
	db.Exec(query, modelID, success)
}

// 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
