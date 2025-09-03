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

func main() {
	fmt.Println("🤖 YUNAI直接对话测试 - 流式 vs 非流式")
	fmt.Println("===========================================")

	// SiliconFlow API配置
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"
	baseURL := "https://api.siliconflow.cn/v1/chat/completions"

	// 使用已知存在的模型
	testModel := "Qwen/Qwen2.5-7B-Instruct"
	fmt.Printf("🎯 测试模型: %s\n", testModel)

	// 测试对话内容
	testMessages := []Message{
		{Role: "user", Content: "请简单介绍一下人工智能的发展历史，大概150字左右。"},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 1. 测试非流式对话
	fmt.Println("📝 测试非流式对话")
	fmt.Println(strings.Repeat("-", 40))
	testNonStreamingChat(apiKey, baseURL, testModel, testMessages)

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 2. 测试流式对话
	fmt.Println("🌊 测试流式对话")
	fmt.Println(strings.Repeat("-", 40))
	testStreamingChat(apiKey, baseURL, testModel, testMessages)

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 3. 对比测试 - 更复杂的问题
	fmt.Println("🔄 对比测试 - 复杂问题")
	fmt.Println(strings.Repeat("-", 40))
	
	complexMessages := []Message{
		{Role: "user", Content: "请详细解释机器学习中的监督学习和无监督学习的区别，并各举一个实际应用例子。大概200字。"},
	}

	fmt.Println("📝 非流式 - 复杂问题:")
	testNonStreamingChat(apiKey, baseURL, testModel, complexMessages)

	fmt.Println("\n🌊 流式 - 复杂问题:")
	testStreamingChat(apiKey, baseURL, testModel, complexMessages)

	fmt.Println("\n🎉 直接对话测试完成！")
}

// 测试非流式对话
func testNonStreamingChat(apiKey, baseURL, modelKey string, messages []Message) {
	fmt.Printf("🤖 模型: %s\n", modelKey)
	fmt.Printf("❓ 问题: %s\n", messages[0].Content)
	fmt.Printf("⏱️ 开始时间: %s\n", time.Now().Format("15:04:05"))
	
	// 构建请求
	request := ChatRequest{
		Model:       modelKey,
		Messages:    messages,
		Stream:      false,
		Temperature: 0.7,
		MaxTokens:   400,
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
		fmt.Printf("⏱️ 结束时间: %s\n", time.Now().Format("15:04:05"))
		fmt.Printf("⚡ 总响应时间: %v\n", duration)
		fmt.Printf("📊 Token使用: %d输入 + %d输出 = %d总计\n", 
			chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)
		fmt.Printf("📏 响应长度: %d 字符\n", len(content))
		fmt.Printf("💬 AI回答:\n")
		fmt.Printf("┌%s┐\n", strings.Repeat("─", 70))
		
		// 格式化输出，每行最多68字符
		lines := wrapText(content, 68)
		for _, line := range lines {
			fmt.Printf("│ %-68s │\n", line)
		}
		
		fmt.Printf("└%s┘\n", strings.Repeat("─", 70))
	} else {
		fmt.Printf("❌ 响应中没有内容\n")
	}
}

// 测试流式对话
func testStreamingChat(apiKey, baseURL, modelKey string, messages []Message) {
	fmt.Printf("🤖 模型: %s\n", modelKey)
	fmt.Printf("❓ 问题: %s\n", messages[0].Content)
	fmt.Printf("⏱️ 开始时间: %s\n", time.Now().Format("15:04:05"))
	
	// 构建请求
	request := ChatRequest{
		Model:       modelKey,
		Messages:    messages,
		Stream:      true,
		Temperature: 0.7,
		MaxTokens:   400,
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
	fmt.Printf("💬 AI回答 (实时流式输出):\n")
	fmt.Printf("┌%s┐\n", strings.Repeat("─", 70))
	fmt.Print("│ ")

	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	chunkCount := 0
	firstChunkTime := time.Time{}
	currentLineLength := 0

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
		var streamResp map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &streamResp); err != nil {
			continue
		}

		// 处理内容
		if choices, ok := streamResp["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok && content != "" {
						if chunkCount == 0 {
							firstChunkTime = time.Now()
						}
						chunkCount++
						fullContent.WriteString(content)
						
						// 实时显示流式输出，处理换行
						for _, char := range content {
							if char == '\n' {
								// 换行
								fmt.Printf("%s │\n│ ", strings.Repeat(" ", 68-currentLineLength))
								currentLineLength = 0
							} else {
								fmt.Print(string(char))
								currentLineLength++
								if currentLineLength >= 68 {
									// 自动换行
									fmt.Printf(" │\n│ ")
									currentLineLength = 0
								}
							}
						}
						
						// 添加小延迟模拟真实打字效果
						time.Sleep(15 * time.Millisecond)
					}
				}
			}
		}
	}

	// 补齐最后一行
	if currentLineLength > 0 {
		fmt.Printf("%s │\n", strings.Repeat(" ", 68-currentLineLength))
	}
	fmt.Printf("└%s┘\n", strings.Repeat("─", 70))

	totalDuration := time.Since(startTime)
	firstChunkDuration := firstChunkTime.Sub(startTime)

	fmt.Printf("⏱️ 结束时间: %s\n", time.Now().Format("15:04:05"))
	fmt.Printf("⚡ 总响应时间: %v\n", totalDuration)
	fmt.Printf("🚀 首字符时间: %v (TTFB - Time To First Byte)\n", firstChunkDuration)
	fmt.Printf("📦 流式块数量: %d\n", chunkCount)
	fmt.Printf("📏 响应长度: %d 字符\n", fullContent.Len())

	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ 读取流式响应出错: %v\n", err)
	}
}

// 文本换行处理
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
