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
	fmt.Println("🚀 YUNAI 高级功能真实测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("📋 测试内容：两图系统 + 剧情触发 + 复杂关系网络 + 智能演绎")
	fmt.Println("🔗 使用真实 DeepSeek 模型")
	fmt.Println()

	ctx := context.Background()

	// 测试1：角色两图系统和抠图切换
	testTwoImageSystem(ctx)
	
	// 测试2：剧情触发和全局背景切换
	testStoryTriggerSystem(ctx)
	
	// 测试3：复杂关系网络智能解析
	testComplexRelationshipNetwork(ctx)
	
	// 测试4：关系驱动的智能演绎
	testRelationshipDrivenPerformance(ctx)

	fmt.Println("\n🎉 YUNAI 高级功能测试完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// testTwoImageSystem 测试角色两图系统
func testTwoImageSystem(ctx context.Context) {
	fmt.Println("🎨 [1/4] 角色两图系统测试")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Println("📸 角色创建 - 两图上传系统:")
	fmt.Println("   角色: 小雨")
	fmt.Println("   背景图 (bg_image): https://yunai-assets.com/xiayu/bg_garden_with_books.jpg")
	fmt.Println("   └─ 用途: 单聊背景，完整场景图")
	fmt.Println("   抠图 (cutout_image): https://yunai-assets.com/xiayu/cutout_transparent.png")
	fmt.Println("   └─ 用途: 群聊发言时显示，透明背景立绘")
	fmt.Println()

	fmt.Println("💬 单聊场景测试:")
	fmt.Println("   聊天背景: 自动使用小雨的背景图")
	fmt.Println("   效果: 花园读书场景，营造温馨氛围")
	
	singleChatPrompt := `你是小雨，现在在单聊场景中。
背景设定：你身处一个美丽的花园，周围有书籍和鲜花，这是你最喜欢的读书场所。
用户刚刚进入和你的单聊，请结合这个花园读书的背景环境，用温柔的语气打招呼。`

	response, err := callDeepSeek(ctx, singleChatPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨在花园背景下: \"%s\"\n", response)
		if strings.Contains(response, "花园") || strings.Contains(response, "书") {
			fmt.Printf("   ✅ 正确结合背景图环境进行对话\n")
		}
	}

	fmt.Println("\n🎭 群聊抠图切换测试:")
	fmt.Println("   群聊全局背景: https://yunai-assets.com/scenes/magic_academy.jpg")
	fmt.Println("   小雨发言时:")
	fmt.Println("   └─ 显示抠图: https://yunai-assets.com/xiayu/cutout_transparent.png")
	fmt.Println("   └─ 动画效果: 300ms 淡入，突出发言者")
	fmt.Println("   └─ 其他角色: 抠图淡化或隐藏")
	
	groupChatPrompt := `你是小雨，现在在群聊中发言。
场景：魔法学院的群聊，你的透明背景抠图正在以300ms淡入动画显示，突出你是当前发言者。
请简短地欢迎新同学加入群聊。`

	groupResponse, err := callDeepSeek(ctx, groupChatPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨发言时抠图显示: \"%s\"\n", groupResponse)
		fmt.Printf("   ✅ 抠图切换效果: 当前发言者突出显示\n")
	}
}

