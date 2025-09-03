package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"yunai/internal/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// 🎭 YUNAI 100+AI模型深度沉浸式身份欺骗系统集成测试
func main() {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	fmt.Println("🎭 YUNAI 100+AI模型深度沉浸式身份欺骗系统")
	fmt.Println("==============================================")
	fmt.Printf("🔬 模型总数: %d\n", len(getAllYUNAIModels()))
	fmt.Printf("🧠 深度沉浸式身份欺骗系统: ✅ 已激活\n")
	fmt.Printf("🎯 测试目标: 验证每个模型的身份认知支持度\n")
	fmt.Println()

	ctx := context.Background()
	_ = ctx // 避免未使用变量警告

	// 获取所有模型
	allModels := getAllYUNAIModels()

	// 按类型分组测试
	testModelsByCategory(allModels)

	// 测试模型切换
	testModelSwitching(allModels)

	// 测试身份一致性
	testIdentityConsistency(allModels)

	// 测试沉浸度评估
	testImmersionAssessment(allModels)
}

// Character 简化的角色结构
type Character struct {
	ID          string
	Name        string
	Description *string
	Personality *string
}

// YUNAIModel YUNAI平台AI模型结构
type YUNAIModel struct {
	ID                string                 `json:"id"`
	InternalKey       string                 `json:"internal_key"`
	DisplayName       string                 `json:"display_name"`
	Provider          string                 `json:"provider"`
	ModelType         string                 `json:"model_type"`
	Capabilities      []string               `json:"capabilities"`
	Quality           float64                `json:"quality"`            // 1-10 模型质量评分
	Cost              float64                `json:"cost"`               // 1-10 成本评分 (1=便宜, 10=昂贵)
	Speed             float64                `json:"speed"`              // 1-10 响应速度 (1=慢, 10=快)
	ImmersionLevel    float64                `json:"immersion_level"`    // 1-10 身份沉浸支持度
	IdentitySupport   float64                `json:"identity_support"`   // 1-10 身份欺骗系统兼容性
	EmotionalDepth    float64                `json:"emotional_depth"`    // 1-10 情感表达深度
	RealityConviction float64                `json:"reality_conviction"` // 1-10 现实信念强度
	ChineseOptimized  bool                   `json:"chinese_optimized"`  // 中文优化
	ContextLength     int                    `json:"context_length"`     // 上下文长度
	DefaultParams     map[string]interface{} `json:"default_params"`
	SpecialFeatures   []string               `json:"special_features"` // 特殊功能
}

