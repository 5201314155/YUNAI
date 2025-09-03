package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SiliconFlow模型获取和分析工具

type SiliconFlowModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type SiliconFlowResponse struct {
	Object string              `json:"object"`
	Data   []SiliconFlowModel `json:"data"`
}

func main() {
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"
	baseURL := "https://api.siliconflow.cn/v1"

	fmt.Println("🔍 获取SiliconFlow所有可用模型")
	fmt.Println("===============================================================================")

	// 获取所有类型的模型
	modelTypes := []string{"text", "image", "audio", "video"}
	allModels := make(map[string][]SiliconFlowModel)

	for _, modelType := range modelTypes {
		fmt.Printf("\n📋 获取 %s 类型模型...\n", modelType)
		
		models, err := fetchModelsByType(baseURL, apiKey, modelType)
		if err != nil {
			fmt.Printf("   ❌ 获取失败: %v\n", err)
			continue
		}

		allModels[modelType] = models
		fmt.Printf("   ✅ 获取成功: %d个模型\n", len(models))
		
		// 显示模型列表
		for _, model := range models {
			fmt.Printf("      - %s\n", model.ID)
		}
	}

	// 分析和分类模型
	fmt.Println("\n===============================================================================")
	fmt.Println("🤖 模型分析和分类")
	fmt.Println("===============================================================================")

	categorizedModels := categorizeModels(allModels)
	
	// 输出分类结果
	for category, models := range categorizedModels {
		fmt.Printf("\n📂 %s (%d个模型)\n", category, len(models))
		for _, model := range models {
			fmt.Printf("   🔹 %s\n", model.ID)
		}
	}

	// 生成YUNAI项目的模型配置
	fmt.Println("\n===============================================================================")
	fmt.Println("⚙️ 生成YUNAI项目模型配置")
	fmt.Println("===============================================================================")

	generateYUNAIModelConfigs(categorizedModels)
}

