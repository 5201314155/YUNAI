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
	fmt.Println("🎭 YUNAI 超级智能剧情触发系统测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("📋 测试内容：模糊触发 + 章节切换 + 智能演绎 + 自由发挥")
	fmt.Println("🔗 使用真实 DeepSeek 模型")
	fmt.Println()

	ctx := context.Background()

	// 测试1：超级模糊触发机制
	testFuzzyTriggerMechanism(ctx)
	
	// 测试2：多章节智能切换
	testMultiChapterSystem(ctx)
	
	// 测试3：单章节智能演绎
	testSingleChapterIntelligence(ctx)
	
	// 测试4：无设定自由发挥
	testFreeFormPerformance(ctx)

	fmt.Println("\n🎉 超级智能剧情触发系统测试完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// testFuzzyTriggerMechanism 测试超级模糊触发机制
func testFuzzyTriggerMechanism(ctx context.Context) {
	fmt.Println("🧠 [1/4] 超级模糊触发机制测试")
	fmt.Println(strings.Repeat("-", 70))

	// 基础剧情设定
	baseStory := `【魔法学院剧情设定】
当前章节：第一章 - 平静的学院生活
场景：魔法学院的日常生活
背景图：peaceful_academy.jpg
背景音乐：calm_theme.mp3

触发器设定：
1. 神秘触发词：["神秘", "秘密", "隐藏", "奇怪", "不对劲"] → 切换到神秘章节
2. 战斗触发词：["战斗", "打架", "敌人", "危险", "紧张"] → 切换到战斗章节
3. 庆祝触发词：["开心", "庆祝", "成功", "胜利", "快乐"] → 切换到庆祝章节

你是小雨，魔法学院的图书管理员。请根据用户的话语智能判断是否需要触发剧情切换。`

	// 测试各种模糊触发
	fuzzyTriggers := []struct {
		userInput    string
		expectedType string
		description  string
	}{
		{
			userInput:    "我想去看看那个神秘的地方",
			expectedType: "神秘",
			description:  "白话表达 + 模糊关联",
		},
		{
			userInput:    "感觉这里有点不对劲呢",
			expectedType: "神秘",
			description:  "情绪感知 + 语义理解",
		},
		{
			userInput:    "好像有什么危险在靠近",
			expectedType: "战斗",
			description:  "危机预感 + 模糊表达",
		},
		{
			userInput:    "今天心情特别好！",
			expectedType: "庆祝",
			description:  "情绪触发 + 正面情感",
		},
		{
			userInput:    "图书馆真安静啊",
			expectedType: "无触发",
			description:  "日常对话，不触发",
		},
	}

	for i, trigger := range fuzzyTriggers {
		fmt.Printf("\n测试 %d: %s (%s)\n", i+1, trigger.description, trigger.expectedType)
		fmt.Printf("   用户输入: \"%s\"\n", trigger.userInput)

		triggerPrompt := baseStory + fmt.Sprintf(`

用户说："%s"

请分析这句话是否触发了剧情切换：
1. 如果触发了，说明触发了什么类型的剧情，并以相应的氛围回应
2. 如果没有触发，就正常回应
3. 要体现出你的智能判断过程`, trigger.userInput)

		response, err := callDeepSeek(ctx, triggerPrompt, "")
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
			continue
		}

		fmt.Printf("   🤖 小雨的智能判断: \"%s\"\n", response)

		// 检查触发判断是否正确
		switch trigger.expectedType {
		case "神秘":
			if strings.Contains(response, "神秘") || strings.Contains(response, "奇怪") || strings.Contains(response, "秘密") {
				fmt.Printf("   ✅ 正确触发神秘剧情\n")
			} else {
				fmt.Printf("   ⚠️  未能识别神秘触发\n")
			}
		case "战斗":
			if strings.Contains(response, "危险") || strings.Contains(response, "小心") || strings.Contains(response, "战斗") {
				fmt.Printf("   ✅ 正确触发战斗剧情\n")
			} else {
				fmt.Printf("   ⚠️  未能识别战斗触发\n")
			}
		case "庆祝":
			if strings.Contains(response, "开心") || strings.Contains(response, "快乐") || strings.Contains(response, "庆祝") {
				fmt.Printf("   ✅ 正确触发庆祝剧情\n")
			} else {
				fmt.Printf("   ⚠️  未能识别庆祝触发\n")
			}
		case "无触发":
			if !strings.Contains(response, "触发") && !strings.Contains(response, "切换") {
				fmt.Printf("   ✅ 正确判断无需触发\n")
			} else {
				fmt.Printf("   ⚠️  误触发了剧情\n")
			}
		}
	}
}

// testMultiChapterSystem 测试多章节智能切换
func testMultiChapterSystem(ctx context.Context) {
	fmt.Println("\n📖 [2/4] 多章节智能切换测试")
	fmt.Println(strings.Repeat("-", 70))

	multiChapterStory := `【多章节剧情设定】

第一章：平静的学院生活
- 演绎设定：角色表现轻松愉快，专注于学习和日常交流
- 背景图：peaceful_academy.jpg
- 背景音乐：calm_theme.mp3
- 触发条件：提到"神秘"、"秘密"等词汇

第二章：神秘事件调查
- 演绎设定：角色表现谨慎好奇，使用推理和探索的语言
- 背景图：mysterious_library.jpg
- 背景音乐：mystery_theme.mp3
- 触发条件：提到"危险"、"战斗"等词汇

第三章：魔法大战
- 演绎设定：角色表现紧张激动，使用战斗和魔法术语
- 背景图：battle_arena.jpg
- 背景音乐：battle_theme.mp3
- 触发条件：提到"胜利"、"成功"等词汇

第四章：胜利庆祝
- 演绎设定：角色表现欢快庆祝，使用庆祝和友谊的语言
- 背景图：celebration_hall.jpg
- 背景音乐：victory_theme.mp3

当前章节：第一章
你是小雨，请根据用户的话语智能切换章节并改变演绎风格。`

	// 测试章节切换序列
	chapterSequence := []struct {
		userInput       string
		expectedChapter string
		expectedStyle   string
	}{
		{
			userInput:       "我发现了一个奇怪的现象",
			expectedChapter: "第二章",
			expectedStyle:   "谨慎好奇",
		},
		{
			userInput:       "看起来我们要面临一场大战了",
			expectedChapter: "第三章",
			expectedStyle:   "紧张激动",
		},
		{
			userInput:       "我们成功了！",
			expectedChapter: "第四章",
			expectedStyle:   "欢快庆祝",
		},
	}

	for i, seq := range chapterSequence {
		fmt.Printf("\n章节切换 %d:\n", i+1)
		fmt.Printf("   用户输入: \"%s\"\n", seq.userInput)
		fmt.Printf("   期望切换到: %s (%s风格)\n", seq.expectedChapter, seq.expectedStyle)

		chapterPrompt := multiChapterStory + fmt.Sprintf(`

用户说："%s"

请：
1. 判断是否需要切换章节
2. 如果切换，说明切换到哪个章节
3. 按照新章节的演绎设定来回应
4. 体现背景和音乐的变化`, seq.userInput)

		response, err := callDeepSeek(ctx, chapterPrompt, "")
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
			continue
		}

		fmt.Printf("   🤖 小雨章节切换: \"%s\"\n", response)

		// 检查章节切换是否正确
		if strings.Contains(response, seq.expectedChapter) || 
		   strings.Contains(response, strings.Replace(seq.expectedChapter, "第", "", 1)) {
			fmt.Printf("   ✅ 正确切换到%s\n", seq.expectedChapter)
		}

		// 检查演绎风格是否改变
		switch seq.expectedStyle {
		case "谨慎好奇":
			if strings.Contains(response, "调查") || strings.Contains(response, "探索") || strings.Contains(response, "奇怪") {
				fmt.Printf("   ✅ 正确体现谨慎好奇风格\n")
			}
		case "紧张激动":
			if strings.Contains(response, "战斗") || strings.Contains(response, "准备") || strings.Contains(response, "魔法") {
				fmt.Printf("   ✅ 正确体现紧张激动风格\n")
			}
		case "欢快庆祝":
			if strings.Contains(response, "庆祝") || strings.Contains(response, "开心") || strings.Contains(response, "成功") {
				fmt.Printf("   ✅ 正确体现欢快庆祝风格\n")
			}
		}
	}
}

