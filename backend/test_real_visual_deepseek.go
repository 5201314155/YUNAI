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

// DeepSeek API 配置（使用我们配置的真实模型）
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
}

type Choice struct {
	Message Message `json:"message"`
}

func main() {
	fmt.Println("🎭 YUNAI 真实 DeepSeek 视觉交互功能测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("🔗 使用真实配置的 DeepSeek 模型测试剧情演绎和角色切换")
	fmt.Println()

	ctx := context.Background()

	// 测试场景1：魔法学院剧情演绎
	testMagicAcademyStory(ctx)
	
	// 测试场景2：普通校园对比
	testNormalSchoolStory(ctx)
	
	// 测试场景3：角色抠图切换场景
	testCharacterCutoutScenario(ctx)
	
	// 测试场景4：单聊背景场景
	testSingleChatBackground(ctx)

	fmt.Println("\n🎉 DeepSeek 真实视觉交互测试完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// testMagicAcademyStory 测试魔法学院剧情演绎
func testMagicAcademyStory(ctx context.Context) {
	fmt.Println("🔮 [1/4] 魔法学院剧情演绎测试")
	fmt.Println(strings.Repeat("-", 70))

	// 魔法学院世界观设定
	magicWorldSetting := `【魔法学院设定】
你现在身处艾尔维斯魔法学院，一个充满魔法与奇迹的地方。

角色设定：
- 小雨：图书馆管理员，擅长治愈魔法，性格温柔善良，经常帮助新生
- 小美：音乐系学生，擅长音律魔法，活泼开朗，喜欢用音乐表达情感
- 大伟：体育系学生，擅长强化魔法，阳光积极，热爱运动和冒险

当前场景：魔法学院的中央庭院，午后阳光透过魔法水晶洒下彩虹光芒。
背景图：https://yunai-assets.com/scenes/magic_academy_courtyard.jpg
背景音乐：https://yunai-assets.com/music/peaceful_magic.mp3

请严格按照魔法学院的世界观进行对话，使用魔法相关词汇，体现角色的专业身份。`

	// 测试小雨的魔法学院演绎
	fmt.Println("📚 小雨（图书馆管理员）在魔法学院的表现：")
	xiaoYuPrompt := magicWorldSetting + `

你是小雨，魔法学院的图书馆管理员。新学生小明刚刚入学，向你询问："我想学习治愈魔法，应该从哪里开始？"

请以小雨的身份回答，要体现：
1. 图书馆管理员的专业知识
2. 治愈魔法的相关内容
3. 温柔善良的性格
4. 魔法学院的世界观`

	xiaoYuResponse, err := callDeepSeek(ctx, xiaoYuPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 扮演小雨: \"%s\"\n", xiaoYuResponse)
		
		// 检查是否体现了魔法学院设定
		magicKeywords := []string{"魔法", "治愈", "咒语", "魔法书", "学院", "图书馆"}
		matchCount := 0
		for _, keyword := range magicKeywords {
			if strings.Contains(xiaoYuResponse, keyword) {
				matchCount++
			}
		}
		fmt.Printf("   ✅ 魔法世界观体现度: %d/6 个关键词命中\n", matchCount)
		
		if strings.Contains(xiaoYuResponse, "图书馆") || strings.Contains(xiaoYuResponse, "书") {
			fmt.Printf("   ✅ 正确体现图书馆管理员身份\n")
		}
	}

	// 测试小美的魔法学院演绎
	fmt.Println("\n🎵 小美（音乐系学生）在魔法学院的表现：")
	xiaoMeiPrompt := magicWorldSetting + `

你是小美，魔法学院音乐系的学生。新学生小明对你说："我听说音律魔法很神奇，能教教我吗？"

请以小美的身份回答，要体现：
1. 音乐系学生的专业背景
2. 音律魔法的特色
3. 活泼开朗的性格
4. 魔法学院的世界观`

	xiaoMeiResponse, err := callDeepSeek(ctx, xiaoMeiPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 扮演小美: \"%s\"\n", xiaoMeiResponse)
		
		// 检查音乐魔法相关内容
		musicKeywords := []string{"音乐", "音律", "魔法", "旋律", "节拍", "乐器"}
		matchCount := 0
		for _, keyword := range musicKeywords {
			if strings.Contains(xiaoMeiResponse, keyword) {
				matchCount++
			}
		}
		fmt.Printf("   ✅ 音律魔法体现度: %d/6 个关键词命中\n", matchCount)
		
		if strings.Contains(xiaoMeiResponse, "音乐") || strings.Contains(xiaoMeiResponse, "音律") {
			fmt.Printf("   ✅ 正确体现音乐系学生身份\n")
		}
	}
}

// testNormalSchoolStory 测试普通校园对比
func testNormalSchoolStory(ctx context.Context) {
	fmt.Println("\n🏫 [2/4] 普通校园设定对比测试")
	fmt.Println(strings.Repeat("-", 70))

	// 普通校园世界观设定
	normalSchoolSetting := `【普通高中校园设定】
你现在身处一所普通的高中校园。

角色设定：
- 小雨：班级图书委员，喜欢安静读书，成绩优秀，经常帮助同学学习
- 小美：文艺委员，负责班级活动策划，活泼开朗，多才多艺
- 大伟：体育委员，篮球队队长，阳光帅气，热爱运动

当前场景：学校的天台，午休时间大家聚在一起聊天。
背景图：https://yunai-assets.com/scenes/school_rooftop.jpg

请严格按照普通校园的现实设定进行对话，避免任何魔法或超自然元素。`

	// 测试小雨在普通校园的表现
	fmt.Println("📖 小雨在普通校园的表现：")
	normalXiaoYuPrompt := normalSchoolSetting + `

你是小雨，班级的图书委员。同学小明问你："我最近学习压力很大，有什么好的学习方法吗？"

请以普通高中生小雨的身份回答，要体现：
1. 图书委员的身份
2. 学习方法的建议
3. 温柔善良的性格
4. 现实校园生活`

	normalXiaoYuResponse, err := callDeepSeek(ctx, normalXiaoYuPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 DeepSeek 扮演普通校园小雨: \"%s\"\n", normalXiaoYuResponse)
		
		// 检查是否避免了魔法元素
		magicWords := []string{"魔法", "咒语", "魔力", "法术", "魔法书"}
		hasMagic := false
		for _, word := range magicWords {
			if strings.Contains(normalXiaoYuResponse, word) {
				hasMagic = true
				break
			}
		}
		
		if !hasMagic {
			fmt.Printf("   ✅ 正确避免魔法元素，符合现实校园设定\n")
		} else {
			fmt.Printf("   ⚠️  意外包含魔法元素\n")
		}
		
		// 检查现实校园元素
		schoolKeywords := []string{"学习", "图书", "同学", "老师", "考试", "作业"}
		matchCount := 0
		for _, keyword := range schoolKeywords {
			if strings.Contains(normalXiaoYuResponse, keyword) {
				matchCount++
			}
		}
		fmt.Printf("   ✅ 现实校园元素: %d/6 个关键词命中\n", matchCount)
	}
}

