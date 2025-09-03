package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// 聊天请求结构
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

const BASE_URL = "http://localhost:8080/api/v1"

func main() {
	fmt.Println("🧪 YUNAI AI对话功能修复测试")
	fmt.Println("===========================================")

	// 1. 首先获取真实的模型ID
	fmt.Println("📋 步骤1: 获取真实模型ID...")
	modelID, err := getRealModelID()
	if err != nil {
		fmt.Printf("❌ 获取模型ID失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 获取到模型ID: %s\n", modelID)

	// 2. 测试非流式对话
	fmt.Println("\n📝 步骤2: 测试非流式对话...")
	testNonStreamingChat(modelID)

	// 3. 测试流式对话
	fmt.Println("\n🌊 步骤3: 测试流式对话...")
	testStreamingChat(modelID)

	fmt.Println("\n🎉 AI对话功能测试完成！")
}

// 获取真实的模型ID
func getRealModelID() (string, error) {
	resp, err := http.Get(BASE_URL + "/models?type=chat")
	if err != nil {
		return "", fmt.Errorf("请求模型列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("获取模型列表失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析JSON失败: %w", err)
	}

	// 提取第一个模型的ID
	if data, ok := result["data"].(map[string]interface{}); ok {
		if models, ok := data["models"].([]interface{}); ok && len(models) > 0 {
			if model, ok := models[0].(map[string]interface{}); ok {
				if id, ok := model["id"].(string); ok {
					return id, nil
				}
			}
		}
	}

	return "", fmt.Errorf("未找到可用的模型")
}

// 测试非流式对话
func testNonStreamingChat(modelID string) {
	fmt.Printf("🤖 使用模型: %s\n", modelID)
	fmt.Printf("❓ 测试问题: 你好，请简单介绍一下自己\n")

	chatData := ChatRequest{
		ModelID: modelID,
		Messages: []Message{
			{Role: "user", Content: "你好，请简单介绍一下自己"},
		},
		Stream:      false,
		Temperature: 0.7,
		MaxTokens:   200,
	}

	jsonData, err := json.Marshal(chatData)
	if err != nil {
		fmt.Printf("❌ 构建请求失败: %v\n", err)
		return
	}

	start := time.Now()
	resp, err := http.Post(BASE_URL+"/chat/completions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(start)

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
	var chatResp map[string]interface{}
	if err := json.Unmarshal(body, &chatResp); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("   原始响应: %s\n", string(body))
		return
	}

	// 显示结果
	if choices, ok := chatResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					fmt.Printf("⏱️ 响应时间: %v\n", duration.Round(time.Millisecond))

					if usage, ok := chatResp["usage"].(map[string]interface{}); ok {
						fmt.Printf("📊 Token使用: %.0f输入 + %.0f输出 = %.0f总计\n",
							usage["prompt_tokens"], usage["completion_tokens"], usage["total_tokens"])
					}

					fmt.Printf("📏 响应长度: %d 字符\n", len(content))
					fmt.Printf("💬 AI回答:\n")
					fmt.Printf("┌%s┐\n", "─────────────────────────────────────────────────────────")

					// 简单的文本换行处理
					lines := wrapText(content, 55)
					for _, line := range lines {
						fmt.Printf("│ %-55s │\n", line)
					}

					fmt.Printf("└%s┘\n", "─────────────────────────────────────────────────────────")

					fmt.Printf("✅ 非流式对话测试成功！\n")
					return
				}
			}
		}
	}

	fmt.Printf("❌ 响应格式异常\n")
	fmt.Printf("   响应内容: %s\n", string(body))
}

// 测试流式对话
func testStreamingChat(modelID string) {
	fmt.Printf("🤖 使用模型: %s\n", modelID)
	fmt.Printf("❓ 测试问题: 请用3句话介绍人工智能的发展历程\n")

	chatData := ChatRequest{
		ModelID: modelID,
		Messages: []Message{
			{Role: "user", Content: "请用3句话介绍人工智能的发展历程"},
		},
		Stream:      true,
		Temperature: 0.7,
		MaxTokens:   300,
	}

	jsonData, err := json.Marshal(chatData)
	if err != nil {
		fmt.Printf("❌ 构建请求失败: %v\n", err)
		return
	}

	start := time.Now()
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
	fmt.Printf("┌%s┐\n", "─────────────────────────────────────────────────────────")
	fmt.Print("│ ")

	// 读取流式响应
	buffer := make([]byte, 4096)
	var fullContent strings.Builder
	chunkCount := 0
	firstChunkTime := time.Time{}
	currentLineLength := 0

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			chunk := string(buffer[:n])
			lines := strings.Split(chunk, "\n")

			for _, line := range lines {
				if !strings.HasPrefix(line, "data: ") {
					continue
				}

				jsonStr := strings.TrimPrefix(line, "data: ")
				if jsonStr == "[DONE]" {
					goto done
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

								// 实时显示
								for _, char := range content {
									if char == '\n' {
										fmt.Printf("%s │\n│ ", strings.Repeat(" ", 55-currentLineLength))
										currentLineLength = 0
									} else {
										fmt.Print(string(char))
										currentLineLength++
										if currentLineLength >= 55 {
											fmt.Printf(" │\n│ ")
											currentLineLength = 0
										}
									}
								}

								time.Sleep(20 * time.Millisecond) // 模拟打字效果
							}
						}
					}
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("\n❌ 读取流式响应出错: %v\n", err)
			return
		}
	}

done:
	// 补齐最后一行
	if currentLineLength > 0 {
		fmt.Printf("%s │\n", strings.Repeat(" ", 55-currentLineLength))
	}
	fmt.Printf("└%s┘\n", "─────────────────────────────────────────────────────────")

	totalDuration := time.Since(start)
	var firstChunkDuration time.Duration
	if !firstChunkTime.IsZero() {
		firstChunkDuration = firstChunkTime.Sub(start)
	}

	fmt.Printf("⏱️ 总响应时间: %v\n", totalDuration.Round(time.Millisecond))
	if firstChunkDuration > 0 {
		fmt.Printf("🚀 首字符时间: %v (TTFB)\n", firstChunkDuration.Round(time.Millisecond))
	}
	fmt.Printf("📦 流式块数量: %d\n", chunkCount)
	fmt.Printf("📏 响应长度: %d 字符\n", fullContent.Len())

	if chunkCount > 0 {
		fmt.Printf("✅ 流式对话测试成功！\n")
	} else {
		fmt.Printf("⚠️ 流式对话未收到有效内容\n")
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