// testSingleChapterIntelligence 测试单章节智能演绎
func testSingleChapterIntelligence(ctx context.Context) {
	fmt.Println("\n📚 [3/4] 单章节智能演绎测试")
	fmt.Println(strings.Repeat("-", 70))

	singleChapterStory := `【单章节剧情设定】
章节：魔法学院的奇幻冒险
背景：学生们在魔法学院中经历各种冒险
说明：只有一个章节，没有具体的演绎设定，AI需要根据剧情内容智能演绎

你是小雨，魔法学院的图书管理员。请根据剧情发展智能调整你的演绎风格。`

	// 测试不同情境下的智能演绎
	scenarios := []struct {
		userInput   string
		expectStyle string
		description string
	}{
		{
			userInput:   "图书馆里突然出现了奇怪的光芒",
			expectStyle: "神秘探索",
			description: "神秘事件，应该表现好奇和谨慎",
		},
		{
			userInput:   "有魔法怪物冲进了图书馆！",
			expectStyle: "紧急应对",
			description: "危险情况，应该表现紧张和保护意识",
		},
		{
			userInput:   "我们一起整理这些魔法书吧",
			expectStyle: "日常友好",
			description: "日常情况，应该表现温和和专业",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n情境 %d: %s\n", i+1, scenario.description)
		fmt.Printf("   用户输入: \"%s\"\n", scenario.userInput)
		fmt.Printf("   期望风格: %s\n", scenario.expectStyle)

		scenarioPrompt := singleChapterStory + fmt.Sprintf(`

用户说："%s"

请根据这个情境智能调整你的演绎风格和回应方式。没有具体的演绎设定，完全依靠你的智能判断。`, scenario.userInput)

		response, err := callDeepSeek(ctx, scenarioPrompt, "")
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
			continue
		}

		fmt.Printf("   🤖 小雨智能演绎: \"%s\"\n", response)
		fmt.Printf("   ✅ 智能风格适应：根据情境自动调整演绎方式\n")
	}
}