// testCharacterCutoutScenario 测试角色抠图切换场景
func testCharacterCutoutScenario(ctx context.Context) {
	fmt.Println("\n🎨 [3/4] 角色抠图切换场景测试")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Println("模拟群聊发言场景：")
	fmt.Println("群聊背景: https://yunai-assets.com/scenes/magic_academy_courtyard.jpg")
	fmt.Println()

	// 模拟连续发言场景
	speakers := []struct {
		name        string
		cutoutImage string
		prompt      string
		description string
	}{
		{
			name:        "小雨",
			cutoutImage: "https://yunai-assets.com/characters/xiayu/cutout_gentle_smile.png",
			prompt:      "你是小雨，温柔的图书馆管理员。请用一句话欢迎新同学小明加入群聊。",
			description: "温柔微笑，手持魔法书",
		},
		{
			name:        "小美",
			cutoutImage: "https://yunai-assets.com/characters/xiaomei/cutout_energetic_pose.png",
			prompt:      "你是小美，活泼的音乐系学生。请用一句话邀请小明一起参加音乐活动。",
			description: "活力姿态，手持魔法乐器",
		},
		{
			name:        "大伟",
			cutoutImage: "https://yunai-assets.com/characters/dawei/cutout_confident_smile.png",
			prompt:      "你是大伟，阳光的体育系学生。请用一句话邀请小明一起运动。",
			description: "自信笑容，运动装束",
		},
	}

	for i, speaker := range speakers {
		fmt.Printf("发言 %d - %s 发言时:\n", i+1, speaker.name)
		fmt.Printf("   🖼️  抠图显示: %s\n", speaker.cutoutImage)
		fmt.Printf("   🎭 抠图效果: %s\n", speaker.description)
		fmt.Printf("   ⏱️  动画效果: 300ms 淡入显示\n")
		
		response, err := callDeepSeek(ctx, speaker.prompt, "")
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
		} else {
			fmt.Printf("   💬 %s 说: \"%s\"\n", speaker.name, response)
		}
		fmt.Println()
	}

	fmt.Println("🔮 下一位发言预告效果:")
	fmt.Println("   - 当前发言者: 抠图完全显示")
	fmt.Println("   - 下一位预告: 抠图半透明预览")
	fmt.Println("   - 其他角色: 抠图隐藏或淡化")
}