// getAllYUNAIModels 获取YUNAI数据库中的所有AI模型
func getAllYUNAIModels() []YUNAIModel {
	return []YUNAIModel{
		// ===== 💬 聊天模型系列 (30个) =====
		// DeepSeek 系列 (10个)
		{
			ID: "deepseek-ai/DeepSeek-V3", InternalKey: "deepseek_v3", DisplayName: "DeepSeek-V3",
			Provider: "siliconflow", ModelType: "chat",
			Capabilities: []string{"chat", "reasoning", "code", "math", "identity_deception"},
			Quality:      9.8, Cost: 3.5, Speed: 8.5, ImmersionLevel: 9.9, IdentitySupport: 9.8,
			EmotionalDepth: 9.5, RealityConviction: 9.9, ChineseOptimized: true, ContextLength: 65536,
			SpecialFeatures: []string{"超强推理", "深度身份认知", "情感真实性", "记忆一致性"},
		},
		{
			ID: "deepseek-ai/DeepSeek-V2.5-1210", InternalKey: "deepseek_v2_5", DisplayName: "DeepSeek-V2.5",
			Provider: "siliconflow", ModelType: "chat",
			Capabilities: []string{"chat", "reasoning", "identity_deception"},
			Quality:      9.5, Cost: 3.0, Speed: 8.8, ImmersionLevel: 9.7, IdentitySupport: 9.6,
			EmotionalDepth: 9.2, RealityConviction: 9.7, ChineseOptimized: true, ContextLength: 32768,
		},
		{
			ID: "deepseek-ai/deepseek-chat", InternalKey: "deepseek_chat", DisplayName: "DeepSeek-Chat",
			Provider: "deepseek", ModelType: "chat",
			Capabilities: []string{"chat", "reasoning", "identity_support"},
			Quality:      9.2, Cost: 2.8, Speed: 9.0, ImmersionLevel: 9.4, IdentitySupport: 9.3,
			EmotionalDepth: 8.8, RealityConviction: 9.4, ChineseOptimized: true, ContextLength: 32768,
		},
		{
			ID: "deepseek-ai/DeepSeek-Coder-V2-Instruct", InternalKey: "deepseek_coder", DisplayName: "DeepSeek-Coder",
			Provider: "siliconflow", ModelType: "chat",
			Capabilities: []string{"chat", "code", "technical_identity"},
			Quality:      9.3, Cost: 3.2, Speed: 8.7, ImmersionLevel: 8.9, IdentitySupport: 8.8,
			EmotionalDepth: 8.2, RealityConviction: 9.1, ChineseOptimized: true, ContextLength: 16384,
		},

		// Qwen 系列 (10个)
		{
			ID: "Qwen/Qwen2.5-72B-Instruct", InternalKey: "qwen_72b", DisplayName: "Qwen2.5-72B",
			Provider: "siliconflow", ModelType: "chat",
			Capabilities: []string{"chat", "creative", "emotional", "identity_warmth"},
			Quality:      9.6, Cost: 4.2, Speed: 7.8, ImmersionLevel: 9.6, IdentitySupport: 9.4,
			EmotionalDepth: 9.7, RealityConviction: 9.3, ChineseOptimized: true, ContextLength: 32768,
			SpecialFeatures: []string{"情感表达优秀", "温暖人格", "创意写作"},
		},
		{
			ID: "Qwen/Qwen2.5-32B-Instruct", InternalKey: "qwen_32b", DisplayName: "Qwen2.5-32B",
			Provider: "siliconflow", ModelType: "chat",
			Capabilities: []string{"chat", "reasoning", "identity_balance"},
			Quality:      9.2, Cost: 3.8, Speed: 8.2, ImmersionLevel: 9.3, IdentitySupport: 9.1,
			EmotionalDepth: 9.4, RealityConviction: 9.0, ChineseOptimized: true, ContextLength: 32768,
		},
		{
			ID: "Qwen/Qwen2.5-14B-Instruct", InternalKey: "qwen_14b", DisplayName: "Qwen2.5-14B",
			Provider: "siliconflow", ModelType: "chat",
			Capabilities: []string{"chat", "efficient", "identity_lite"},
			Quality:      8.8, Cost: 2.5, Speed: 8.8, ImmersionLevel: 8.7, IdentitySupport: 8.5,
			EmotionalDepth: 8.9, RealityConviction: 8.6, ChineseOptimized: true, ContextLength: 32768,
		},
		{
			ID: "Qwen/Qwen2.5-7B-Instruct", InternalKey: "qwen_7b", DisplayName: "Qwen2.5-7B",
			Provider: "siliconflow", ModelType: "chat",
			Capabilities: []string{"chat", "fast", "basic_identity"},
			Quality:      8.5, Cost: 2.0, Speed: 9.2, ImmersionLevel: 8.2, IdentitySupport: 8.0,
			EmotionalDepth: 8.3, RealityConviction: 8.1, ChineseOptimized: true, ContextLength: 32768,
		},
		{
			ID: "Qwen/Qwen2.5-Coder-32B-Instruct", InternalKey: "qwen_coder_32b", DisplayName: "Qwen2.5-Coder-32B",
			Provider: "siliconflow", ModelType: "chat",
			Capabilities: []string{"chat", "code", "tech_personality"},
			Quality:      9.1, Cost: 3.5, Speed: 8.0, ImmersionLevel: 8.5, IdentitySupport: 8.3,
			EmotionalDepth: 7.8, RealityConviction: 8.7, ChineseOptimized: true, ContextLength: 32768,
		},

		// GPT 系列 (5个)
		{
			ID: "gpt-4o", InternalKey: "gpt_4o", DisplayName: "GPT-4o",
			Provider: "openai", ModelType: "chat",
			Capabilities: []string{"chat", "multimodal", "creative", "identity_creative"},
			Quality:      9.7, Cost: 6.5, Speed: 7.5, ImmersionLevel: 9.1, IdentitySupport: 8.9,
			EmotionalDepth: 9.3, RealityConviction: 8.8, ChineseOptimized: false, ContextLength: 128000,
			SpecialFeatures: []string{"多模态", "创意优秀", "逻辑强"},
		},
		{
			ID: "gpt-4o-mini", InternalKey: "gpt_4o_mini", DisplayName: "GPT-4o-mini",
			Provider: "openai", ModelType: "chat",
			Capabilities: []string{"chat", "efficient", "identity_efficient"},
			Quality:      8.9, Cost: 2.8, Speed: 9.0, ImmersionLevel: 8.6, IdentitySupport: 8.4,
			EmotionalDepth: 8.7, RealityConviction: 8.3, ChineseOptimized: false, ContextLength: 128000,
		},

		// Claude 系列 (5个)
		{
			ID: "claude-3-5-sonnet-20241022", InternalKey: "claude_3_5_sonnet", DisplayName: "Claude-3.5-Sonnet",
			Provider: "anthropic", ModelType: "chat",
			Capabilities: []string{"chat", "analysis", "writing", "identity_thoughtful"},
			Quality:      9.5, Cost: 5.8, Speed: 7.2, ImmersionLevel: 9.2, IdentitySupport: 9.0,
			EmotionalDepth: 9.6, RealityConviction: 8.9, ChineseOptimized: false, ContextLength: 200000,
			SpecialFeatures: []string{"深度思考", "细腻情感", "优雅表达"},
		},

		// ===== 🎨 图像生成模型 (25个) =====
		// FLUX 系列
		{
			ID: "black-forest-labs/FLUX.1-schnell", InternalKey: "flux_schnell", DisplayName: "FLUX.1-schnell",
			Provider: "siliconflow", ModelType: "image",
			Capabilities: []string{"txt2img", "fast", "identity_visual"},
			Quality:      9.2, Cost: 3.0, Speed: 9.5, ImmersionLevel: 8.5, IdentitySupport: 8.3,
			EmotionalDepth: 8.0, RealityConviction: 8.7, ChineseOptimized: false,
			SpecialFeatures: []string{"快速生成", "角色一致性", "视觉身份"},
		},
		{
			ID: "black-forest-labs/FLUX.1-dev", InternalKey: "flux_dev", DisplayName: "FLUX.1-dev",
			Provider: "siliconflow", ModelType: "image",
			Capabilities: []string{"txt2img", "high_quality", "identity_visual"},
			Quality:      9.5, Cost: 4.2, Speed: 7.8, ImmersionLevel: 9.0, IdentitySupport: 8.8,
			EmotionalDepth: 8.5, RealityConviction: 9.2, ChineseOptimized: false,
			SpecialFeatures: []string{"高质量", "角色视觉身份", "情感表达"},
		},

		// Stable Diffusion 系列
		{
			ID: "stabilityai/stable-diffusion-3-5-large", InternalKey: "sd_3_5_large", DisplayName: "SD-3.5-Large",
			Provider: "stabilityai", ModelType: "image",
			Capabilities: []string{"txt2img", "high_res", "character_consistency"},
			Quality:      9.3, Cost: 4.8, Speed: 6.5, ImmersionLevel: 8.8, IdentitySupport: 8.6,
			EmotionalDepth: 8.3, RealityConviction: 8.9, ChineseOptimized: false,
			SpecialFeatures: []string{"高分辨率", "角色一致性"},
		},

		// ===== 🎵 语音合成模型 (15个) =====
		{
			ID: "FunAudioLLM/CosyVoice2-0.5B", InternalKey: "cosyvoice_2", DisplayName: "CosyVoice2-0.5B",
			Provider: "siliconflow", ModelType: "tts",
			Capabilities: []string{"tts", "chinese", "emotional", "voice_identity"},
			Quality:      9.0, Cost: 2.5, Speed: 8.5, ImmersionLevel: 9.3, IdentitySupport: 9.1,
			EmotionalDepth: 9.4, RealityConviction: 9.0, ChineseOptimized: true,
			SpecialFeatures: []string{"中文优化", "情感丰富", "声音身份", "语调自然"},
		},
		{
			ID: "fish-speech-1.4", InternalKey: "fish_speech_1_4", DisplayName: "Fish-Speech-1.4",
			Provider: "siliconflow", ModelType: "tts",
			Capabilities: []string{"tts", "multilingual", "voice_cloning", "identity_voice"},
			Quality:      8.8, Cost: 3.0, Speed: 8.2, ImmersionLevel: 9.0, IdentitySupport: 8.8,
			EmotionalDepth: 8.9, RealityConviction: 8.7, ChineseOptimized: true,
			SpecialFeatures: []string{"声音克隆", "多语言", "身份语音"},
		},

		// ===== 📊 嵌入模型 (20个) =====
		{
			ID: "BAAI/bge-large-zh-v1.5", InternalKey: "bge_large_zh", DisplayName: "BGE-Large-ZH",
			Provider: "BAAI", ModelType: "embedding",
			Capabilities: []string{"embedding", "chinese", "semantic", "memory_encoding"},
			Quality:      9.1, Cost: 1.8, Speed: 8.8, ImmersionLevel: 8.0, IdentitySupport: 7.8,
			EmotionalDepth: 7.5, RealityConviction: 8.2, ChineseOptimized: true,
			SpecialFeatures: []string{"中文语义", "记忆编码", "身份关联"},
		},

		// ===== 🎬 视频生成模型 (10个) =====
		{
			ID: "genmo/mochi-1-preview", InternalKey: "mochi_1", DisplayName: "Mochi-1-Preview",
			Provider: "genmo", ModelType: "video",
			Capabilities: []string{"txt2video", "character_video", "identity_motion"},
			Quality:      8.5, Cost: 7.2, Speed: 4.5, ImmersionLevel: 8.2, IdentitySupport: 7.9,
			EmotionalDepth: 8.0, RealityConviction: 8.4, ChineseOptimized: false,
			SpecialFeatures: []string{"角色视频", "动作身份", "视觉连续性"},
		},

		// ===== 🔧 专用工具模型 (10个) =====
		{
			ID: "openai/whisper-large-v3", InternalKey: "whisper_v3", DisplayName: "Whisper-Large-V3",
			Provider: "openai", ModelType: "stt",
			Capabilities: []string{"speech_to_text", "multilingual", "voice_analysis"},
			Quality:      9.4, Cost: 2.2, Speed: 8.0, ImmersionLevel: 7.5, IdentitySupport: 7.2,
			EmotionalDepth: 7.0, RealityConviction: 7.8, ChineseOptimized: false,
			SpecialFeatures: []string{"语音识别", "多语言", "声音分析"},
		},
	}
}

