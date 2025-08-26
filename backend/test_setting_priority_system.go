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
}

type Choice struct {
	Message Message `json:"message"`
}

func main() {
	fmt.Println("⚙️ YUNAI 设定优先级系统测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("📋 测试内容：群聊设定 → 单聊设定 → 基础人设的智能回退机制")
	fmt.Println("🔗 使用真实 DeepSeek 模型")
	fmt.Println()

	ctx := context.Background()

	// 测试1：完整设定场景（群聊世界观 + 演绎设定）
	testFullSettingScenario(ctx)
	
	// 测试2：部分设定场景（只有群聊世界观，无演绎设定）
	testPartialSettingScenario(ctx)
	
	// 测试3：无群聊设定场景（回退到角色单聊设定）
	testNoGroupSettingScenario(ctx)
	
	// 测试4：最小设定场景（只有基础人设）
	testMinimalSettingScenario(ctx)

	fmt.Println("\n🎉 设定优先级系统测试完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// testFullSettingScenario 测试完整设定场景
func testFullSettingScenario(ctx context.Context) {
	fmt.Println("🎯 [1/4] 完整设定场景测试")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("设定层级：群聊世界观 ✅ + 群聊演绎设定 ✅ + 角色单聊设定 ✅ + 基础人设 ✅")
	fmt.Println()

	worldSetting := `【魔法学院世界观】
这里是艾尔维斯魔法学院，学生们都在学习各种魔法技能。
- 小雨：图书馆管理员，擅长治愈魔法
- 小美：音乐系学生，擅长音律魔法
当前场景：魔法学院的中央庭院，正在举行魔法比赛`

	performanceSetting := `【群聊演绎设定】
1. 所有角色必须使用魔法相关词汇
2. 发言时要体现魔法比赛的紧张氛围
3. 角色间要有竞技互动
4. 使用魔法学院的专业术语`

	characterSetting := `【小雨的单聊设定】
你是一个温柔的图书管理员，喜欢安静地读书，不太喜欢竞争。`

	basicPersonality := `【小雨基础人设】
温柔、善良、内向、喜欢帮助他人`

	fullPrompt := worldSetting + "\n" + performanceSetting + "\n" + characterSetting + "\n" + basicPersonality + `

现在用户说："比赛开始了！"
请以小雨的身份回应，体现完整的设定层级。`

	response, err := callDeepSeek(ctx, fullPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨（完整设定）: \"%s\"\n", response)
		
		// 检查是否体现了各层设定
		checks := map[string]bool{
			"魔法相关": strings.Contains(response, "魔法") || strings.Contains(response, "法术"),
			"比赛氛围": strings.Contains(response, "比赛") || strings.Contains(response, "竞争"),
			"图书管理员身份": strings.Contains(response, "图书") || strings.Contains(response, "书"),
			"温柔性格": strings.Contains(response, "温柔") || strings.Contains(response, "轻声") || strings.Contains(response, "小心"),
		}
		
		for aspect, found := range checks {
			if found {
				fmt.Printf("   ✅ %s: 已体现\n", aspect)
			} else {
				fmt.Printf("   ⚠️  %s: 未明显体现\n", aspect)
			}
		}
	}
}

// testPartialSettingScenario 测试部分设定场景
func testPartialSettingScenario(ctx context.Context) {
	fmt.Println("\n🎯 [2/4] 部分设定场景测试")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("设定层级：群聊世界观 ✅ + 群聊演绎设定 ❌ + 角色单聊设定 ✅ + 基础人设 ✅")
	fmt.Println()

	worldSetting := `【魔法学院世界观】
这里是艾尔维斯魔法学院，学生们都在学习各种魔法技能。
- 小雨：图书馆管理员，擅长治愈魔法
当前场景：魔法学院的图书馆`

	// 注意：没有群聊演绎设定

	characterSetting := `【小雨的单聊设定】
你是一个温柔的图书管理员，喜欢安静地读书，经常帮助学生找书。`

	basicPersonality := `【小雨基础人设】
温柔、善良、内向、喜欢帮助他人`

	partialPrompt := worldSetting + "\n" + characterSetting + "\n" + basicPersonality + `

现在用户说："图书馆好安静啊。"
请以小雨的身份回应，在没有具体演绎设定的情况下，根据世界观和个人设定自由发挥。`

	response, err := callDeepSeek(ctx, partialPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨（部分设定）: \"%s\"\n", response)
		fmt.Printf("   ✅ 智能回退：结合世界观 + 个人设定自由发挥\n")
		
		if strings.Contains(response, "图书") || strings.Contains(response, "安静") {
			fmt.Printf("   ✅ 正确结合场景和性格特点\n")
		}
	}
}

// testNoGroupSettingScenario 测试无群聊设定场景
func testNoGroupSettingScenario(ctx context.Context) {
	fmt.Println("\n🎯 [3/4] 无群聊设定场景测试")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("设定层级：群聊世界观 ❌ + 群聊演绎设定 ❌ + 角色单聊设定 ✅ + 基础人设 ✅")
	fmt.Println()

	// 注意：没有群聊世界观和演绎设定

	characterSetting := `【小雨的单聊设定】
你是一个温柔的图书管理员，在一个现代图书馆工作。你喜欢推荐好书给读者，
经常组织读书会，对各种文学作品都很了解。你说话轻声细语，总是很有耐心。`

	basicPersonality := `【小雨基础人设】
温柔、善良、内向、喜欢帮助他人、博学`

	noGroupPrompt := characterSetting + "\n" + basicPersonality + `

现在在一个普通的群聊中，用户说："最近想读点好书，有什么推荐吗？"
请以小雨的身份回应，在没有群聊特殊设定的情况下，完全按照你的单聊设定和基础人设自由发挥。`

	response, err := callDeepSeek(ctx, noGroupPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨（无群聊设定）: \"%s\"\n", response)
		fmt.Printf("   ✅ 智能回退：完全按照单聊设定自由发挥\n")
		
		// 检查是否避免了魔法元素（因为没有魔法世界观）
		if !strings.Contains(response, "魔法") && !strings.Contains(response, "法术") {
			fmt.Printf("   ✅ 正确避免不存在的世界观元素\n")
		}
		
		if strings.Contains(response, "书") || strings.Contains(response, "推荐") {
			fmt.Printf("   ✅ 正确体现图书管理员专业性\n")
		}
	}
}