func fetchModelsByType(baseURL, apiKey, modelType string) ([]SiliconFlowModel, error) {
	url := fmt.Sprintf("%s/models?type=%s", baseURL, modelType)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败 (状态码: %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response SiliconFlowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func categorizeModels(allModels map[string][]SiliconFlowModel) map[string][]SiliconFlowModel {
	categorized := make(map[string][]SiliconFlowModel)

	for modelType, models := range allModels {
		switch modelType {
		case "text":
			// 进一步细分文本模型
			for _, model := range models {
				if isLLMModel(model.ID) {
					categorized["💬 对话聊天模型"] = append(categorized["💬 对话聊天模型"], model)
				} else if isCodeModel(model.ID) {
					categorized["💻 代码生成模型"] = append(categorized["💻 代码生成模型"], model)
				} else if isEmbeddingModel(model.ID) {
					categorized["🔍 文本嵌入模型"] = append(categorized["🔍 文本嵌入模型"], model)
				} else if isRerankerModel(model.ID) {
					categorized["📊 重排序模型"] = append(categorized["📊 重排序模型"], model)
				} else {
					categorized["📝 通用文本模型"] = append(categorized["📝 通用文本模型"], model)
				}
			}
		case "image":
			categorized["🎨 图像生成模型"] = append(categorized["🎨 图像生成模型"], models...)
		case "audio":
			categorized["🎙️ 语音处理模型"] = append(categorized["🎙️ 语音处理模型"], models...)
		case "video":
			categorized["🎬 视频生成模型"] = append(categorized["🎬 视频生成模型"], models...)
		}
	}

	return categorized
}

func isLLMModel(modelID string) bool {
	llmKeywords := []string{"chat", "llama", "qwen", "deepseek", "gpt", "claude", "gemini", "yi", "baichuan"}
	for _, keyword := range llmKeywords {
		if contains(modelID, keyword) {
			return true
		}
	}
	return false
}

func isCodeModel(modelID string) bool {
	codeKeywords := []string{"code", "coder", "codellama", "starcoder", "codegeex"}
	for _, keyword := range codeKeywords {
		if contains(modelID, keyword) {
			return true
		}
	}
	return false
}

func isEmbeddingModel(modelID string) bool {
	embeddingKeywords := []string{"embedding", "embed", "bge", "gte", "text-embedding"}
	for _, keyword := range embeddingKeywords {
		if contains(modelID, keyword) {
			return true
		}
	}
	return false
}

func isRerankerModel(modelID string) bool {
	rerankerKeywords := []string{"reranker", "rerank", "bge-reranker"}
	for _, keyword := range rerankerKeywords {
		if contains(modelID, keyword) {
			return true
		}
	}
	return false
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || 
		    (len(str) > len(substr) && 
		     (str[:len(substr)] == substr || 
		      str[len(str)-len(substr):] == substr ||
		      findInString(str, substr))))
}

func findInString(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func generateYUNAIModelConfigs(categorizedModels map[string][]SiliconFlowModel) {
	fmt.Println("\n🔧 为YUNAI项目生成模型配置...")

	for category, models := range categorizedModels {
		fmt.Printf("\n📂 %s\n", category)
		
		for _, model := range models {
			config := generateModelConfig(model, category)
			fmt.Printf("   🔹 %s\n", config.DisplayName)
			fmt.Printf("      内部键: %s\n", config.InternalKey)
			fmt.Printf("      用途: %s\n", config.YUNAIUsage)
			fmt.Printf("      优化场景: %s\n", config.OptimizationScenario)
		}
	}
}

type YUNAIModelConfig struct {
	InternalKey          string  `json:"internal_key"`
	DisplayName          string  `json:"display_name"`
	Provider             string  `json:"provider"`
	ModelType            string  `json:"model_type"`
	YUNAIUsage           string  `json:"yunai_usage"`
	OptimizationScenario string  `json:"optimization_scenario"`
	PricingTier          string  `json:"pricing_tier"`
	RecommendedFor       []string `json:"recommended_for"`
}

func generateModelConfig(model SiliconFlowModel, category string) YUNAIModelConfig {
	config := YUNAIModelConfig{
		InternalKey: model.ID,
		DisplayName: getDisplayName(model.ID),
		Provider:    "siliconflow",
	}

	switch category {
	case "💬 对话聊天模型":
		config.ModelType = "chat"
		config.YUNAIUsage = "用户与AI角色的智能对话、群聊互动、情感交流"
		config.OptimizationScenario = "提升对话质量、角色一致性、情感理解能力"
		config.RecommendedFor = []string{"单聊对话", "群聊互动", "角色扮演", "情感陪伴"}
		
	case "💻 代码生成模型":
		config.ModelType = "code"
		config.YUNAIUsage = "AI角色的编程助手功能、代码生成、技术讨论"
		config.OptimizationScenario = "提升代码质量、编程教学、技术问答准确性"
		config.RecommendedFor = []string{"编程助手角色", "技术讨论", "代码生成", "编程教学"}
		
	case "🔍 文本嵌入模型":
		config.ModelType = "embedding"
		config.YUNAIUsage = "用户身份识别、内容相似度匹配、智能推荐"
		config.OptimizationScenario = "提升身份识别准确率、内容推荐精度"
		config.RecommendedFor = []string{"身份识别", "内容推荐", "相似度匹配", "智能搜索"}
		
	case "📊 重排序模型":
		config.ModelType = "reranker"
		config.YUNAIUsage = "搜索结果优化、内容排序、推荐系统"
		config.OptimizationScenario = "提升搜索准确性、优化推荐排序"
		config.RecommendedFor = []string{"搜索优化", "内容排序", "推荐系统"}
		
	case "🎨 图像生成模型":
		config.ModelType = "image"
		config.YUNAIUsage = "AI角色头像生成、朋友圈图片、场景背景图"
		config.OptimizationScenario = "提升图像质量、风格一致性、生成速度"
		config.RecommendedFor = []string{"角色头像", "朋友圈图片", "背景图生成", "表情包"}
		
	case "🎙️ 语音处理模型":
		config.ModelType = "audio"
		config.YUNAIUsage = "语音通话、语音消息、文字转语音"
		config.OptimizationScenario = "提升语音质量、降低延迟、增强自然度"
		config.RecommendedFor = []string{"语音通话", "语音消息", "TTS", "STT"}
		
	case "🎬 视频生成模型":
		config.ModelType = "video"
		config.YUNAIUsage = "动态表情、短视频生成、场景动画"
		config.OptimizationScenario = "提升视频质量、减少生成时间"
		config.RecommendedFor = []string{"动态表情", "短视频", "场景动画"}
	}

	// 根据模型名称设置定价层级
	if contains(model.ID, "gpt-4") || contains(model.ID, "claude") {
		config.PricingTier = "premium"
	} else if contains(model.ID, "deepseek") || contains(model.ID, "qwen") {
		config.PricingTier = "standard"
	} else {
		config.PricingTier = "basic"
	}

	return config
}

func getDisplayName(modelID string) string {
	// 将模型ID转换为友好的显示名称
	displayNames := map[string]string{
		"deepseek-ai/DeepSeek-V3":                    "DeepSeek V3 - 超强推理模型",
		"deepseek-ai/deepseek-llm-67b-chat":         "DeepSeek LLM 67B - 大型对话模型",
		"deepseek-ai/deepseek-coder-33b-instruct":   "DeepSeek Coder 33B - 专业代码模型",
		"Qwen/Qwen2.5-72B-Instruct":                 "通义千问 2.5 72B - 中文优化",
		"Qwen/Qwen2.5-Coder-32B-Instruct":          "通义千问 Coder 32B - 代码专家",
		"meta-llama/Llama-3.1-405B-Instruct":       "Llama 3.1 405B - 超大规模模型",
		"meta-llama/Llama-3.1-70B-Instruct":        "Llama 3.1 70B - 平衡性能模型",
		"meta-llama/Llama-3.1-8B-Instruct":         "Llama 3.1 8B - 轻量级模型",
		"01-ai/Yi-1.5-34B-Chat":                     "零一万物 Yi 1.5 34B - 中文对话",
		"01-ai/Yi-1.5-9B-Chat":                      "零一万物 Yi 1.5 9B - 高效对话",
		"google/gemma-2-27b-it":                     "Gemma 2 27B - Google开源模型",
		"microsoft/DialoGPT-medium":                 "DialoGPT - 对话专用模型",
		"stabilityai/stable-diffusion-xl-base-1.0":  "Stable Diffusion XL - 图像生成",
		"stabilityai/stable-diffusion-3-5-large":    "Stable Diffusion 3.5 - 最新图像模型",
		"black-forest-labs/FLUX.1-schnell":         "FLUX.1 Schnell - 快速图像生成",
		"black-forest-labs/FLUX.1-dev":             "FLUX.1 Dev - 开发版图像模型",
		"runwayml/stable-video-diffusion-img2vid-xt": "Stable Video Diffusion - 图转视频",
		"BAAI/bge-large-zh-v1.5":                   "BGE Large 中文 - 文本嵌入",
		"BAAI/bge-reranker-v2-m3":                  "BGE Reranker - 重排序模型",
		"FishAudio/fish-speech-1.4":               "Fish Speech - 语音合成",
		"openai/whisper-large-v3":                 "Whisper Large V3 - 语音识别",
	}

	if displayName, exists := displayNames[modelID]; exists {
		return displayName
	}

	// 如果没有预定义名称，生成一个友好的名称
	return modelID
}