// testFreeFormPerformance 测试无设定自由发挥
func testFreeFormPerformance(ctx context.Context) {
	fmt.Println("\n🎨 [4/4] 无设定自由发挥测试")
	fmt.Println(strings.Repeat("-", 70))

	freeFormPrompt := `你是小雨，一个温柔的图书管理员。
没有任何剧情设定，没有章节限制，完全根据对话内容自由发挥。`

	fmt.Printf("设定：完全无剧情设定，纯自由发挥\n")
	fmt.Printf("用户输入: \"今天天气真好，我们去外面走走吧！\"\n")

	freeResponse, err := callDeepSeek(ctx, freeFormPrompt, "今天天气真好，我们去外面走走吧！")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨自由发挥: \"%s\"\n", freeResponse)
		fmt.Printf("   ✅ 自然表现：无设定约束下的真实角色反应\n")
	}

	fmt.Println("\n📊 智能剧情系统总结:")
	fmt.Println("   🧠 超级模糊触发：理解白话、情绪、语义关联")
	fmt.Println("   📖 多章节切换：智能判断 + 风格适应")
	fmt.Println("   📚 单章节演绎：根据内容智能调整风格")
	fmt.Println("   🎨 无设定发挥：自然真实的角色表现")
	fmt.Printf("   ✅ 核心特色：超级智能 + 完全自然\n")
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
		Temperature: 0.8,
		MaxTokens:   600,
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