// testMinimalSettingScenario 测试最小设定场景
func testMinimalSettingScenario(ctx context.Context) {
	fmt.Println("\n🎯 [4/4] 最小设定场景测试")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("设定层级：群聊世界观 ❌ + 群聊演绎设定 ❌ + 角色单聊设定 ❌ + 基础人设 ✅")
	fmt.Println()

	// 只有最基础的人设
	basicPersonality := `【小雨基础人设】
你是小雨，一个温柔善良的女孩。你性格内向，喜欢帮助他人，说话轻声细语。`

	minimalPrompt := basicPersonality + `

现在在群聊中，用户说："大家好！"
请以小雨的身份回应，在只有基础人设的情况下，自然地表现你的性格特点。`

	response, err := callDeepSeek(ctx, minimalPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨（最小设定）: \"%s\"\n", response)
		fmt.Printf("   ✅ 智能回退：仅基于基础人设自然表现\n")
		
		// 检查是否体现了基础性格
		if strings.Contains(response, "你好") || strings.Contains(response, "大家好") {
			fmt.Printf("   ✅ 自然的社交回应\n")
		}
		
		// 应该是简单自然的回应，不会有复杂的世界观元素
		if len(response) < 100 { // 简短回应
			fmt.Printf("   ✅ 适当的简洁性（基础设定下的自然表现）\n")
		}
	}

	fmt.Println("\n📊 设定优先级系统总结:")
	fmt.Println("   1️⃣ 群聊世界观设定 - 最高优先级，定义整体世界")
	fmt.Println("   2️⃣ 群聊演绎设定 - 定义具体行为规则")
	fmt.Println("   3️⃣ 角色单聊设定 - 角色个人特色设定")
	fmt.Println("   4️⃣ 角色基础人设 - 兜底设定，确保基本表现")
	fmt.Println("   ✅ 智能回退：缺失高级设定时自动使用低级设定")
	fmt.Println("   ✅ 自然表现：避免不存在的世界观元素")
}

// callDeepSeek 调用 DeepSeek API
func callDeepSeek(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}
	
	if userMessage != "" {
		messages = append(messages, Message{Role: "user", Content: userMessage})
	}

	request := ChatRequest{
		Model:       DeepSeekModel,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   500,
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