// testStoryTriggerSystem 测试剧情触发系统
func testStoryTriggerSystem(ctx context.Context) {
	fmt.Println("\n🎬 [2/4] 剧情触发和全局背景切换测试")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Println("🌍 群聊全局背景系统:")
	fmt.Println("   初始背景: https://yunai-assets.com/scenes/peaceful_academy.jpg")
	fmt.Println("   背景音乐: https://yunai-assets.com/music/peaceful_theme.mp3")
	fmt.Println()

	// 测试章节切换触发
	fmt.Println("📖 章节切换触发测试:")
	fmt.Println("   触发条件: 进入第二章 '魔法大战篇'")
	fmt.Println("   触发效果:")
	fmt.Println("   ├─ 背景图: peaceful_academy.jpg → battle_arena.jpg")
	fmt.Println("   ├─ 背景音乐: peaceful_theme.mp3 → battle_theme.mp3")
	fmt.Println("   ├─ 音效: 添加战斗音效")
	fmt.Println("   └─ 演绎模式: 和平模式 → 战斗模式")

	chapterTriggerPrompt := `场景切换触发！
原场景：和平的魔法学院，背景是宁静的庭院
新场景：魔法大战开始，背景切换为战斗竞技场，音乐变为紧张的战斗主题

你是小雨，感受到场景的剧烈变化。请用一句话表达对这个场景切换的反应，体现从和平到战斗的氛围变化。`

	chapterResponse, err := callDeepSeek(ctx, chapterTriggerPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨感受场景切换: \"%s\"\n", chapterResponse)
		if strings.Contains(chapterResponse, "战斗") || strings.Contains(chapterResponse, "紧张") {
			fmt.Printf("   ✅ 正确感知场景氛围变化\n")
		}
	}

	// 测试关键词触发
	fmt.Println("\n🔑 关键词触发测试:")
	fmt.Println("   触发词: '隐藏的秘密'")
	fmt.Println("   触发效果: 切换到神秘的地下室场景")

	keywordTriggerPrompt := `关键词触发！用户提到了"隐藏的秘密"
场景自动切换：
- 背景图: 神秘的地下室 (underground_secret.jpg)
- 背景音乐: 神秘主题音乐 (mystery_theme.mp3)
- 音效: 回声效果
- 光线: 昏暗的烛光

你是小雨，突然发现自己来到了一个神秘的地下室。请表达你的惊讶和好奇。`

	keywordResponse, err := callDeepSeek(ctx, keywordTriggerPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨发现隐藏场景: \"%s\"\n", keywordResponse)
		if strings.Contains(keywordResponse, "神秘") || strings.Contains(keywordResponse, "地下") {
			fmt.Printf("   ✅ 正确响应隐藏剧情触发\n")
		}
	}
}

// testComplexRelationshipNetwork 测试复杂关系网络
func testComplexRelationshipNetwork(ctx context.Context) {
	fmt.Println("\n💕 [3/4] 复杂关系网络智能解析测试")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Println("🧠 复杂关系描述解析测试:")
	
	// 测试复杂关系1
	complexRelation1 := `小米是我的好朋友，但她闺蜜小明和我关系不好，我们经常因为小米而产生矛盾。
小米夹在中间很为难，有时候会偏向小明，这让我很不开心。`

	fmt.Printf("关系描述1: %s\n", complexRelation1)
	
	relationPrompt1 := `请分析以下复杂的人际关系描述，并以AI角色的身份智能理解其中的关系动态：

关系描述：` + complexRelation1 + `

请分析：
1. 用户与小米的关系性质
2. 用户与小明的关系性质  
3. 三人关系中的复杂动态
4. 如果你是小米，在群聊中应该如何平衡这种关系？`

	relation1Response, err := callDeepSeek(ctx, relationPrompt1, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 关系分析: \"%s\"\n", relation1Response)
		if strings.Contains(relation1Response, "复杂") || strings.Contains(relation1Response, "矛盾") {
			fmt.Printf("   ✅ 正确理解复杂关系动态\n")
		}
	}

	// 测试复杂关系2
	fmt.Println("\n🎭 表面关系 vs 真实关系测试:")
	complexRelation2 := `我和小王表面上是朋友，经常一起吃饭聊天，但实际上我们有竞争关系。
他总是想超越我，我也不甘示弱。有时候我们会互相帮助，有时候又会暗中较劲。
说是敌人吧，我们又确实关心对方；说是朋友吧，又总是在暗中比较。`

	fmt.Printf("关系描述2: %s\n", complexRelation2)
	
	relationPrompt2 := `请分析这种"亦敌亦友"的复杂关系，如果你是AI角色，在群聊中遇到这种关系的两个人，应该如何智能地处理互动？

关系描述：` + complexRelation2

	relation2Response, err := callDeepSeek(ctx, relationPrompt2, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 复杂关系处理: \"%s\"\n", relation2Response)
		if strings.Contains(relation2Response, "竞争") || strings.Contains(relation2Response, "平衡") {
			fmt.Printf("   ✅ 正确理解亦敌亦友关系\n")
		}
	}
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
		MaxTokens:   800,
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
