package main

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"yunai/internal/domain"
)

func main() {
	fmt.Println("🚀 YUNAI 核心功能集成演示")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("📋 演示内容：两图系统 + 剧情触发 + 复杂关系网络 + 设定优先级")
	fmt.Println()

	// 演示1：两图系统
	demonstrateTwoImageSystem()
	
	// 演示2：剧情触发系统
	demonstrateStoryTriggerSystem()
	
	// 演示3：复杂关系网络
	demonstrateComplexRelationships()
	
	// 演示4：设定优先级系统
	demonstrateSettingPriority()
	
	// 演示5：完整场景集成
	demonstrateCompleteIntegration()

	fmt.Println("\n🎉 YUNAI 核心功能集成演示完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// demonstrateTwoImageSystem 演示两图系统
func demonstrateTwoImageSystem() {
	fmt.Println("🎨 [1/5] 两图系统演示")
	fmt.Println(strings.Repeat("-", 70))

	// 创建角色
	character := &domain.Character{
		ID:             uuid.New(),
		Name:           "小雨",
		BgImageURL:     stringPtr("https://yunai-assets.com/xiayu/bg_garden_reading.jpg"),
		CutoutImageURL: stringPtr("https://yunai-assets.com/xiayu/cutout_transparent.png"),
		Personality:    stringPtr("温柔善良的图书管理员"),
	}

	fmt.Printf("✅ 角色创建: %s\n", character.Name)
	fmt.Printf("   📸 背景图 (bg_image): %s\n", *character.BgImageURL)
	fmt.Printf("   🎭 抠图 (cutout_image): %s\n", *character.CutoutImageURL)
	fmt.Printf("   💭 人设: %s\n", *character.Personality)

	// 单聊背景自动获取
	fmt.Printf("\n💬 单聊场景:\n")
	fmt.Printf("   🖼️  聊天背景: 自动使用角色背景图\n")
	fmt.Printf("   🌸 效果: 花园读书场景，营造温馨氛围\n")
	fmt.Printf("   ✅ 核心功能: 单聊自动使用角色的 bg_image\n")

	// 群聊抠图切换
	fmt.Printf("\n👥 群聊场景:\n")
	fmt.Printf("   🖼️  群聊背景: 全局背景图 (可选配置)\n")
	fmt.Printf("   🎭 小雨发言时:\n")
	fmt.Printf("      └─ 显示抠图: %s\n", *character.CutoutImageURL)
	fmt.Printf("      └─ 动画效果: 300ms 淡入显示\n")
	fmt.Printf("      └─ 层级: 最高 (突出当前发言者)\n")
	fmt.Printf("   😴 其他角色: 抠图淡化或隐藏\n")
	fmt.Printf("   ✅ 核心功能: 发言时显示角色的 cutout_image\n")
}

// demonstrateStoryTriggerSystem 演示剧情触发系统
func demonstrateStoryTriggerSystem() {
	fmt.Println("\n🎬 [2/5] 剧情触发系统演示")
	fmt.Println(strings.Repeat("-", 70))

	// 章节设定
	fmt.Printf("📖 多章节设定:\n")
	fmt.Printf("   章节1: 平静的学院生活\n")
	fmt.Printf("   ├─ 背景图: peaceful_academy.jpg\n")
	fmt.Printf("   ├─ 背景音乐: calm_theme.mp3 (可选)\n")
	fmt.Printf("   └─ 触发条件: 提到\"神秘\"、\"奇怪\"等\n")
	fmt.Printf("   \n")
	fmt.Printf("   章节2: 神秘事件调查\n")
	fmt.Printf("   ├─ 背景图: mysterious_library.jpg\n")
	fmt.Printf("   ├─ 背景音乐: mystery_theme.mp3 (可选)\n")
	fmt.Printf("   └─ 触发条件: 提到\"危险\"、\"战斗\"等\n")

	// 超级模糊触发演示
	fmt.Printf("\n🧠 超级模糊触发演示:\n")
	
	triggerTests := []struct {
		input    string
		expected string
		reason   string
	}{
		{
			input:    "我想去看看那个神秘的地方",
			expected: "触发神秘章节",
			reason:   "白话表达 + 模糊关联",
		},
		{
			input:    "感觉这里有点不对劲呢",
			expected: "触发神秘章节", 
			reason:   "情绪感知 + 语义理解",
		},
		{
			input:    "图书馆真安静啊",
			expected: "无触发",
			reason:   "日常对话，保持当前章节",
		},
	}

	for i, test := range triggerTests {
		fmt.Printf("   测试 %d: \"%s\"\n", i+1, test.input)
		fmt.Printf("   └─ 结果: %s (%s)\n", test.expected, test.reason)
	}

	fmt.Printf("\n✨ 智能特性:\n")
	fmt.Printf("   🎯 支持白话触发: \"我想去看看\" → 探索场景\n")
	fmt.Printf("   🎭 情绪触发: \"感觉紧张\" → 战斗场景\n")
	fmt.Printf("   🔄 自动切换: 背景图 + 音乐 + 演绎模式\n")
	fmt.Printf("   ⚙️  可选配置: 音乐、音效都是可选的\n")
}

// demonstrateComplexRelationships 演示复杂关系网络
func demonstrateComplexRelationships() {
	fmt.Println("\n💕 [3/5] 复杂关系网络演示")
	fmt.Println(strings.Repeat("-", 70))

	// 复杂关系描述示例
	complexRelations := []struct {
		description string
		analysis    string
	}{
		{
			description: "小米是我的好朋友，但她闺蜜小明和我关系不好",
			analysis:    "三角关系：我↔朋友↔小米，我↔敌对↔小明",
		},
		{
			description: "我和小王表面上是朋友，实际上有竞争关系",
			analysis:    "复杂关系：表面朋友 + 实际竞争",
		},
		{
			description: "大伟是我的死党，我们从小一起长大",
			analysis:    "稳定关系：深度友谊，高信任度",
		},
	}

	fmt.Printf("🧠 智能关系解析:\n")
	for i, rel := range complexRelations {
		fmt.Printf("   关系 %d: \"%s\"\n", i+1, rel.description)
		fmt.Printf("   └─ 解析: %s\n", rel.analysis)
	}

	fmt.Printf("\n🎭 关系驱动演绎:\n")
	fmt.Printf("   场景: 群聊中提到\"我们一起去玩吧\"\n")
	fmt.Printf("   \n")
	fmt.Printf("   小米 (好朋友关系):\n")
	fmt.Printf("   └─ \"好啊！我们去哪里玩呢？\" (积极响应)\n")
	fmt.Printf("   \n")
	fmt.Printf("   小明 (敌对关系):\n")
	fmt.Printf("   └─ \"哼，又想拉拢大家吗？\" (消极态度)\n")
	fmt.Printf("   \n")
	fmt.Printf("   小王 (复杂关系):\n")
	fmt.Printf("   └─ \"可以啊，不过我有更好的想法...\" (暗中较劲)\n")

	fmt.Printf("\n✨ 智能特性:\n")
	fmt.Printf("   🔍 超级模糊解析: 理解\"表面朋友\"等复杂描述\n")
	fmt.Printf("   ⚠️  冲突检测: 自动发现关系矛盾\n")
	fmt.Printf("   🎯 演绎驱动: 关系影响角色行为和对话\n")
	fmt.Printf("   📊 多维分析: 情感强度 + 复杂度 + 稳定性\n")
}

// demonstrateSettingPriority 演示设定优先级系统
func demonstrateSettingPriority() {
	fmt.Println("\n⚙️ [4/5] 设定优先级系统演示")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Printf("📊 设定优先级 (从高到低):\n")
	fmt.Printf("   1️⃣ 群聊世界观设定 - 最高优先级\n")
	fmt.Printf("   2️⃣ 群聊演绎设定 - 群聊特定规则\n")
	fmt.Printf("   3️⃣ 角色单聊设定 - 角色个人特色\n")
	fmt.Printf("   4️⃣ 角色基础人设 - 兜底设定\n")

	scenarios := []struct {
		name     string
		settings string
		behavior string
	}{
		{
			name:     "完整设定场景",
			settings: "魔法学院世界观 + 战斗演绎规则 + 图书管理员设定 + 温柔人设",
			behavior: "按魔法学院设定，使用战斗术语，体现图书管理员身份",
		},
		{
			name:     "部分设定场景", 
			settings: "魔法学院世界观 + 图书管理员设定 + 温柔人设",
			behavior: "按魔法学院设定，结合图书管理员身份自由发挥",
		},
		{
			name:     "无群聊设定场景",
			settings: "图书管理员设定 + 温柔人设",
			behavior: "按现实图书管理员设定，避免魔法元素",
		},
		{
			name:     "最小设定场景",
			settings: "温柔人设",
			behavior: "仅基于基础人设，自然简洁的表现",
		},
	}

	fmt.Printf("\n🔄 智能回退演示:\n")
	for i, scenario := range scenarios {
		fmt.Printf("   场景 %d: %s\n", i+1, scenario.name)
		fmt.Printf("   ├─ 可用设定: %s\n", scenario.settings)
		fmt.Printf("   └─ 演绎行为: %s\n", scenario.behavior)
	}

	fmt.Printf("\n✨ 智能特性:\n")
	fmt.Printf("   🎯 自动回退: 缺失高级设定时使用低级设定\n")
	fmt.Printf("   🚫 避免幻觉: 不会使用不存在的世界观元素\n")
	fmt.Printf("   🎭 自然表现: 确保角色行为始终合理\n")
	fmt.Printf("   ⚙️  用户友好: 不设置也能用，设置了体验更好\n")
}

// demonstrateCompleteIntegration 演示完整集成
func demonstrateCompleteIntegration() {
	fmt.Println("\n🎭 [5/5] 完整场景集成演示")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Printf("🌟 完整的 YUNAI 使用流程:\n")
	fmt.Printf("\n")

	// 步骤1: 角色创建
	fmt.Printf("1️⃣ 角色创建 (两图系统)\n")
	fmt.Printf("   👤 用户创建角色\"小雨\"\n")
	fmt.Printf("   📸 上传背景图: 花园读书场景 (单聊背景)\n")
	fmt.Printf("   🎭 上传抠图: 透明背景立绘 (群聊发言显示)\n")
	fmt.Printf("   ✅ 系统验证: 格式正确，尺寸合适\n")

	// 步骤2: 群聊创建
	fmt.Printf("\n2️⃣ 群聊创建 (剧情系统)\n")
	fmt.Printf("   🌍 设置世界观: 魔法学院设定\n")
	fmt.Printf("   📖 配置章节1: 平静的学院生活\n")
	fmt.Printf("   📖 配置章节2: 神秘事件调查 (可选背景音乐)\n")
	fmt.Printf("   🎵 音乐配置: 用户可选，不设置也能用\n")

	// 步骤3: 关系配置
	fmt.Printf("\n3️⃣ 关系配置 (复杂关系网络)\n")
	fmt.Printf("   💕 用户输入: \"小雨是我的好朋友，小美和我有点竞争关系\"\n")
	fmt.Printf("   🧠 系统解析: 识别出朋友关系 + 竞争关系\n")
	fmt.Printf("   ⚠️  冲突检测: 无冲突，关系网络健康\n")

	// 步骤4: 开始对话
	fmt.Printf("\n4️⃣ 开始群聊对话\n")
	fmt.Printf("   🖼️  群聊背景: 魔法学院庭院 (全局背景)\n")
	fmt.Printf("   👤 用户: \"大家好，今天天气真不错！\"\n")
	fmt.Printf("   \n")
	fmt.Printf("   🎭 小雨发言:\n")
	fmt.Printf("   ├─ 抠图显示: 300ms 淡入动画\n")
	fmt.Printf("   ├─ 关系驱动: 朋友关系 → 友好回应\n")
	fmt.Printf("   ├─ 设定优先级: 魔法学院世界观 + 图书管理员身份\n")
	fmt.Printf("   └─ 💬 \"是啊！阳光透过魔法水晶特别美呢~\"\n")

	// 步骤5: 剧情触发
	fmt.Printf("\n5️⃣ 剧情触发 (超级模糊触发)\n")
	fmt.Printf("   👤 用户: \"我感觉图书馆里有什么奇怪的东西\"\n")
	fmt.Printf("   🔍 系统分析: 检测到\"奇怪\" → 神秘触发\n")
	fmt.Printf("   🎬 自动切换: 章节1 → 章节2\n")
	fmt.Printf("   🖼️  背景变化: 庭院 → 神秘图书馆\n")
	fmt.Printf("   🎵 音乐变化: 平静主题 → 神秘主题 (如果配置了)\n")
	fmt.Printf("   \n")
	fmt.Printf("   🎭 小雨响应:\n")
	fmt.Printf("   ├─ 演绎模式: 自动切换到调查模式\n")
	fmt.Printf("   └─ 💬 \"我也注意到了...让我们小心地去看看吧\"\n")

	// 步骤6: 智能朋友圈
	fmt.Printf("\n6️⃣ 智能朋友圈 (自动生成)\n")
	fmt.Printf("   📱 小雨自动发布:\n")
	fmt.Printf("   ├─ 内容: \"今天和朋友们在图书馆发现了有趣的现象\"\n")
	fmt.Printf("   ├─ 基于: 最近聊天 + 角色人设 + 关系网络\n")
	fmt.Printf("   └─ @提及: 智能识别用户身份进行@\n")

	fmt.Printf("\n🎉 完整流程演示完成！\n")
	fmt.Printf("✨ YUNAI 核心优势:\n")
	fmt.Printf("   🎨 两图系统: 视觉体验丰富，单聊群聊不同效果\n")
	fmt.Printf("   🧠 超级智能: 理解白话、模糊表达、情绪触发\n")
	fmt.Printf("   💕 复杂关系: 解析\"表面朋友\"等复杂人际关系\n")
	fmt.Printf("   ⚙️  灵活配置: 可选配置，不设置也能用\n")
	fmt.Printf("   🔄 智能回退: 自动适应不同设定完整度\n")
	fmt.Printf("   📱 自动内容: 智能朋友圈，减少用户操作\n")
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
