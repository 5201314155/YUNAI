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

const BASE_URL = "http://localhost:8080/api/v1"

func main() {
	fmt.Println("🧪 YUNAI双API系统测试 - 管理员 vs 用户视图")
	fmt.Println("===========================================")

	// 等待API服务启动
	fmt.Println("⏳ 等待API服务启动...")
	time.Sleep(3 * time.Second)

	// 1. 获取系统信息
	fmt.Println("📊 获取系统信息...")
	getSystemInfo()

	// 2. 测试管理员API (显示真实模型名称)
	fmt.Println("\n🔧 管理员API - 显示真实模型名称")
	fmt.Println(strings.Repeat("-", 50))
	_ = getAdminModels()

	// 3. 测试用户API (显示友好名称)
	fmt.Println("\n👤 用户API - 显示友好名称")
	fmt.Println(strings.Repeat("-", 50))
	userModels := getUserModels()

	// 4. 选择一个模型进行对话测试
	if len(userModels) == 0 {
		fmt.Println("❌ 没有找到可用的对话模型")
		return
	}

	// 找到一个支持流式的模型
	var testModel map[string]interface{}
	for _, model := range userModels {
		if modelMap, ok := model.(map[string]interface{}); ok {
			if modelType, ok := modelMap["type"].(string); ok && modelType == "chat" {
				if supportStreaming, ok := modelMap["support_streaming"].(bool); ok && supportStreaming {
					testModel = modelMap
					break
				}
			}
		}
	}

	if testModel == nil {
		fmt.Println("❌ 没有找到支持流式的对话模型")
		return
	}

	modelID := testModel["id"].(string)
	modelName := testModel["name"].(string)

	fmt.Printf("\n🎯 选择测试模型: %s (ID: %s)\n", modelName, modelID)

	// 测试对话内容
	testMessages := []Message{
		{Role: "user", Content: "请简单介绍一下人工智能的三个主要应用领域，每个领域用一句话概括。"},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 5. 测试非流式对话
	fmt.Println("📝 测试非流式对话")
	fmt.Println(strings.Repeat("-", 40))
	testNonStreamingChat(modelID, testMessages)

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 6. 测试流式对话
	fmt.Println("🌊 测试流式对话")
	fmt.Println(strings.Repeat("-", 40))
	testStreamingChat(modelID, testMessages)

	fmt.Println("\n🎉 双API系统测试完成！")
	fmt.Println("✅ 管理员看到真实模型名称，用户看到友好名称")
	fmt.Println("✅ 对话时自动转换为真实模型名称调用SiliconFlow")
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
	}
}

// 获取管理员模型列表 (显示真实模型名称)
func getAdminModels() []interface{} {
	resp, err := http.Get(BASE_URL + "/admin/models?type=chat")
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
			fmt.Printf("✅ 管理员视图: 找到 %d 个对话模型\n", len(modelsData))

			for i, modelData := range modelsData {
				if i >= 5 { // 只显示前5个
					fmt.Printf("   ... 还有 %d 个模型\n", len(modelsData)-5)
					break
				}

				if modelMap, ok := modelData.(map[string]interface{}); ok {
					internalKey := getString(modelMap, "internal_key")
					customName := getString(modelMap, "custom_display_name")
					provider := getString(modelMap, "provider")

					streamIcon := "❌"
					if getBool(modelMap, "support_streaming") {
						streamIcon = "🌊"
					}

					fmt.Printf("   %d. %s %s [%s]\n", i+1, streamIcon, customName, provider)
					fmt.Printf("      真实模型ID: %s\n", internalKey)
				}
			}

			return modelsData
		}
	}

	return nil
}

// 获取用户模型列表 (显示友好名称)
func getUserModels() []interface{} {
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
			fmt.Printf("✅ 用户视图: 找到 %d 个对话模型\n", len(modelsData))

			for i, modelData := range modelsData {
				if i >= 5 { // 只显示前5个
					fmt.Printf("   ... 还有 %d 个模型\n", len(modelsData)-5)
					break
				}

				if modelMap, ok := modelData.(map[string]interface{}); ok {
					name := getString(modelMap, "name")
					provider := getString(modelMap, "provider")

					streamIcon := "❌"
					if getBool(modelMap, "support_streaming") {
						streamIcon = "🌊"
					}

					featuredIcon := ""
					if getBool(modelMap, "is_featured") {
						featuredIcon = "⭐"
					}

					fmt.Printf("   %d. %s %s %s [%s]\n", i+1, streamIcon, featuredIcon, name, provider)
					fmt.Printf("      (真实模型ID对用户隐藏)\n")
				}
			}

			return modelsData
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
		MaxTokens:   300,
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

	var chatResp map[string]interface{}
	if err := json.Unmarshal(body, &chatResp); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}

	if choices, ok := chatResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					fmt.Printf("⏱️ 结束时间: %s\n", time.Now().Format("15:04:05"))
					fmt.Printf("⚡ 总响应时间: %v\n", duration)

					if usage, ok := chatResp["usage"].(map[string]interface{}); ok {
						fmt.Printf("📊 Token使用: %.0f输入 + %.0f输出 = %.0f总计\n",
							usage["prompt_tokens"], usage["completion_tokens"], usage["total_tokens"])
					}

					fmt.Printf("📏 响应长度: %d 字符\n", len(content))
					fmt.Printf("💬 AI回答:\n")
					fmt.Printf("┌%s┐\n", strings.Repeat("─", 60))

					lines := wrapText(content, 58)
					for _, line := range lines {
						fmt.Printf("│ %-58s │\n", line)
					}

					fmt.Printf("└%s┘\n", strings.Repeat("─", 60))
				}
			}
		}
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
		MaxTokens:   300,
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
