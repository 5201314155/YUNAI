package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// API请求/响应结构
type ChatRequest struct {
	ModelID     string    `json:"model_id"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ModelInfo struct {
	ID                string `json:"id"`
	InternalKey       string `json:"internal_key"`
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	Type              string `json:"type"`
	Category          string `json:"category"`
	Description       string `json:"description"`
	MaxTokens         int    `json:"max_tokens"`
	SupportStreaming  bool   `json:"support_streaming"`
	IsFeatured        bool   `json:"is_featured"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage,omitempty"`
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

const BASE_URL = "http://localhost:8080/api/v1"

func main() {
	fmt.Println("🧪 YUNAI API客户端测试 - 流式 vs 非流式对话")
	fmt.Println("===========================================")

	// 等待API服务启动
	fmt.Println("⏳ 等待API服务启动...")
	time.Sleep(3 * time.Second)

	// 1. 获取系统信息
	fmt.Println("📊 获取系统信息...")
	getSystemInfo()

	// 2. 获取模型列表
	fmt.Println("\n📋 获取对话模型列表...")
	models := getModels()

	if len(models) == 0 {
		fmt.Println("❌ 没有找到可用的对话模型")
		return
	}

	// 选择一个支持流式的模型进行测试
	var testModel *ModelInfo
	for _, model := range models {
		if model.Type == "chat" && model.SupportStreaming {
			testModel = &model
			break
		}
	}

	if testModel == nil {
		fmt.Println("❌ 没有找到支持流式的对话模型")
		return
	}

	fmt.Printf("🎯 选择测试模型: %s (%s)\n", testModel.Name, testModel.InternalKey)

	// 测试对话内容
	testMessages := []Message{
		{Role: "user", Content: "请简单介绍一下人工智能的发展历史，大概150字左右。"},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 3. 测试非流式对话
	fmt.Println("📝 测试非流式对话")
	fmt.Println(strings.Repeat("-", 40))
	testNonStreamingChat(testModel.ID, testMessages)

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 4. 测试流式对话
	fmt.Println("🌊 测试流式对话")
	fmt.Println(strings.Repeat("-", 40))
	testStreamingChat(testModel.ID, testMessages)

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 5. 对比测试 - 更复杂的问题
	fmt.Println("🔄 对比测试 - 复杂问题")
	fmt.Println(strings.Repeat("-", 40))
	
	complexMessages := []Message{
		{Role: "user", Content: "请详细解释机器学习中的监督学习和无监督学习的区别，并各举一个实际应用例子。"},
	}

	fmt.Println("📝 非流式 - 复杂问题:")
	testNonStreamingChat(testModel.ID, complexMessages)

	fmt.Println("\n🌊 流式 - 复杂问题:")
	testStreamingChat(testModel.ID, complexMessages)

	fmt.Println("\n🎉 API对话测试完成！")
}

// 获取系统信息
func getSystemInfo() {
	resp, err := http.Get(BASE_URL + "/system/info")
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if system, ok := data["system"].(map[string]interface{}); ok {
			fmt.Printf("✅ 系统: %s v%s (%s)\n", system["name"], system["version"], system["status"])
		}
		if models, ok := data["models"].(map[string]interface{}); ok {
			fmt.Printf("✅ 模型: 总计%v个, 活跃%v个, 对话%v个, 流式%v个\n", 
				models["total"], models["active"], models["chat"], models["streaming"])
		}
		if voices, ok := data["voices"].(map[string]interface{}); ok {
			fmt.Printf("✅ 音色: 总计%v个, 公开%v个\n", voices["total"], voices["public"])
		}
	}
}

// 获取模型列表
func getModels() []ModelInfo {
	resp, err := http.Get(BASE_URL + "/models?type=chat")
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return nil
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return nil
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if modelsData, ok := data["models"].([]interface{}); ok {
			var models []ModelInfo
			for _, modelData := range modelsData {
				if modelMap, ok := modelData.(map[string]interface{}); ok {
					model := ModelInfo{
						ID:               getString(modelMap, "id"),
						InternalKey:      getString(modelMap, "internal_key"),
						Name:             getString(modelMap, "name"),
						Provider:         getString(modelMap, "provider"),
						Type:             getString(modelMap, "type"),
						Category:         getString(modelMap, "category"),
						Description:      getString(modelMap, "description"),
						MaxTokens:        getInt(modelMap, "max_tokens"),
						SupportStreaming: getBool(modelMap, "support_streaming"),
						IsFeatured:       getBool(modelMap, "is_featured"),
					}
					models = append(models, model)
				}
			}
			
			fmt.Printf("✅ 找到 %d 个对话模型\n", len(models))
			for i, model := range models {
				streamIcon := "❌"
				if model.SupportStreaming {
					streamIcon = "🌊"
				}
				featuredIcon := ""
				if model.IsFeatured {
					featuredIcon = "⭐"
				}
				fmt.Printf("   %d. %s %s %s [%s]\n", i+1, streamIcon, featuredIcon, model.Name, model.Provider)
			}
			
			return models
		}
	}

	return nil
}