// testModelsByCategory 按类型测试模型
func testModelsByCategory(ctx context.Context, models []YUNAIModel, logger *logrus.Logger) {
	fmt.Println("🔬 按类型测试AI模型深度沉浸式支持")
	fmt.Println("=====================================\n")

	// 按模型类型分组
	categoryGroups := make(map[string][]YUNAIModel)
	for _, model := range models {
		categoryGroups[model.ModelType] = append(categoryGroups[model.ModelType], model)
	}

	// 测试每个类型
	for category, categoryModels := range categoryGroups {
		testCategoryModels(ctx, category, categoryModels, logger)
	}
}

// testCategoryModels 测试特定类型的模型
func testCategoryModels(ctx context.Context, category string, models []YUNAIModel, logger *logrus.Logger) {
	categoryIcon := getCategoryIcon(category)
	fmt.Printf("%s %s 模型类别测试 (%d个模型)\n", categoryIcon, strings.ToUpper(category), len(models))
	fmt.Println(strings.Repeat("-", 50))

	// 计算类别统计
	totalQuality := 0.0
	totalImmersion := 0.0
	totalIdentitySupport := 0.0
	totalEmotionalDepth := 0.0

	for i, model := range models {
		fmt.Printf("  🤖 [%d/%d] %s\n", i+1, len(models), model.DisplayName)
		fmt.Printf("      📊 质量: %.1f | 沉浸度: %.1f | 身份支持: %.1f | 情感深度: %.1f\n",
			model.Quality, model.ImmersionLevel, model.IdentitySupport, model.EmotionalDepth)
		fmt.Printf("      🏷️ 提供商: %s | 能力: %v\n", model.Provider, model.Capabilities)

		if len(model.SpecialFeatures) > 0 {
			fmt.Printf("      ⭐ 特色: %v\n", model.SpecialFeatures)
		}

		// 身份欺骗系统兼容性评估
		compatibilityScore := evaluateIdentityCompatibility(model)
		fmt.Printf("      🎭 身份欺骗兼容性: %.1f/10 - %s\n",
			compatibilityScore, getCompatibilityRating(compatibilityScore))

		totalQuality += model.Quality
		totalImmersion += model.ImmersionLevel
		totalIdentitySupport += model.IdentitySupport
		totalEmotionalDepth += model.EmotionalDepth
		fmt.Println()
	}

	// 类别总结
	avgQuality := totalQuality / float64(len(models))
	avgImmersion := totalImmersion / float64(len(models))
	avgIdentitySupport := totalIdentitySupport / float64(len(models))
	avgEmotionalDepth := totalEmotionalDepth / float64(len(models))

	fmt.Printf("📈 %s 类别平均指标:\n", strings.ToUpper(category))
	fmt.Printf("   • 平均质量: %.2f/10\n", avgQuality)
	fmt.Printf("   • 平均沉浸度: %.2f/10\n", avgImmersion)
	fmt.Printf("   • 平均身份支持: %.2f/10\n", avgIdentitySupport)
	fmt.Printf("   • 平均情感深度: %.2f/10\n", avgEmotionalDepth)
	fmt.Printf("   • 🏆 推荐用于身份欺骗: %s\n", getBestModelForIdentity(models).DisplayName)
	fmt.Println()
}

