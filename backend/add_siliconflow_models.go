package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// SiliconFlow API响应结构
type SiliconFlowResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// YUNAI模型结构
type YUNAIModel struct {
	InternalKey       string `json:"internal_key"`
	DisplayName       string `json:"display_name"`
	Provider          string `json:"provider"`
	ModelType         string `json:"model_type"`
	Category          string `json:"category"`
	Description       string `json:"description"`
	Capabilities      string `json:"capabilities"`
	ParamsSchema      string `json:"params_schema"`
	SystemPrompt      string `json:"system_prompt"`
	Pricing           string `json:"pricing"`
	MaxTokens         int    `json:"max_tokens"`
	SupportStreaming  bool   `json:"support_streaming"`
	IsActive          bool   `json:"is_active"`
	IsFeatured        bool   `json:"is_featured"`
	Weight            int    `json:"weight"`
}

func main() {
	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 获取SiliconFlow模型
	models, err := fetchSiliconFlowModels()
	if err != nil {
		log.Fatal("获取SiliconFlow模型失败:", err)
	}

	fmt.Printf("🚀 成功获取到 %d 个SiliconFlow模型\n\n", len(models))

	// 转换为YUNAI模型格式并插入数据库
	insertedCount := 0
	for i, model := range models {
		yunaiModel := convertToYUNAIModel(model, i+1)
		
		err := insertModel(db, yunaiModel)
		if err != nil {
			log.Printf("插入模型 %s 失败: %v", yunaiModel.InternalKey, err)
			continue
		}
		
		insertedCount++
		fmt.Printf("✅ [%d/%d] 已添加: %s - %s\n", insertedCount, len(models), yunaiModel.InternalKey, yunaiModel.DisplayName)
	}

	fmt.Printf("\n🎉 成功添加 %d 个SiliconFlow模型到数据库！\n", insertedCount)
	
	// 显示分类统计
	showCategoryStats(db)
}