// 测试非流式对话
func testNonStreamingChat(modelID string, messages []Message) {
	fmt.Printf("🤖 模型ID: %s\n", modelID)
	fmt.Printf("❓ 问题: %s\n", messages[0].Content)
	fmt.Printf("⏱️ 开始时间: %s\n", time.Now().Format("15:04:05"))

	request := ChatRequest{
		ModelID:     modelID,
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

	startTime := time.Now()
	resp, err := http.Post(BASE_URL+"/chat/completions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

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

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}

	if len(chatResp.Choices) > 0 {
		content := chatResp.Choices[0].Message.Content
		fmt.Printf("⏱️ 结束时间: %s\n", time.Now().Format("15:04:05"))
		fmt.Printf("⚡ 总响应时间: %v\n", duration)
		if chatResp.Usage.TotalTokens > 0 {
			fmt.Printf("📊 Token使用: %d输入 + %d输出 = %d总计\n", 
				chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)
		}
		fmt.Printf("📏 响应长度: %d 字符\n", len(content))
		fmt.Printf("💬 AI回答:\n")
		fmt.Printf("┌%s┐\n", strings.Repeat("─", 60))
		
		lines := wrapText(content, 58)
		for _, line := range lines {
			fmt.Printf("│ %-58s │\n", line)
		}
		
		fmt.Printf("└%s┘\n", strings.Repeat("─", 60))
	} else {
		fmt.Printf("❌ 响应中没有内容\n")
	}
}

// 测试流式对话
func testStreamingChat(modelID string, messages []Message) {
	fmt.Printf("🤖 模型ID: %s\n", modelID)
	fmt.Printf("❓ 问题: %s\n", messages[0].Content)
	fmt.Printf("⏱️ 开始时间: %s\n", time.Now().Format("15:04:05"))

	request := ChatRequest{
		ModelID:     modelID,
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

	startTime := time.Now()
	resp, err := http.Post(BASE_URL+"/chat/completions", "application/json", bytes.NewBuffer(jsonData))
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

	fmt.Printf("💬 AI回答 (实时流式输出):\n")
	fmt.Printf("┌%s┐\n", strings.Repeat("─", 60))
	fmt.Print("│ ")

	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	chunkCount := 0
	firstChunkTime := time.Time{}
	currentLineLength := 0

	for scanner.Scan() {
		line := scanner.Text()
		
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		jsonStr := strings.TrimPrefix(line, "data: ")
		if jsonStr == "[DONE]" {
			break
		}

		var streamResp map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &streamResp); err != nil {
			continue
		}

		if choices, ok := streamResp["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok && content != "" {
						if chunkCount == 0 {
							firstChunkTime = time.Now()
						}
						chunkCount++
						fullContent.WriteString(content)
						
						// 实时显示流式输出
						for _, char := range content {
							if char == '\n' {
								fmt.Printf("%s │\n│ ", strings.Repeat(" ", 58-currentLineLength))
								currentLineLength = 0
							} else {
								fmt.Print(string(char))
								currentLineLength++
								if currentLineLength >= 58 {
									fmt.Printf(" │\n│ ")
									currentLineLength = 0
								}
							}
						}
						
						time.Sleep(20 * time.Millisecond)
					}
				}
			}
		}
	}

	if currentLineLength > 0 {
		fmt.Printf("%s │\n", strings.Repeat(" ", 58-currentLineLength))
	}
	fmt.Printf("└%s┘\n", strings.Repeat("─", 60))

	totalDuration := time.Since(startTime)
	firstChunkDuration := firstChunkTime.Sub(startTime)

	fmt.Printf("⏱️ 结束时间: %s\n", time.Now().Format("15:04:05"))
	fmt.Printf("⚡ 总响应时间: %v\n", totalDuration)
	fmt.Printf("🚀 首字符时间: %v (TTFB)\n", firstChunkDuration)
	fmt.Printf("📦 流式块数量: %d\n", chunkCount)
	fmt.Printf("📏 响应长度: %d 字符\n", fullContent.Len())
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func wrapText(text string, width int) []string {
	var lines []string
	var currentLine strings.Builder
	
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		} else {
			currentLine.WriteRune(char)
			if currentLine.Len() >= width {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
		}
	}
	
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	
	return lines
}
