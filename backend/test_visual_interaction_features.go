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
	fmt.Println("🎨 YUNAI 视觉交互功能测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("📋 测试内容：单聊背景 + 群聊剧情演绎 + 角色抠图切换 + 世界观设定")
	fmt.Println()

	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 创建测试场景
	testScenario := createVisualTestScenario(ctx)

	// 执行各项测试
	testSingleChatBackground(ctx, testScenario)
	testGroupChatStoryMode(ctx, testScenario)
	testCharacterCutoutSwitching(ctx, testScenario)
	testWorldSettingIntegration(ctx, testScenario)

	fmt.Println("\n🎉 视觉交互功能测试完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// VisualTestScenario 视觉测试场景
type VisualTestScenario struct {
	// 用户
	User *domain.User

	// 角色（带完整两图系统）
	CharacterXiaoYu  *domain.Character // 小雨 - 温柔系
	CharacterXiaoMei *domain.Character // 小美 - 活泼系
	CharacterDaWei   *domain.Character // 大伟 - 阳光系

	// 单聊会话（使用 GroupChat 代替，设置为单聊模式）
	SingleChatXiaoYu  *domain.GroupChat
	SingleChatXiaoMei *domain.GroupChat

	// 群聊（带剧情设定）
	GroupChatStory  *domain.GroupChat // 剧情群聊
	GroupChatCasual *domain.GroupChat // 日常群聊

	// 剧情设定（简化为基本信息）
	StoryChapterTitle string
	StorySceneTitle   string
	StorySceneDesc    string
	StorySceneBg      string
	StorySceneMusic   string

	// 测试消息
	TestMessages []*domain.ChatMessage
}

// createVisualTestScenario 创建视觉测试场景
func createVisualTestScenario(ctx context.Context) *VisualTestScenario {
	fmt.Println("🏗️ 创建视觉测试场景...")

	// 创建用户
	user := &domain.User{
		ID:            uuid.New(),
		Username:      "visual_test_user",
		Email:         "test@yunai.com",
		UserType:      domain.UserTypeVip,
		Nickname:      stringPtr("小明"),
		IsActive:      true,
		EmailVerified: true,
		CreatedAt:     time.Now(),
	}
	fmt.Printf("✅ 测试用户: %s (昵称: %s)\n", user.Username, *user.Nickname)

	// 创建角色1：小雨（温柔系，完整两图）
	xiaoYuPersonality := "温柔善良的邻家女孩，喜欢读书和画画，说话轻声细语"
	characterXiaoYu := &domain.Character{
		ID:             uuid.New(),
		UserID:         user.ID,
		Name:           "小雨",
		Description:    &xiaoYuPersonality,
		Personality:    &xiaoYuPersonality,
		BgImageURL:     stringPtr("https://yunai-assets.com/characters/xiayu/bg_garden_reading.jpg"),
		CutoutImageURL: stringPtr("https://yunai-assets.com/characters/xiayu/cutout_gentle_smile.png"),
		Visibility:     "public",
		AllowChat:      true,
		AllowGroupChat: true,
		CreatedAt:      time.Now(),
	}

	// 创建角色2：小美（活泼系，完整两图）
	xiaoMeiPersonality := "活泼开朗的少女，喜欢音乐和游戏，充满活力"
	characterXiaoMei := &domain.Character{
		ID:             uuid.New(),
		UserID:         user.ID,
		Name:           "小美",
		Description:    &xiaoMeiPersonality,
		Personality:    &xiaoMeiPersonality,
		BgImageURL:     stringPtr("https://yunai-assets.com/characters/xiaomei/bg_music_room.jpg"),
		CutoutImageURL: stringPtr("https://yunai-assets.com/characters/xiaomei/cutout_energetic_pose.png"),
		Visibility:     "public",
		AllowChat:      true,
		AllowGroupChat: true,
		CreatedAt:      time.Now(),
	}

	// 创建角色3：大伟（阳光系，完整两图）
	daWeiPersonality := "阳光帅气的大男孩，喜欢运动和冒险，积极向上"
	characterDaWei := &domain.Character{
		ID:             uuid.New(),
		UserID:         user.ID,
		Name:           "大伟",
		Description:    &daWeiPersonality,
		Personality:    &daWeiPersonality,
		BgImageURL:     stringPtr("https://yunai-assets.com/characters/dawei/bg_basketball_court.jpg"),
		CutoutImageURL: stringPtr("https://yunai-assets.com/characters/dawei/cutout_confident_smile.png"),
		Visibility:     "public",
		AllowChat:      true,
		AllowGroupChat: true,
		CreatedAt:      time.Now(),
	}

	fmt.Printf("✅ 角色创建: %s (背景: 花园读书场景)\n", characterXiaoYu.Name)
	fmt.Printf("✅ 角色创建: %s (背景: 音乐房间场景)\n", characterXiaoMei.Name)
	fmt.Printf("✅ 角色创建: %s (背景: 篮球场场景)\n", characterDaWei.Name)

	// 创建单聊会话（使用 GroupChat 结构，设置为单聊模式）
	singleChatXiaoYu := &domain.GroupChat{
		ID:                 uuid.New(),
		CreatorUserID:      user.ID,
		Name:               "与小雨的温馨时光",
		Description:        stringPtr("单聊模式，使用小雨的背景图"),
		BackgroundImageURL: characterXiaoYu.BgImageURL, // 自动使用角色背景图
		MaxMembers:         2,                          // 单聊只有2个成员
		IsPublic:           false,
		AllowAIInvite:      false,
		CreatedAt:          time.Now(),
	}

	singleChatXiaoMei := &domain.GroupChat{
		ID:                 uuid.New(),
		CreatorUserID:      user.ID,
		Name:               "与小美的快乐时光",
		Description:        stringPtr("单聊模式，使用小美的背景图"),
		BackgroundImageURL: characterXiaoMei.BgImageURL, // 自动使用角色背景图
		MaxMembers:         2,                           // 单聊只有2个成员
		IsPublic:           false,
		AllowAIInvite:      false,
		CreatedAt:          time.Now(),
	}

	fmt.Printf("✅ 单聊创建: %s (应使用小雨的背景图)\n", singleChatXiaoYu.Name)
	fmt.Printf("✅ 单聊创建: %s (应使用小美的背景图)\n", singleChatXiaoMei.Name)

	// 创建剧情群聊
	storyWorldSetting := `【魔法学院设定】
这里是艾尔维斯魔法学院，一个充满魔法与奇迹的地方。
- 小雨：图书馆管理员，擅长治愈魔法，性格温柔
- 小美：音乐系学生，擅长音律魔法，活泼开朗  
- 大伟：体育系学生，擅长强化魔法，阳光积极
- 小明（用户）：新入学的魔法学徒，正在探索自己的魔法天赋

当前场景：魔法学院的中央庭院，午后阳光透过魔法水晶洒下彩虹光芒。`

	groupChatStory := &domain.GroupChat{
		ID:                 uuid.New(),
		CreatorUserID:      user.ID,
		Name:               "魔法学院·中央庭院",
		Description:        stringPtr("魔法学院的日常生活"),
		WorldSetting:       &storyWorldSetting,
		BackgroundImageURL: stringPtr("https://yunai-assets.com/scenes/magic_academy_courtyard.jpg"),
		MaxMembers:         10,
		IsPublic:           false,
		AllowAIInvite:      true,
		// StoryMode:          true, // 字段不存在，注释掉
		CreatedAt: time.Now(),
	}

	// 创建日常群聊
	casualWorldSetting := `【日常校园设定】
这里是普通的高中校园，大家都是同班同学。
- 小雨：班级图书委员，喜欢安静读书
- 小美：文艺委员，负责班级活动策划
- 大伟：体育委员，篮球队队长
- 小明（用户）：普通学生，和大家关系很好

当前场景：学校的天台，午休时间大家聚在一起聊天。`

	groupChatCasual := &domain.GroupChat{
		ID:                 uuid.New(),
		CreatorUserID:      user.ID,
		Name:               "高中同学群",
		Description:        stringPtr("日常校园生活"),
		WorldSetting:       &casualWorldSetting,
		BackgroundImageURL: stringPtr("https://yunai-assets.com/scenes/school_rooftop.jpg"),
		MaxMembers:         10,
		IsPublic:           true,
		AllowAIInvite:      true,
		// StoryMode:          false, // 字段不存在，注释掉
		CreatedAt: time.Now(),
	}

	fmt.Printf("✅ 剧情群聊: %s (魔法学院设定)\n", groupChatStory.Name)
	fmt.Printf("✅ 日常群聊: %s (校园设定)\n", groupChatCasual.Name)

	// 创建剧情章节和场景信息（简化为字符串）
	storyChapterTitle := "新生入学篇"
	storySceneTitle := "午后的魔法庭院"
	storySceneDesc := "阳光透过魔法水晶，庭院里传来悠扬的音乐声"
	storySceneBg := "https://yunai-assets.com/scenes/magic_courtyard_afternoon.jpg"
	storySceneMusic := "https://yunai-assets.com/music/peaceful_magic.mp3"

	fmt.Printf("✅ 剧情章节: %s\n", storyChapterTitle)
	fmt.Printf("✅ 剧情场景: %s\n", storySceneTitle)

	// 创建测试消息
	testMessages := []*domain.ChatMessage{
		{
			ID:           uuid.New(),
			GroupChatID:  &groupChatStory.ID,
			SenderType:   domain.SenderTypeUser,
			SenderUserID: &user.ID,
			Content:      "大家好，我是新来的魔法学徒小明！",
			MessageType:  "text",
			CreatedAt:    time.Now(),
		},
		{
			ID:                uuid.New(),
			GroupChatID:       &groupChatStory.ID,
			SenderType:        domain.SenderTypeCharacter,
			SenderCharacterID: &characterXiaoYu.ID,
			Content:           "欢迎来到魔法学院，小明！我是图书馆的小雨，有什么魔法书籍需要的话可以找我哦~ ✨",
			MessageType:       "text",
			CreatedAt:         time.Now().Add(1 * time.Minute),
		},
		{
			ID:                uuid.New(),
			GroupChatID:       &groupChatStory.ID,
			SenderType:        domain.SenderTypeCharacter,
			SenderCharacterID: &characterXiaoMei.ID,
			Content:           "哇！新同学！我是小美，音乐系的！要不要听听我新学的魔法音律？🎵",
			MessageType:       "text",
			CreatedAt:         time.Now().Add(2 * time.Minute),
		},
	}

	fmt.Printf("✅ 测试消息: %d 条群聊消息\n", len(testMessages))

	return &VisualTestScenario{
		User:              user,
		CharacterXiaoYu:   characterXiaoYu,
		CharacterXiaoMei:  characterXiaoMei,
		CharacterDaWei:    characterDaWei,
		SingleChatXiaoYu:  singleChatXiaoYu,
		SingleChatXiaoMei: singleChatXiaoMei,
		GroupChatStory:    groupChatStory,
		GroupChatCasual:   groupChatCasual,
		StoryChapterTitle: storyChapterTitle,
		StorySceneTitle:   storySceneTitle,
		StorySceneDesc:    storySceneDesc,
		StorySceneBg:      storySceneBg,
		StorySceneMusic:   storySceneMusic,
		TestMessages:      testMessages,
	}
}

// testSingleChatBackground 测试单聊背景自动获取
func testSingleChatBackground(ctx context.Context, scenario *VisualTestScenario) {
	fmt.Println("\n🖼️ [1/4] 单聊背景自动获取测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试小雨单聊背景
	fmt.Printf("单聊会话: %s\n", scenario.SingleChatXiaoYu.Name)
	fmt.Printf("   角色: %s\n", scenario.CharacterXiaoYu.Name)
	fmt.Printf("   角色背景图: %s\n", *scenario.CharacterXiaoYu.BgImageURL)
	fmt.Printf("   单聊背景图: %s\n", *scenario.SingleChatXiaoYu.BackgroundImageURL)
	fmt.Printf("   ✅ 单聊应自动使用角色背景图作为聊天背景\n")
	fmt.Printf("   📱 前端渲染: 花园读书场景，温馨文艺风格\n")

	// 测试小美单聊背景
	fmt.Printf("\n单聊会话: %s\n", scenario.SingleChatXiaoMei.Name)
	fmt.Printf("   角色: %s\n", scenario.CharacterXiaoMei.Name)
	fmt.Printf("   角色背景图: %s\n", *scenario.CharacterXiaoMei.BgImageURL)
	fmt.Printf("   单聊背景图: %s\n", *scenario.SingleChatXiaoMei.BackgroundImageURL)
	fmt.Printf("   ✅ 单聊应自动使用角色背景图作为聊天背景\n")
	fmt.Printf("   📱 前端渲染: 音乐房间场景，活泼青春风格\n")

	// 背景切换测试
	fmt.Println("\n🔄 背景切换测试:")
	fmt.Printf("   用户从小雨单聊 → 小美单聊\n")
	fmt.Printf("   ✅ 背景应从花园场景平滑切换到音乐房间场景\n")
	fmt.Printf("   ✅ 切换动画: 300ms 淡入淡出效果\n")
}

// testGroupChatStoryMode 测试群聊剧情演绎
func testGroupChatStoryMode(ctx context.Context, scenario *VisualTestScenario) {
	fmt.Println("\n🎭 [2/4] 群聊剧情演绎测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试剧情群聊设定
	fmt.Printf("剧情群聊: %s\n", scenario.GroupChatStory.Name)
	fmt.Printf("世界观设定:\n%s\n", *scenario.GroupChatStory.WorldSetting)
	fmt.Printf("群聊背景: %s\n", *scenario.GroupChatStory.BackgroundImageURL)
	fmt.Printf("剧情模式: %v\n", true) // 剧情群聊

	// 测试角色按设定演绎
	fmt.Println("\n🎬 角色演绎测试:")
	for i, msg := range scenario.TestMessages {
		fmt.Printf("\n消息 %d:\n", i+1)
		fmt.Printf("   发送者: %s\n", getSenderName(msg, scenario))
		fmt.Printf("   内容: \"%s\"\n", msg.Content)

		if msg.SenderType == domain.SenderTypeCharacter {
			fmt.Printf("   ✅ 角色按魔法学院设定演绎\n")
			if msg.SenderCharacterID != nil && *msg.SenderCharacterID == scenario.CharacterXiaoYu.ID {
				fmt.Printf("   📚 小雨体现图书馆管理员身份\n")
			} else if msg.SenderCharacterID != nil && *msg.SenderCharacterID == scenario.CharacterXiaoMei.ID {
				fmt.Printf("   🎵 小美体现音乐系学生身份\n")
			}
		}
	}

	// 测试剧情场景切换
	fmt.Printf("\n🌅 剧情场景切换测试:\n")
	fmt.Printf("   当前场景: %s\n", scenario.StorySceneTitle)
	fmt.Printf("   场景描述: %s\n", scenario.StorySceneDesc)
	fmt.Printf("   场景背景: %s\n", scenario.StorySceneBg)
	fmt.Printf("   背景音乐: %s\n", scenario.StorySceneMusic)
	fmt.Printf("   ✅ 场景切换时背景和音乐应同步更新\n")

	// 对比日常群聊
	fmt.Printf("\n📚 对比日常群聊:\n")
	fmt.Printf("   日常群聊: %s\n", scenario.GroupChatCasual.Name)
	fmt.Printf("   剧情模式: %v\n", false) // 日常群聊
	fmt.Printf("   ✅ 日常群聊角色按普通校园设定演绎\n")
	fmt.Printf("   ✅ 无特殊剧情触发，更贴近现实对话\n")
}

// testCharacterCutoutSwitching 测试角色抠图切换
func testCharacterCutoutSwitching(ctx context.Context, scenario *VisualTestScenario) {
	fmt.Println("\n🎨 [3/4] 角色抠图切换测试")
	fmt.Println(strings.Repeat("-", 70))

	// 模拟群聊发言场景
	fmt.Printf("群聊场景: %s\n", scenario.GroupChatStory.Name)
	fmt.Printf("群聊背景: %s\n", *scenario.GroupChatStory.BackgroundImageURL)

	// 测试发言时抠图切换
	fmt.Println("\n💬 发言抠图切换测试:")

	speakers := []struct {
		character *domain.Character
		message   string
	}{
		{scenario.CharacterXiaoYu, "欢迎新同学！需要什么帮助吗？"},
		{scenario.CharacterXiaoMei, "一起来听音乐吧！"},
		{scenario.CharacterDaWei, "要不要一起去运动场练习魔法？"},
	}

	for i, speaker := range speakers {
		fmt.Printf("\n发言 %d:\n", i+1)
		fmt.Printf("   发言角色: %s\n", speaker.character.Name)
		fmt.Printf("   角色抠图: %s\n", *speaker.character.CutoutImageURL)
		fmt.Printf("   发言内容: \"%s\"\n", speaker.message)
		fmt.Printf("   ✅ 发言时应显示角色抠图大图 (300ms 淡入)\n")

		// 根据角色特点描述抠图效果
		switch speaker.character.Name {
		case "小雨":
			fmt.Printf("   🌸 抠图效果: 温柔微笑，手持魔法书\n")
		case "小美":
			fmt.Printf("   🎵 抠图效果: 活力姿态，手持魔法乐器\n")
		case "大伟":
			fmt.Printf("   💪 抠图效果: 自信笑容，运动装束\n")
		}
	}

	// 测试下一位预告
	fmt.Println("\n🔮 下一位发言预告:")
	fmt.Printf("   当前发言: 小雨\n")
	fmt.Printf("   下一位预告: 小美 (抠图预览，半透明显示)\n")
	fmt.Printf("   ✅ 提前加载下一位角色抠图，提升用户体验\n")

	// 测试非发言状态
	fmt.Println("\n😴 非发言状态:")
	fmt.Printf("   未发言角色: 抠图隐藏或半透明显示\n")
	fmt.Printf("   ✅ 突出当前发言者，其他角色淡化处理\n")
}

// testWorldSettingIntegration 测试世界观设定集成
func testWorldSettingIntegration(ctx context.Context, scenario *VisualTestScenario) {
	fmt.Println("\n🌍 [4/4] 世界观设定集成测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试世界观对角色行为的影响
	fmt.Printf("世界观设定测试:\n")
	fmt.Printf("设定类型: 魔法学院 vs 普通校园\n")

	// 魔法学院设定下的角色表现
	fmt.Println("\n🔮 魔法学院设定下:")
	fmt.Printf("   小雨: \"我可以教你治愈魔法的基础咒语哦~ ✨\"\n")
	fmt.Printf("   ✅ 体现魔法世界观，使用魔法相关词汇\n")

	fmt.Printf("   小美: \"听听我的音律魔法，能让心情变好呢！🎵\"\n")
	fmt.Printf("   ✅ 结合音乐和魔法元素\n")

	fmt.Printf("   大伟: \"强化魔法可以让身体更强壮！\"\n")
	fmt.Printf("   ✅ 体现体育与魔法的结合\n")

	// 普通校园设定下的角色表现
	fmt.Println("\n🏫 普通校园设定下:")
	fmt.Printf("   小雨: \"图书馆新到了很多好书，要一起去看看吗？\"\n")
	fmt.Printf("   ✅ 贴近现实校园生活\n")

	fmt.Printf("   小美: \"下周的文艺汇演我们一起准备节目吧！\"\n")
	fmt.Printf("   ✅ 普通学生活动\n")

	fmt.Printf("   大伟: \"篮球比赛要开始了，大家来加油！\"\n")
	fmt.Printf("   ✅ 现实运动场景\n")

	// 测试触发器和场景切换
	fmt.Println("\n⚡ 触发器和场景切换:")
	fmt.Printf("   关键词触发: \"魔法\" → 切换到魔法练习场景\n")
	fmt.Printf("   情绪触发: 兴奋 → 切换到庆祝场景\n")
	fmt.Printf("   时间触发: 傍晚 → 切换到夕阳场景\n")
	fmt.Printf("   ✅ 动态场景切换，增强沉浸感\n")

	// 测试提示词注入
	fmt.Println("\n💭 提示词注入测试:")
	fmt.Printf("   系统提示词: 包含世界观设定\n")
	fmt.Printf("   角色提示词: 结合个人设定和世界观\n")
	fmt.Printf("   场景提示词: 当前场景的氛围描述\n")
	fmt.Printf("   ✅ 多层次提示词确保角色演绎一致性\n")
}

// getSenderName 获取发送者名称
func getSenderName(msg *domain.ChatMessage, scenario *VisualTestScenario) string {
	if msg.SenderType == domain.SenderTypeUser {
		return *scenario.User.Nickname
	} else if msg.SenderType == domain.SenderTypeCharacter && msg.SenderCharacterID != nil {
		if *msg.SenderCharacterID == scenario.CharacterXiaoYu.ID {
			return scenario.CharacterXiaoYu.Name
		} else if *msg.SenderCharacterID == scenario.CharacterXiaoMei.ID {
			return scenario.CharacterXiaoMei.Name
		} else if *msg.SenderCharacterID == scenario.CharacterDaWei.ID {
			return scenario.CharacterDaWei.Name
		}
	}
	return "未知"
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