func fetchSiliconFlowModels() ([]Model, error) {
	apiURL := "https://api.siliconflow.cn/v1/models"
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	req, err := http.NewRequest("GET", apiURL, nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var sfResponse SiliconFlowResponse
	if err := json.Unmarshal(body, &sfResponse); err != nil {
		return nil, err
	}

	return sfResponse.Data, nil
}

func convertToYUNAIModel(model Model, sortOrder int) YUNAIModel {
	category := categorizeModel(model.ID)
	modelType := getModelType(model.ID)
	displayName := generateDisplayName(model.ID)
	description := generateDescription(model.ID, category)
	capabilities := generateCapabilities(model.ID, modelType)
	paramsSchema := generateParamsSchema(modelType)
	systemPrompt := generateSystemPrompt(model.ID, modelType)
	pricing := generatePricing(model.ID)
	maxTokens := getMaxTokens(model.ID)
	supportStreaming := getSupportStreaming(modelType)

	return YUNAIModel{
		InternalKey:       generateInternalKey(model.ID),
		DisplayName:       displayName,
		Provider:          "siliconflow",
		ModelType:         modelType,
		Category:          category,
		Description:       description,
		Capabilities:      capabilities,
		ParamsSchema:      paramsSchema,
		SystemPrompt:      systemPrompt,
		Pricing:           pricing,
		MaxTokens:         maxTokens,
		SupportStreaming:  supportStreaming,
		IsActive:          true,
		IsFeatured:        sortOrder <= 10, // 前10个设为推荐
		Weight:            110 - sortOrder,  // 权重递减
	}
}

func categorizeModel(modelID string) string {
	modelID = strings.ToLower(modelID)
	
	if strings.Contains(modelID, "gpt") || strings.Contains(modelID, "chatgpt") {
		return "OpenAI系列"
	} else if strings.Contains(modelID, "claude") {
		return "Anthropic系列"
	} else if strings.Contains(modelID, "llama") {
		return "Meta开源"
	} else if strings.Contains(modelID, "qwen") || strings.Contains(modelID, "通义") {
		return "阿里通义"
	} else if strings.Contains(modelID, "baichuan") || strings.Contains(modelID, "百川") {
		return "百川智能"
	} else if strings.Contains(modelID, "chatglm") || strings.Contains(modelID, "glm") {
		return "清华GLM"
	} else if strings.Contains(modelID, "deepseek") {
		return "深度求索"
	} else if strings.Contains(modelID, "embedding") || strings.Contains(modelID, "bge") {
		return "嵌入模型"
	} else if strings.Contains(modelID, "flux") || strings.Contains(modelID, "stable-diffusion") || strings.Contains(modelID, "kolors") {
		return "图像生成"
	} else if strings.Contains(modelID, "fish") || strings.Contains(modelID, "sensevoice") || strings.Contains(modelID, "cosyvoice") {
		return "语音模型"
	} else if strings.Contains(modelID, "wan") && strings.Contains(modelID, "video") {
		return "视频生成"
	} else if strings.Contains(modelID, "reranker") {
		return "重排序"
	} else if strings.Contains(modelID, "hunyuan") {
		return "腾讯混元"
	} else if strings.Contains(modelID, "ernie") {
		return "百度文心"
	} else if strings.Contains(modelID, "kimi") {
		return "月之暗面"
	} else {
		return "通用模型"
	}
}

func getModelType(modelID string) string {
	modelID = strings.ToLower(modelID)
	
	if strings.Contains(modelID, "embedding") || strings.Contains(modelID, "bge") {
		return "embedding"
	} else if strings.Contains(modelID, "reranker") {
		return "reranking"
	} else if strings.Contains(modelID, "flux") || strings.Contains(modelID, "stable-diffusion") || strings.Contains(modelID, "kolors") {
		return "image"
	} else if strings.Contains(modelID, "fish") || strings.Contains(modelID, "sensevoice") || strings.Contains(modelID, "cosyvoice") {
		return "audio"
	} else if strings.Contains(modelID, "wan") && (strings.Contains(modelID, "t2v") || strings.Contains(modelID, "i2v")) {
		return "video"
	} else {
		return "chat"
	}
}

func generateDisplayName(modelID string) string {
	parts := strings.Split(modelID, "/")
	if len(parts) > 1 {
		name := parts[len(parts)-1]
		// 美化名称
		name = strings.ReplaceAll(name, "-", " ")
		name = strings.ReplaceAll(name, "_", " ")
		return strings.Title(name)
	}
	return modelID
}

func generateDescription(modelID string, category string) string {
	modelID = strings.ToLower(modelID)
	
	descriptions := map[string]string{
		"gpt":           "OpenAI GPT系列模型，具有强大的对话和推理能力，适用于各种自然语言处理任务",
		"claude":        "Anthropic Claude系列模型，注重安全性和有用性，擅长长文本理解和分析",
		"llama":         "Meta开源的大语言模型，性能优异且可商用，支持多种下游任务",
		"qwen":          "阿里巴巴通义千问系列，中文理解能力突出，支持多模态和长文本处理",
		"baichuan":      "百川智能开源模型，专为中文优化，在中文任务上表现优秀",
		"chatglm":       "清华大学开源的中英双语模型，支持对话和代码生成",
		"deepseek":      "深度求索模型，专业的编程助手，在代码生成和推理方面表现出色",
		"embedding":     "文本嵌入模型，用于语义搜索、相似度计算和向量检索",
		"flux":          "Black Forest Labs图像生成模型，支持高质量的文本到图像生成",
		"stable-diffusion": "Stability AI的经典图像生成模型，开源且功能强大",
		"fish":          "Fish Audio语音合成模型，支持自然流畅的语音生成",
		"sensevoice":    "阿里巴巴语音识别模型，支持多语言实时语音识别",
		"wan":           "Wan AI视频生成模型，支持文本到视频和图像到视频生成",
		"reranker":      "重排序模型，用于优化搜索结果的相关性排序",
		"hunyuan":       "腾讯混元大模型，企业级AI服务，支持多种商业应用场景",
		"ernie":         "百度文心大模型，中文知识丰富，适用于各种中文NLP任务",
		"kimi":          "月之暗面Kimi模型，支持超长文本处理，擅长文档分析和总结",
	}
	
	for key, desc := range descriptions {
		if strings.Contains(modelID, key) {
			return desc
		}
	}
	
	return fmt.Sprintf("SiliconFlow提供的%s，适用于各种AI应用场景", category)
}

func generateCapabilities(modelID string, modelType string) string {
	modelID = strings.ToLower(modelID)
	
	switch modelType {
	case "chat":
		caps := []string{"chat", "text_generation"}
		if strings.Contains(modelID, "code") || strings.Contains(modelID, "deepseek") {
			caps = append(caps, "code_generation", "programming")
		}
		if strings.Contains(modelID, "qwen") || strings.Contains(modelID, "glm") || strings.Contains(modelID, "baichuan") {
			caps = append(caps, "chinese", "multilingual")
		}
		if strings.Contains(modelID, "reasoning") || strings.Contains(modelID, "deepseek") {
			caps = append(caps, "reasoning", "analysis")
		}
		return fmt.Sprintf("[\"%s\"]", strings.Join(caps, "\", \""))
	case "embedding":
		return "[\"embedding\", \"semantic_search\", \"similarity\", \"vector_retrieval\"]"
	case "image":
		return "[\"image_generation\", \"text_to_image\", \"creative_art\", \"design\"]"
	case "audio":
		if strings.Contains(modelID, "speech") || strings.Contains(modelID, "fish") {
			return "[\"tts\", \"voice_synthesis\", \"speech_generation\"]"
		}
		return "[\"asr\", \"voice_recognition\", \"speech_to_text\"]"
	case "video":
		return "[\"video_generation\", \"text_to_video\", \"animation\", \"creative_video\"]"
	case "reranking":
		return "[\"reranking\", \"search_optimization\", \"relevance_scoring\"]"
	default:
		return "[\"general\", \"ai_assistant\"]"
	}
}

func generateParamsSchema(modelType string) string {
	switch modelType {
	case "chat":
		return `{
			"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2, "description": "控制输出的随机性"},
			"max_tokens": {"type": "integer", "default": 2000, "min": 1, "max": 8000, "description": "最大输出token数"},
			"top_p": {"type": "number", "default": 0.9, "min": 0, "max": 1, "description": "核采样参数"},
			"frequency_penalty": {"type": "number", "default": 0, "min": -2, "max": 2, "description": "频率惩罚"},
			"presence_penalty": {"type": "number", "default": 0, "min": -2, "max": 2, "description": "存在惩罚"},
			"stream": {"type": "boolean", "default": false, "description": "是否流式输出"}
		}`
	case "embedding":
		return `{
			"dimensions": {"type": "integer", "default": 1024, "description": "向量维度"},
			"normalize": {"type": "boolean", "default": true, "description": "是否归一化"},
			"batch_size": {"type": "integer", "default": 32, "min": 1, "max": 100, "description": "批处理大小"}
		}`
	case "image":
		return `{
			"width": {"type": "integer", "default": 1024, "min": 256, "max": 2048, "description": "图像宽度"},
			"height": {"type": "integer", "default": 1024, "min": 256, "max": 2048, "description": "图像高度"},
			"steps": {"type": "integer", "default": 20, "min": 1, "max": 100, "description": "生成步数"},
			"guidance_scale": {"type": "number", "default": 7.5, "min": 1, "max": 20, "description": "引导强度"},
			"seed": {"type": "integer", "default": -1, "description": "随机种子"}
		}`
	case "audio":
		return `{
			"voice_id": {"type": "string", "default": "default", "description": "声音ID"},
			"speed": {"type": "number", "default": 1.0, "min": 0.5, "max": 2.0, "description": "语速"},
			"pitch": {"type": "number", "default": 1.0, "min": 0.5, "max": 2.0, "description": "音调"},
			"emotion": {"type": "string", "default": "neutral", "description": "情感"}
		}`
	case "video":
		return `{
			"duration": {"type": "integer", "default": 5, "min": 1, "max": 30, "description": "视频时长(秒)"},
			"fps": {"type": "integer", "default": 24, "min": 12, "max": 60, "description": "帧率"},
			"resolution": {"type": "string", "default": "720p", "description": "分辨率"},
			"style": {"type": "string", "default": "realistic", "description": "视频风格"}
		}`
	default:
		return `{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}}`
	}
}

func generateSystemPrompt(modelID string, modelType string) string {
	modelID = strings.ToLower(modelID)
	
	if modelType == "chat" {
		if strings.Contains(modelID, "deepseek") {
			return "你是DeepSeek AI助手，擅长逻辑推理和代码生成。请提供准确、有用的回答。"
		} else if strings.Contains(modelID, "qwen") {
			return "你是通义千问，阿里巴巴开发的AI助手。请用中文回答，提供准确、有帮助的信息。"
		} else if strings.Contains(modelID, "glm") {
			return "你是GLM AI助手，由清华大学开发。请提供专业、准确的回答。"
		} else if strings.Contains(modelID, "claude") {
			return "你是Claude，由Anthropic开发的AI助手。请提供有帮助、无害、诚实的回答。"
		} else if strings.Contains(modelID, "gpt") {
			return "你是GPT AI助手，请提供准确、有用的回答。"
		}
		return "你是一个有用的AI助手，请提供准确、有帮助的回答。"
	}
	
	return fmt.Sprintf("这是一个%s模型，专门用于相关任务处理。", modelType)
}

func generatePricing(modelID string) string {
	modelID = strings.ToLower(modelID)
	
	// 根据模型类型和复杂度设置价格
	if strings.Contains(modelID, "gpt-4") {
		return `{"input_token_price": 0.105, "output_token_price": 0.315, "unit": "1k_tokens", "currency": "CNY"}`
	} else if strings.Contains(modelID, "gpt-3.5") {
		return `{"input_token_price": 0.0105, "output_token_price": 0.021, "unit": "1k_tokens", "currency": "CNY"}`
	} else if strings.Contains(modelID, "claude") {
		return `{"input_token_price": 0.084, "output_token_price": 0.252, "unit": "1k_tokens", "currency": "CNY"}`
	} else if strings.Contains(modelID, "deepseek") {
		return `{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}`
	} else if strings.Contains(modelID, "embedding") || strings.Contains(modelID, "bge") {
		return `{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}`
	} else if strings.Contains(modelID, "flux") || strings.Contains(modelID, "stable") {
		return `{"input_token_price": 0.014, "output_token_price": 0.014, "unit": "1k_tokens", "currency": "CNY"}`
	} else if strings.Contains(modelID, "video") || strings.Contains(modelID, "wan") {
		return `{"input_token_price": 0.028, "output_token_price": 0.028, "unit": "1k_tokens", "currency": "CNY"}`
	} else {
		return `{"input_token_price": 0.014, "output_token_price": 0.014, "unit": "1k_tokens", "currency": "CNY"}`
	}
}

func getMaxTokens(modelID string) int {
	modelID = strings.ToLower(modelID)
	
	if strings.Contains(modelID, "32k") {
		return 32768
	} else if strings.Contains(modelID, "16k") {
		return 16384
	} else if strings.Contains(modelID, "8k") {
		return 8192
	} else if strings.Contains(modelID, "4k") {
		return 4096
	} else if strings.Contains(modelID, "128k") {
		return 131072
	} else if strings.Contains(modelID, "200k") {
		return 200000
	} else if strings.Contains(modelID, "kimi") {
		return 200000 // Kimi支持长文本
	} else {
		return 8192 // 默认值
	}
}

func getSupportStreaming(modelType string) bool {
	return modelType == "chat" // 只有聊天模型支持流式
}

func generateInternalKey(modelID string) string {
	// 生成内部键名
	key := strings.ReplaceAll(modelID, "/", "_")
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ToLower(key)
	return "sf_" + key
}

func insertModel(db *sql.DB, model YUNAIModel) error {
	query := `
		INSERT INTO ai_models (
			internal_key, display_name, provider, model_type, category, description,
			capabilities, params_schema, model_system_prompt, pricing, max_tokens,
			support_streaming, is_active, is_featured, weight, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW()
		) ON CONFLICT (internal_key) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			category = EXCLUDED.category,
			description = EXCLUDED.description,
			capabilities = EXCLUDED.capabilities,
			params_schema = EXCLUDED.params_schema,
			model_system_prompt = EXCLUDED.model_system_prompt,
			pricing = EXCLUDED.pricing,
			max_tokens = EXCLUDED.max_tokens,
			support_streaming = EXCLUDED.support_streaming,
			weight = EXCLUDED.weight,
			updated_at = NOW()
	`
	
	_, err := db.Exec(query,
		model.InternalKey, model.DisplayName, model.Provider, model.ModelType,
		model.Category, model.Description, model.Capabilities, model.ParamsSchema,
		model.SystemPrompt, model.Pricing, model.MaxTokens, model.SupportStreaming,
		model.IsActive, model.IsFeatured, model.Weight,
	)
	
	return err
}

func showCategoryStats(db *sql.DB) {
	fmt.Println("\n📊 模型分类统计:")
	fmt.Println("-------------------------------------------")
	
	query := `
		SELECT category, model_type, COUNT(*) as count
		FROM ai_models 
		WHERE provider = 'siliconflow' AND is_active = true
		GROUP BY category, model_type
		ORDER BY category, count DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询统计失败: %v", err)
		return
	}
	defer rows.Close()
	
	currentCategory := ""
	totalCount := 0
	
	for rows.Next() {
		var category, modelType string
		var count int
		
		if err := rows.Scan(&category, &modelType, &count); err != nil {
			continue
		}
		
		if category != currentCategory {
			if currentCategory != "" {
				fmt.Println()
			}
			fmt.Printf("📁 %s:\n", category)
			currentCategory = category
		}
		
		fmt.Printf("   %s: %d个模型\n", modelType, count)
		totalCount += count
	}
	
	fmt.Printf("\n🎯 总计: %d个SiliconFlow模型已添加到系统\n", totalCount)
}