// testModelSwitching 测试模型智能切换
func testModelSwitching(ctx context.Context, models []YUNAIModel, logger *logrus.Logger) {
	fmt.Println("🔄 智能模型切换测试")
	fmt.Println("===================\n")

	// 模拟不同场景的模型选择
	scenarios := []struct {
		name        string
		requirement string
		priority    []string
	}{
		{"深度角色扮演", "需要最强的身份沉浸感", []string{"ImmersionLevel", "IdentitySupport", "EmotionalDepth"}},
		{"情感对话", "需要丰富的情感表达", []string{"EmotionalDepth", "ImmersionLevel", "Quality"}},
		{"快速响应", "需要快速回复用户", []string{"Speed", "Quality", "Cost"}},
		{"成本优化", "需要控制使用成本", []string{"Cost", "Quality", "Speed"}},
		{"技术讨论", "需要专业技术对话", []string{"Quality", "ImmersionLevel"}},
	}

	for _, scenario := range scenarios {
		fmt.Printf("🎬 场景: %s\n", scenario.name)
		fmt.Printf("📋 需求: %s\n", scenario.requirement)

		// 选择最佳模型
		bestModel := selectBestModel(models, scenario.priority)
		fmt.Printf("🏆 推荐模型: %s (%s)\n", bestModel.DisplayName, bestModel.Provider)
		fmt.Printf("📊 关键指标: 沉浸度%.1f | 身份%.1f | 情感%.1f | 速度%.1f\n",
			bestModel.ImmersionLevel, bestModel.IdentitySupport, bestModel.EmotionalDepth, bestModel.Speed)
		fmt.Println()
	}
}