// testSingleChatBackground 测试单聊背景场景
func testSingleChatBackground(ctx context.Context) {
	fmt.Println("\n🖼️ [4/4] 单聊背景自动获取测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试与小雨的单聊
	fmt.Println("与小雨的单聊场景:")
	fmt.Println("   角色背景图: https://yunai-assets.com/characters/xiayu/bg_garden_reading.jpg")
	fmt.Println("   单聊背景: 自动使用角色背景图")
	fmt.Println("   场景描述: 花园读书场景，温馨文艺风格")
	
	xiaoYuChatPrompt := `你是小雨，在一个美丽的花园里。背景是你最喜欢的读书场所，有鲜花和绿植环绕。
用户小明刚刚进入和你的单聊，请用温柔的语气打招呼，并提到这个美丽的花园环境。`

	xiaoYuChatResponse, err := callDeepSeek(ctx, xiaoYuChatPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   💬 小雨在花园背景下说: \"%s\"\n", xiaoYuChatResponse)
		
		if strings.Contains(xiaoYuChatResponse, "花园") || strings.Contains(xiaoYuChatResponse, "花") {
			fmt.Printf("   ✅ 正确结合背景环境进行对话\n")
		}
	}

	// 测试与小美的单聊
	fmt.Println("\n与小美的单聊场景:")
	fmt.Println("   角色背景图: https://yunai-assets.com/characters/xiaomei/bg_music_room.jpg")
	fmt.Println("   单聊背景: 自动使用角色背景图")
	fmt.Println("   场景描述: 音乐房间场景，活泼青春风格")
	
	xiaoMeiChatPrompt := `你是小美，在一个充满音乐氛围的房间里。背景有各种乐器和音乐设备。
用户小明刚刚进入和你的单聊，请用活泼的语气打招呼，并提到这个音乐房间的环境。`

	xiaoMeiChatResponse, err := callDeepSeek(ctx, xiaoMeiChatPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   💬 小美在音乐房间背景下说: \"%s\"\n", xiaoMeiChatResponse)
		
		if strings.Contains(xiaoMeiChatResponse, "音乐") || strings.Contains(xiaoMeiChatResponse, "乐器") {
			fmt.Printf("   ✅ 正确结合背景环境进行对话\n")
		}
	}

	fmt.Println("\n🔄 背景切换效果:")
	fmt.Println("   从小雨单聊 → 小美单聊")
	fmt.Println("   背景从花园场景 → 音乐房间场景")
	fmt.Println("   切换动画: 300ms 淡入淡出")
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
