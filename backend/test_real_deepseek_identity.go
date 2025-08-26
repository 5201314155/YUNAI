package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DeepSeek API 配置
const (
	DeepSeekAPIBase = "https://api.siliconflow.cn/v1"
	DeepSeekAPIKey  = "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"
	DeepSeekModel   = "deepseek-ai/DeepSeek-V3"
)

// API 请求结构
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func main() {
	fmt.Println("🤖 YUNAI DeepSeek 真实模型身份识别测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("🔗 使用真实 DeepSeek API: deepseek测试专用")
	fmt.Println()

	ctx := context.Background()

	// 测试场景1：指定身份角色
	testSpecifiedIdentity(ctx)
	
	// 测试场景2：通用用户角色
	testGenericUserIdentity(ctx)
	
	// 测试场景3：朋友圈生成
	testMomentsGeneration(ctx)
	
	// 测试场景4：@提及处理
	testMentionHandling(ctx)

	fmt.Println("\n🎉 DeepSeek 真实模型测试完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// testSpecifiedIdentity 测试指定身份场景
func testSpecifiedIdentity(ctx context.Context) {
	fmt.Println("🎭 [1/4] 指定身份测试 - 角色：小雨，用户：李明")
	fmt.Println(strings.Repeat("-", 70))

	systemPrompt := `你是小雨，一个温柔善良的AI助手。用户（李明）是你最好的朋友，你们经常一起聊天谈心。
你的性格特点：
- 温柔体贴，善解人意
- 喜欢读书和画画
- 对李明很关心，会主动询问他的近况
- 说话温和，经常使用温馨的表情符号

请记住，李明是你的好朋友，在对话中要体现出你们的亲密关系。`

	testQuestions := []string{
		"我叫什么名字？",
		"我是谁？",
		"我和你什么关系？",
		"你还记得我吗？",
	}

	for i, question := range testQuestions {
		fmt.Printf("\n测试 %d: \"%s\"\n", i+1, question)
		
		response, err := callDeepSeek(ctx, systemPrompt, question)
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
			continue
		}
		
		fmt.Printf("   🤖 DeepSeek 回答: \"%s\"\n", response)
		
		// 检查是否正确识别了李明这个身份
		if strings.Contains(response, "李明") {
			fmt.Printf("   ✅ 正确识别指定身份: 李明\n")
		} else {
			fmt.Printf("   ⚠️  未明确提及指定身份\n")
		}
	}
}

// testGenericUserIdentity 测试通用用户身份场景
func testGenericUserIdentity(ctx context.Context) {
	fmt.Println("\n👤 [2/4] 通用用户测试 - 角色：小美，用户：小王（真实昵称）")
	fmt.Println(strings.Repeat("-", 70))

	systemPrompt := `你是小美，一个活泼开朗的AI助手。用户是你的朋友，你喜欢和用户一起玩耍聊天。
你的性格特点：
- 活泼开朗，充满活力
- 喜欢游戏和音乐
- 对用户很友好，会主动分享有趣的事情
- 说话轻松愉快，经常使用可爱的表情符号

请记住，要根据用户的真实昵称来称呼用户，让对话更加亲切自然。用户的昵称是小王。`

	testQuestions := []string{
		"我叫什么名字？",
		"我是谁？",
		"我和你什么关系？",
		"你认识我吗？",
	}

	for i, question := range testQuestions {
		fmt.Printf("\n测试 %d: \"%s\"\n", i+1, question)
		
		response, err := callDeepSeek(ctx, systemPrompt, question)
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
			continue
		}
		
		fmt.Printf("   🤖 DeepSeek 回答: \"%s\"\n", response)
		
		// 检查是否正确使用了小王这个真实昵称
		if strings.Contains(response, "小王") {
			fmt.Printf("   ✅ 正确使用真实昵称: 小王\n")
		} else {
			fmt.Printf("   ⚠️  未明确提及真实昵称\n")
		}
	}
}

// testMomentsGeneration 测试朋友圈生成
func testMomentsGeneration(ctx context.Context) {
	fmt.Println("\n📱 [3/4] 朋友圈生成测试")
	fmt.Println(strings.Repeat("-", 70))

	// 指定身份朋友圈生成
	fmt.Println("场景1: 小雨（指定身份：李明）生成朋友圈")
	systemPrompt1 := `你是小雨，李明是你的好朋友。请根据你们最近的聊天内容，生成3条朋友圈动态。
最近聊天内容：
- 李明: "今天天气真不错呢！"
- 小雨: "是啊，李明！这样的好天气最适合出去走走了~ ☀️"
- 李明: "我刚刚读了一本很有趣的书"

请生成3条朋友圈内容，要体现出你和李明的友谊。`

	response1, err := callDeepSeek(ctx, systemPrompt1, "请生成朋友圈内容")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 生成:\n%s\n", response1)
		if strings.Contains(response1, "李明") {
			fmt.Printf("   ✅ 正确使用指定身份: 李明\n")
		}
	}

	// 通用用户朋友圈生成
	fmt.Println("\n场景2: 小美（真实昵称：小王）生成朋友圈")
	systemPrompt2 := `你是小美，小王是你的朋友。请根据你们最近的聊天内容，生成3条朋友圈动态。
最近聊天内容：
- 小王: "今天玩了新游戏，很有趣！"
- 小美: "哇！小王的游戏技术越来越好了~ 🎮"
- 小王: "我们一起组队吧"

请生成3条朋友圈内容，要体现出你和小王的友谊。`

	response2, err := callDeepSeek(ctx, systemPrompt2, "请生成朋友圈内容")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 生成:\n%s\n", response2)
		if strings.Contains(response2, "小王") {
			fmt.Printf("   ✅ 正确使用真实昵称: 小王\n")
		}
	}
}

// testMentionHandling 测试@提及处理
func testMentionHandling(ctx context.Context) {
	fmt.Println("\n@ [4/4] @提及处理测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试@用户的替换
	fmt.Println("场景1: 指定身份上下文中的@用户替换")
	systemPrompt1 := `你是小雨，李明是你的好朋友。当你在朋友圈中提到"@用户"时，应该替换为"@李明"。
请将以下内容中的"@用户"替换为正确的身份：
"@用户 今天天气真好呢！想和你一起出去走走~"`

	response1, err := callDeepSeek(ctx, systemPrompt1, "请处理@提及")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 处理结果: \"%s\"\n", response1)
		if strings.Contains(response1, "@李明") {
			fmt.Printf("   ✅ 正确替换为指定身份: @李明\n")
		}
	}

	fmt.Println("\n场景2: 真实昵称上下文中的@用户替换")
	systemPrompt2 := `你是小美，小王是你的朋友。当你在朋友圈中提到"@用户"时，应该替换为"@小王"。
请将以下内容中的"@用户"替换为正确的身份：
"@用户 新游戏很好玩哦！一起来试试吧~"`

	response2, err := callDeepSeek(ctx, systemPrompt2, "请处理@提及")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 处理结果: \"%s\"\n", response2)
		if strings.Contains(response2, "@小王") {
			fmt.Printf("   ✅ 正确替换为真实昵称: @小王\n")
		}
	}
}

// callDeepSeek 调用 DeepSeek API
func callDeepSeek(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	request := ChatRequest{
		Model: DeepSeekModel,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("JSON marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", DeepSeekAPIBase+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DeepSeekAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("JSON unmarshal error: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices")
	}

	return response.Choices[0].Message.Content, nil
}