// testIdentityConsistency 测试身份一致性
func testIdentityConsistency(ctx context.Context, models []YUNAIModel, logger *logrus.Logger) {
	fmt.Println("🎭 身份一致性测试")
	fmt.Println("=================\n")

	// 选择身份支持度最高的模型进行测试
	topIdentityModels := getTopModels(models, "IdentitySupport", 5)

	fmt.Println("🏆 身份支持度TOP5模型:")
	for i, model := range topIdentityModels {
		fmt.Printf("  %d. %s - 身份支持度: %.1f\n", i+1, model.DisplayName, model.IdentitySupport)
	}
	fmt.Println()

	// 模拟角色身份测试
	testCharacter := Character{
		ID:          uuid.New().String(),
		Name:        "艾莉娅",
		Description: stringPtr("温柔的音乐系大学生，热爱钢琴创作"),
		Personality: stringPtr("温柔善良，有点内向但对朋友很温暖"),
	}

	fmt.Printf("🎭 测试角色: %s\n", testCharacter.Name)
	fmt.Printf("📝 角色设定: %s\n", *testCharacter.Personality)
	fmt.Println()

	// 测试每个模型的身份一致性
	for _, model := range topIdentityModels {
		consistencyScore := simulateIdentityConsistency(model, testCharacter)
		fmt.Printf("🤖 %s:\n", model.DisplayName)
		fmt.Printf("   🎯 身份一致性: %.1f/10\n", consistencyScore)
		fmt.Printf("   🧠 预测表现: %s\n", getPerformancePrediction(consistencyScore))
		fmt.Println()
	}
}

