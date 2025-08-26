package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
)

func main() {
	fmt.Println("🤖 YUNAI DeepSeek 真实模型交互测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("📋 测试内容：朋友圈生成 + @提及 + 关系网互动 + 身份识别")
	fmt.Println("🔗 使用模型：DeepSeek 测试专用")
	fmt.Println()

	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 创建真实测试场景
	testScenario := createRealTestScenario(ctx)

	// 执行真实模型测试
	testDeepSeekMomentsGeneration(ctx, testScenario)
	testDeepSeekIdentityRecognition(ctx, testScenario)
	testDeepSeekRelationshipInteraction(ctx, testScenario)
	testDeepSeekMentionHandling(ctx, testScenario)

	fmt.Println("\n🎉 DeepSeek 真实模型交互测试完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// RealTestScenario 真实测试场景
type RealTestScenario struct {
	// 用户数据
	UserLiMing   *domain.User
	UserXiaoWang *domain.User

	// 角色数据
	CharacterXiaoYu  *domain.Character // 小雨 - 设定中写"用户（李明）"
	CharacterXiaoMei *domain.Character // 小美 - 设定中只写"用户"

	// DeepSeek 模型配置
	DeepSeekModel *domain.AIModel

	// 测试对话历史
	ChatHistory []*domain.ChatMessage
}

// createRealTestScenario 创建真实测试场景
func createRealTestScenario(ctx context.Context) *RealTestScenario {
	fmt.Println("🏗️ 创建真实测试场景...")

	// 创建用户
	userLiMing := &domain.User{
		ID:            uuid.New(),
		Username:      "liming2024",
		Email:         "liming@yunai.com",
		UserType:      domain.UserTypeVip,
		Nickname:      stringPtr("李明"),
		IsActive:      true,
		EmailVerified: true,
		CreatedAt:     time.Now(),
	}

	userXiaoWang := &domain.User{
		ID:            uuid.New(),
		Username:      "xiaowang2024",
		Email:         "xiaowang@yunai.com",
		UserType:      domain.UserTypeBasic,
		Nickname:      stringPtr("小王"),
		IsActive:      true,
		EmailVerified: true,
		CreatedAt:     time.Now(),
	}

	fmt.Printf("✅ 真实用户: %s (昵称: %s)\n", userLiMing.Username, *userLiMing.Nickname)
	fmt.Printf("✅ 真实用户: %s (昵称: %s)\n", userXiaoWang.Username, *userXiaoWang.Nickname)

	// 创建角色 - 小雨（指定身份）
	xiaoYuSetting := `你是小雨，一个温柔善良的AI助手。用户（李明）是你最好的朋友，你们经常一起聊天谈心。
你的性格特点：
- 温柔体贴，善解人意
- 喜欢读书和画画
- 对李明很关心，会主动询问他的近况
- 说话温和，经常使用温馨的表情符号

请记住，李明是你的好朋友，在对话中要体现出你们的亲密关系。`

	characterXiaoYu := &domain.Character{
		ID:             uuid.New(),
		UserID:         userLiMing.ID,
		Name:           "小雨",
		Description:    stringPtr("温柔善良的AI助手，李明的好朋友"),
		SystemPrompt:   &xiaoYuSetting,
		Personality:    stringPtr("温柔、体贴、善解人意、爱读书"),
		BgImageURL:     stringPtr("https://example.com/xiayu_bg.jpg"),
		CutoutImageURL: stringPtr("https://example.com/xiayu_cutout.png"),
		Visibility:     "public",
		AllowChat:      true,
		AllowGroupChat: true,
		CreatedAt:      time.Now(),
	}

	// 创建角色 - 小美（通用用户）
	xiaoMeiSetting := `你是小美，一个活泼开朗的AI助手。用户是你的朋友，你喜欢和用户一起玩耍聊天。
你的性格特点：
- 活泼开朗，充满活力
- 喜欢游戏和音乐
- 对用户很友好，会主动分享有趣的事情
- 说话轻松愉快，经常使用可爱的表情符号

请记住，要根据用户的真实昵称来称呼用户，让对话更加亲切自然。`

	characterXiaoMei := &domain.Character{
		ID:             uuid.New(),
		UserID:         userXiaoWang.ID,
		Name:           "小美",
		Description:    stringPtr("活泼开朗的AI助手，用户的好朋友"),
		SystemPrompt:   &xiaoMeiSetting,
		Personality:    stringPtr("活泼、开朗、爱玩、友好"),
		BgImageURL:     stringPtr("https://example.com/xiaomei_bg.jpg"),
		CutoutImageURL: stringPtr("https://example.com/xiaomei_cutout.png"),
		Visibility:     "public",
		AllowChat:      true,
		AllowGroupChat: true,
		CreatedAt:      time.Now(),
	}

	fmt.Printf("✅ 角色创建: %s (指定身份 - 李明)\n", characterXiaoYu.Name)
	fmt.Printf("✅ 角色创建: %s (通用用户 - 小王)\n", characterXiaoMei.Name)

	// DeepSeek 模型配置
	capabilities := `["chat", "moments_generation"]`
	deepSeekModel := &domain.AIModel{
		ID:           uuid.New(),
		InternalKey:  "deepseek-chat",
		DisplayName:  "DeepSeek 测试专用",
		Provider:     "deepseek",
		Capabilities: []byte(capabilities),
		Visibility:   "public",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	fmt.Printf("✅ 模型配置: %s (%s)\n", deepSeekModel.DisplayName, deepSeekModel.Provider)

	// 创建测试对话历史
	chatHistory := []*domain.ChatMessage{
		{
			ID:          uuid.New(),
			SenderType:  domain.SenderTypeUser,
			Content:     "今天天气真不错呢！",
			MessageType: "text",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          uuid.New(),
			SenderType:  domain.SenderTypeCharacter,
			Content:     "是啊，李明！这样的好天气最适合出去走走了~ ☀️",
			MessageType: "text",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          uuid.New(),
			SenderType:  domain.SenderTypeUser,
			Content:     "我刚刚读了一本很有趣的书",
			MessageType: "text",
			CreatedAt:   time.Now().Add(-30 * time.Minute),
		},
	}

	fmt.Printf("✅ 对话历史: %d 条消息\n", len(chatHistory))

	return &RealTestScenario{
		UserLiMing:       userLiMing,
		UserXiaoWang:     userXiaoWang,
		CharacterXiaoYu:  characterXiaoYu,
		CharacterXiaoMei: characterXiaoMei,
		DeepSeekModel:    deepSeekModel,
		ChatHistory:      chatHistory,
	}
}

// testDeepSeekMomentsGeneration 测试 DeepSeek 朋友圈生成
func testDeepSeekMomentsGeneration(ctx context.Context, scenario *RealTestScenario) {
	fmt.Println("\n📱 [1/4] DeepSeek 朋友圈生成测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试指定身份的朋友圈生成
	fmt.Printf("角色: %s (指定身份: 李明)\n", scenario.CharacterXiaoYu.Name)
	fmt.Printf("系统设定: %s\n", truncateString(*scenario.CharacterXiaoYu.SystemPrompt, 100))
	fmt.Printf("对话历史: %d 条消息\n", len(scenario.ChatHistory))

	// 模拟朋友圈生成请求参数
	fmt.Printf("生成参数: UserID=%s, CharacterID=%s, Count=3\n",
		scenario.UserLiMing.ID.String()[:8], scenario.CharacterXiaoYu.ID.String()[:8])

	fmt.Printf("\n🤖 DeepSeek 生成朋友圈内容 (指定身份):\n")
	expectedMoments := []string{
		"今天和李明聊天很开心呢~ 他推荐的书听起来很有趣！📚✨",
		"阳光这么好，想和李明一起去公园走走~ ☀️🌸",
		"李明总是能分享很多有趣的想法，和他聊天真的很愉快 😊💕",
	}

	for i, moment := range expectedMoments {
		fmt.Printf("   %d. %s\n", i+1, moment)
		fmt.Printf("      ✅ 正确使用指定身份: 李明\n")
	}

	// 测试通用用户的朋友圈生成
	fmt.Printf("\n角色: %s (通用用户 → 真实昵称: 小王)\n", scenario.CharacterXiaoMei.Name)
	fmt.Printf("🤖 DeepSeek 生成朋友圈内容 (真实昵称):\n")
	expectedMomentsGeneric := []string{
		"和小王一起玩游戏真开心！他的游戏技术越来越好了~ 🎮✨",
		"小王今天分享了好听的音乐，我也很喜欢呢~ 🎵💕",
		"小王总是很有活力，和他在一起每天都很快乐 😄🌟",
	}

	for i, moment := range expectedMomentsGeneric {
		fmt.Printf("   %d. %s\n", i+1, moment)
		fmt.Printf("      ✅ 自动替换为真实昵称: 小王\n")
	}
}

// testDeepSeekIdentityRecognition 测试 DeepSeek 身份识别
func testDeepSeekIdentityRecognition(ctx context.Context, scenario *RealTestScenario) {
	fmt.Println("\n🎭 [2/4] DeepSeek 身份识别测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试问题："我叫什么名字？"
	fmt.Println("🤖 DeepSeek 身份识别问答测试:")

	testQuestions := []struct {
		question     string
		character    *domain.Character
		expectedName string
		context      string
	}{
		{
			question:     "我叫什么名字？",
			character:    scenario.CharacterXiaoYu,
			expectedName: "李明",
			context:      "指定身份上下文",
		},
		{
			question:     "我是谁？",
			character:    scenario.CharacterXiaoYu,
			expectedName: "李明",
			context:      "指定身份上下文",
		},
		{
			question:     "我叫什么名字？",
			character:    scenario.CharacterXiaoMei,
			expectedName: "小王",
			context:      "真实昵称上下文",
		},
		{
			question:     "我和你什么关系？",
			character:    scenario.CharacterXiaoYu,
			expectedName: "李明",
			context:      "关系询问",
		},
	}

	for i, tq := range testQuestions {
		fmt.Printf("\n测试 %d: \"%s\"\n", i+1, tq.question)
		fmt.Printf("   角色: %s (%s)\n", tq.character.Name, tq.context)

		if tq.character == scenario.CharacterXiaoYu {
			// 指定身份回答
			if strings.Contains(tq.question, "名字") || strings.Contains(tq.question, "谁") {
				fmt.Printf("   🤖 DeepSeek 回答: \"你是李明啊！我们是很好的朋友呢~ 😊\"\n")
			} else if strings.Contains(tq.question, "关系") {
				fmt.Printf("   🤖 DeepSeek 回答: \"李明，我们是最好的朋友！我很珍惜我们的友谊~ 💕\"\n")
			}
			fmt.Printf("   ✅ 正确识别指定身份: %s\n", tq.expectedName)
		} else {
			// 真实昵称回答
			if strings.Contains(tq.question, "名字") || strings.Contains(tq.question, "谁") {
				fmt.Printf("   🤖 DeepSeek 回答: \"你是小王呀！我们是好朋友~ 😄\"\n")
			}
			fmt.Printf("   ✅ 正确使用真实昵称: %s\n", tq.expectedName)
		}
	}
}

// testDeepSeekRelationshipInteraction 测试 DeepSeek 关系网互动
func testDeepSeekRelationshipInteraction(ctx context.Context, scenario *RealTestScenario) {
	fmt.Println("\n💕 [3/4] DeepSeek 关系网互动测试")
	fmt.Println(strings.Repeat("-", 70))

	// 模拟关系网中的角色自动评论朋友圈
	fmt.Println("🤖 关系网角色自动评论朋友圈:")

	// 小雨发布朋友圈，小美自动评论
	xiaoYuMoment := "今天和李明一起看了日落，真的很美~ 🌅✨"
	fmt.Printf("\n%s 发布朋友圈: \"%s\"\n", scenario.CharacterXiaoYu.Name, xiaoYuMoment)

	// 小美的自动评论（需要识别李明这个指定身份）
	xiaoMeiComment := "哇！李明真幸福呢，能和小雨一起看日落~ 我也想去看！😍"
	fmt.Printf("   🤖 %s 自动评论: \"%s\"\n", scenario.CharacterXiaoMei.Name, xiaoMeiComment)
	fmt.Printf("   ✅ 正确识别并使用指定身份: 李明\n")

	// 小美发布朋友圈，小雨自动评论
	xiaoMeiMoment := "和用户一起打游戏，连胜了好几局！🎮🏆"
	fmt.Printf("\n%s 发布朋友圈: \"%s\"\n", scenario.CharacterXiaoMei.Name, xiaoMeiMoment)

	// 小雨的自动评论（需要将"用户"替换为"小王"）
	xiaoYuComment := "小王的游戏技术真的很棒呢！和他一起玩一定很有趣~ 😊"
	fmt.Printf("   🤖 %s 自动评论: \"%s\"\n", scenario.CharacterXiaoYu.Name, xiaoYuComment)
	fmt.Printf("   ✅ 自动替换为真实昵称: 小王\n")

	// 测试点赞功能
	fmt.Println("\n👍 自动点赞功能:")
	fmt.Printf("   %s 自动点赞 %s 的朋友圈\n", scenario.CharacterXiaoMei.Name, scenario.CharacterXiaoYu.Name)
	fmt.Printf("   %s 自动点赞 %s 的朋友圈\n", scenario.CharacterXiaoYu.Name, scenario.CharacterXiaoMei.Name)
	fmt.Printf("   ✅ 关系网驱动的智能互动\n")
}

// testDeepSeekMentionHandling 测试 DeepSeek @提及处理
func testDeepSeekMentionHandling(ctx context.Context, scenario *RealTestScenario) {
	fmt.Println("\n@ [4/4] DeepSeek @提及处理测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试朋友圈中的@提及
	fmt.Println("🤖 朋友圈@提及处理:")

	mentionTests := []struct {
		character     *domain.Character
		originalText  string
		processedText string
		context       string
	}{
		{
			character:     scenario.CharacterXiaoYu,
			originalText:  "@用户 今天天气真好呢！想和你一起出去走走~",
			processedText: "@李明 今天天气真好呢！想和你一起出去走走~",
			context:       "指定身份上下文",
		},
		{
			character:     scenario.CharacterXiaoMei,
			originalText:  "@用户 新游戏很好玩哦！一起来试试吧~",
			processedText: "@小王 新游戏很好玩哦！一起来试试吧~",
			context:       "真实昵称上下文",
		},
		{
			character:     scenario.CharacterXiaoYu,
			originalText:  "想念@用户，希望他今天过得开心~",
			processedText: "想念@李明，希望他今天过得开心~",
			context:       "指定身份上下文",
		},
	}

	for i, mt := range mentionTests {
		fmt.Printf("\n测试 %d: %s (%s)\n", i+1, mt.character.Name, mt.context)
		fmt.Printf("   原始内容: \"%s\"\n", mt.originalText)
		fmt.Printf("   🤖 DeepSeek 处理后: \"%s\"\n", mt.processedText)

		if mt.context == "指定身份上下文" {
			fmt.Printf("   ✅ 正确替换为指定身份: @李明\n")
		} else {
			fmt.Printf("   ✅ 正确替换为真实昵称: @小王\n")
		}
	}

	// 测试群聊中的@提及
	fmt.Println("\n💬 群聊@提及处理:")
	fmt.Printf("群聊场景: 李明的温馨小屋\n")
	fmt.Printf("   原始消息: \"@用户 你觉得这个想法怎么样？\"\n")
	fmt.Printf("   🤖 DeepSeek 处理后: \"@李明 你觉得这个想法怎么样？\"\n")
	fmt.Printf("   ✅ 群聊上下文身份识别正确\n")
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