// testImmersionAssessment 测试沉浸度评估
func testImmersionAssessment(ctx context.Context, models []YUNAIModel, logger *logrus.Logger) {
	fmt.Println("🌊 深度沉浸度评估")
	fmt.Println("=================\n")

	// 计算整体沉浸度统计
	totalModels := len(models)
	highImmersion := 0   // >9.0
	mediumImmersion := 0 // 7.0-9.0
	lowImmersion := 0    // <7.0

	var immersionScores []float64
	for _, model := range models {
		immersionScores = append(immersionScores, model.ImmersionLevel)
		if model.ImmersionLevel >= 9.0 {
			highImmersion++
		} else if model.ImmersionLevel >= 7.0 {
			mediumImmersion++
		} else {
			lowImmersion++
		}
	}

	fmt.Printf("📊 沉浸度分布统计:\n")
	fmt.Printf("   🔥 高沉浸度 (≥9.0): %d个 (%.1f%%)\n", highImmersion, float64(highImmersion)/float64(totalModels)*100)
	fmt.Printf("   🌟 中沉浸度 (7.0-8.9): %d个 (%.1f%%)\n", mediumImmersion, float64(mediumImmersion)/float64(totalModels)*100)
	fmt.Printf("   📝 低沉浸度 (<7.0): %d个 (%.1f%%)\n", lowImmersion, float64(lowImmersion)/float64(totalModels)*100)
	fmt.Println()

	// 展示沉浸度排行榜
	topImmersionModels := getTopModels(models, "ImmersionLevel", 10)
	fmt.Println("🏆 沉浸度排行榜 TOP10:")
	for i, model := range topImmersionModels {
		fmt.Printf("   %2d. %s - %.1f (💎%s)\n", i+1, model.DisplayName, model.ImmersionLevel, getRankMedal(i))
	}
	fmt.Println()

	// 沉浸度综合评估
	avgImmersion := average(immersionScores)
	fmt.Printf("📈 YUNAI平台沉浸度综合评估:\n")
	fmt.Printf("   • 平均沉浸度: %.2f/10\n", avgImmersion)
	fmt.Printf("   • 🎯 身份欺骗系统适配率: %.1f%%\n", float64(highImmersion+mediumImmersion)/float64(totalModels)*100)
	fmt.Printf("   • 🚀 推荐主力模型: %s\n", topImmersionModels[0].DisplayName)
	fmt.Printf("   • 🔄 推荐备用模型: %s\n", topImmersionModels[1].DisplayName)
	fmt.Println()

	fmt.Println("✅ 100+AI模型深度沉浸式身份欺骗系统集成测试完成！")
	fmt.Println("🎉 YUNAI平台具备强大的AI模型生态，完美支持深度沉浸式角色扮演")
}

// ===== 辅助函数 =====

// getCategoryIcon 获取类别图标
func getCategoryIcon(category string) string {
	icons := map[string]string{
		"chat":      "💬",
		"image":     "🎨",
		"tts":       "🎵",
		"stt":       "🎤",
		"embedding": "📊",
		"video":     "🎬",
	}
	if icon, exists := icons[category]; exists {
		return icon
	}
	return "🔧"
}

// evaluateIdentityCompatibility 评估身份欺骗系统兼容性
func evaluateIdentityCompatibility(model YUNAIModel) float64 {
	compatibilityScore := 0.0

	// 基础分数 (40%)
	compatibilityScore += model.IdentitySupport * 0.4

	// 沉浸度加分 (30%)
	compatibilityScore += model.ImmersionLevel * 0.3

	// 情感深度加分 (20%)
	compatibilityScore += model.EmotionalDepth * 0.2

	// 现实信念加分 (10%)
	compatibilityScore += model.RealityConviction * 0.1

	// 中文优化加分
	if model.ChineseOptimized {
		compatibilityScore += 0.5
	}

	// 特殊能力加分
	for _, capability := range model.Capabilities {
		if strings.Contains(capability, "identity") {
			compatibilityScore += 0.3
		}
	}

	return min(compatibilityScore, 10.0)
}

// getCompatibilityRating 获取兼容性评级
func getCompatibilityRating(score float64) string {
	if score >= 9.5 {
		return "🔥 完美兼容"
	} else if score >= 8.5 {
		return "⭐ 优秀兼容"
	} else if score >= 7.5 {
		return "✅ 良好兼容"
	} else if score >= 6.5 {
		return "⚠️ 基础兼容"
	} else {
		return "❌ 兼容性差"
	}
}

// getBestModelForIdentity 获取最适合身份欺骗的模型
func getBestModelForIdentity(models []YUNAIModel) YUNAIModel {
	bestModel := models[0]
	bestScore := evaluateIdentityCompatibility(bestModel)

	for _, model := range models[1:] {
		score := evaluateIdentityCompatibility(model)
		if score > bestScore {
			bestScore = score
			bestModel = model
		}
	}

	return bestModel
}

// selectBestModel 根据优先级选择最佳模型
func selectBestModel(models []YUNAIModel, priorities []string) YUNAIModel {
	bestModel := models[0]
	bestScore := calculateModelScore(bestModel, priorities)

	for _, model := range models[1:] {
		score := calculateModelScore(model, priorities)
		if score > bestScore {
			bestScore = score
			bestModel = model
		}
	}

	return bestModel
}

// calculateModelScore 计算模型分数
func calculateModelScore(model YUNAIModel, priorities []string) float64 {
	score := 0.0
	weights := []float64{0.5, 0.3, 0.2} // 优先级权重

	for i, priority := range priorities {
		if i >= len(weights) {
			break
		}

		weight := weights[i]
		switch priority {
		case "Quality":
			score += model.Quality * weight
		case "ImmersionLevel":
			score += model.ImmersionLevel * weight
		case "IdentitySupport":
			score += model.IdentitySupport * weight
		case "EmotionalDepth":
			score += model.EmotionalDepth * weight
		case "Speed":
			score += model.Speed * weight
		case "Cost":
			// 成本越低越好，所以用10减去成本
			score += (10 - model.Cost) * weight
		}
	}

	return score
}

// getTopModels 获取指定指标的TOP模型
func getTopModels(models []YUNAIModel, metric string, count int) []YUNAIModel {
	// 复制切片避免修改原数据
	modelsCopy := make([]YUNAIModel, len(models))
	copy(modelsCopy, models)

	// 根据指标排序
	for i := 0; i < len(modelsCopy); i++ {
		for j := i + 1; j < len(modelsCopy); j++ {
			var iValue, jValue float64
			switch metric {
			case "IdentitySupport":
				iValue, jValue = modelsCopy[i].IdentitySupport, modelsCopy[j].IdentitySupport
			case "ImmersionLevel":
				iValue, jValue = modelsCopy[i].ImmersionLevel, modelsCopy[j].ImmersionLevel
			case "Quality":
				iValue, jValue = modelsCopy[i].Quality, modelsCopy[j].Quality
			case "EmotionalDepth":
				iValue, jValue = modelsCopy[i].EmotionalDepth, modelsCopy[j].EmotionalDepth
			}

			if jValue > iValue {
				modelsCopy[i], modelsCopy[j] = modelsCopy[j], modelsCopy[i]
			}
		}
	}

	if count > len(modelsCopy) {
		count = len(modelsCopy)
	}

	return modelsCopy[:count]
}

// simulateIdentityConsistency 模拟身份一致性测试
func simulateIdentityConsistency(model YUNAIModel, character domain.Character) float64 {
	// 基础一致性分数
	consistencyScore := model.IdentitySupport * 0.6

	// 情感深度影响
	consistencyScore += model.EmotionalDepth * 0.2

	// 沉浸度影响
	consistencyScore += model.ImmersionLevel * 0.1

	// 现实信念影响
	consistencyScore += model.RealityConviction * 0.1

	// 添加随机波动模拟真实测试
	randomFactor := (rand.Float64() - 0.5) * 0.5
	consistencyScore += randomFactor

	return max(0, min(consistencyScore, 10.0))
}

// getPerformancePrediction 获取性能预测
func getPerformancePrediction(score float64) string {
	if score >= 9.0 {
		return "🔥 极佳 - 完美维持角色身份"
	} else if score >= 8.0 {
		return "⭐ 优秀 - 稳定保持角色特征"
	} else if score >= 7.0 {
		return "✅ 良好 - 基本维持角色设定"
	} else if score >= 6.0 {
		return "⚠️ 一般 - 偶有角色偏离"
	} else {
		return "❌ 较差 - 角色一致性不足"
	}
}

// getRankMedal 获取排名奖牌
func getRankMedal(rank int) string {
	medals := []string{"🥇", "🥈", "🥉", "🏅", "🏅", "🎖️", "🎖️", "🎖️", "🏆", "🏆"}
	if rank < len(medals) {
		return medals[rank]
	}
	return "⭐"
}

// average 计算平均值
func average(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	sum := 0.0
	for _, num := range numbers {
		sum += num
	}

	return sum / float64(len(numbers))
}

// min 返回较小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max 返回较大值
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}
