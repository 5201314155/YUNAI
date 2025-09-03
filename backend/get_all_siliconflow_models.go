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

func main() {
	fmt.Println("🚀 获取SiliconFlow所有100+模型")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// SiliconFlow API配置
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	// 获取所有模型
	fmt.Println("📡 正在获取所有SiliconFlow模型...")
	allModels, err := getAllModels(apiKey)
	if err != nil {
		log.Fatal("获取模型失败:", err)
	}

	fmt.Printf("✅ 成功获取到 %d 个模型\n\n", len(allModels))

	// 按类型分类统计
	typeCount := make(map[string]int)
	categoryCount := make(map[string]int)

	fmt.Println("📊 模型分类统计:")
	fmt.Println("-------------------------------------------")

	for _, model := range allModels {
		modelType := getModelType(model.ID)
		category := categorizeModel(model.ID)
		
		typeCount[modelType]++
		categoryCount[category]++
	}

	// 显示类型统计
	fmt.Println("按类型统计:")
	for modelType, count := range typeCount {
		fmt.Printf("  %s: %d个\n", modelType, count)
	}

	fmt.Println("\n按厂商分类:")
	for category, count := range categoryCount {
		fmt.Printf("  %s: %d个\n", category, count)
	}

	// 显示所有模型列表
	fmt.Println("\n📋 完整模型列表:")
	fmt.Println("-------------------------------------------")
	
	currentCategory := ""
	for i, model := range allModels {
		category := categorizeModel(model.ID)
		modelType := getModelType(model.ID)
		
		if category != currentCategory {
			if currentCategory != "" {
				fmt.Println()
			}
			fmt.Printf("📁 %s:\n", category)
			currentCategory = category
		}
		
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, model.ID, modelType, generateDisplayName(model.ID))
	}

	// 添加所有模型到数据库
	fmt.Println("\n💾 添加所有模型到数据库...")
	insertedCount := 0
	for i, model := range allModels {
		yunaiModel := convertToYUNAIModel(model, i+1)
		
		err := insertOrUpdateModel(db, yunaiModel)
		if err != nil {
			log.Printf("插入模型 %s 失败: %v", yunaiModel.InternalKey, err)
			continue
		}
		
		insertedCount++
		if insertedCount%20 == 0 || insertedCount <= 10 {
			fmt.Printf("✅ [%d/%d] 已添加: %s\n", insertedCount, len(allModels), yunaiModel.DisplayName)
		}
	}

	fmt.Printf("\n🎉 成功添加 %d 个SiliconFlow模型到数据库！\n", insertedCount)
}

func getAllModels(apiKey string) ([]Model, error) {
	url := "https://api.siliconflow.cn/v1/models"
	
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

func categorizeModel(modelID string) string {
	modelID = strings.ToLower(modelID)
	
	categories := map[string]string{
		"gpt":           "OpenAI系列",
		"claude":        "Anthropic系列", 
		"qwen":          "阿里通义",
		"glm":           "清华GLM",
		"deepseek":      "深度求索",
		"llama":         "Meta开源",
		"baichuan":      "百川智能",
		"embedding":     "嵌入模型",
		"bge":           "嵌入模型",
		"flux":          "图像生成",
		"stable":        "图像生成",
		"kolors":        "图像生成",
		"fish":          "语音模型",
		"cosyvoice":     "语音模型",
		"sensevoice":    "语音模型",
		"wan":           "视频生成",
		"video":         "视频生成",
		"reranker":      "重排序",
		"hunyuan":       "腾讯混元",
		"ernie":         "百度文心",
		"kimi":          "月之暗面",
		"minimax":       "MiniMax",
		"step":          "阶跃星辰",
		"internlm":      "上海AI实验室",
		"yi":            "零一万物",
	}
	
	for key, category := range categories {
		if strings.Contains(modelID, key) {
			return category
		}
	}
	
	return "通用模型"
}

func getModelType(modelID string) string {
	modelID = strings.ToLower(modelID)
	
	if strings.Contains(modelID, "embedding") || strings.Contains(modelID, "bge") {
		return "embedding"
	} else if strings.Contains(modelID, "reranker") {
		return "reranking"
	} else if strings.Contains(modelID, "flux") || strings.Contains(modelID, "stable") || strings.Contains(modelID, "kolors") {
		return "image"
	} else if strings.Contains(modelID, "fish") || strings.Contains(modelID, "cosyvoice") || strings.Contains(modelID, "sensevoice") {
		return "audio"
	} else if strings.Contains(modelID, "wan") || strings.Contains(modelID, "video") {
		return "video"
	} else {
		return "chat"
	}
}

func generateDisplayName(modelID string) string {
	parts := strings.Split(modelID, "/")
	if len(parts) > 1 {
		name := parts[len(parts)-1]
		name = strings.ReplaceAll(name, "-", " ")
		name = strings.ReplaceAll(name, "_", " ")
		return strings.Title(name)
	}
	return modelID
}

func generateInternalKey(modelID string) string {
	key := strings.ReplaceAll(modelID, "/", "_")
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ToLower(key)
	return "sf_" + key
}

type YUNAIModel struct {
	InternalKey       string
	DisplayName       string
	CustomDisplayName string // 后台可自定义的显示名称
	Provider          string
	ModelType         string
	Category          string
	Description       string
	Capabilities      string
	ParamsSchema      string
	SystemPrompt      string
	Pricing           string
	MaxTokens         int
	SupportStreaming  bool
	IsActive          bool
	IsFeatured        bool
	Weight            int
}

func convertToYUNAIModel(model Model, sortOrder int) YUNAIModel {
	category := categorizeModel(model.ID)
	modelType := getModelType(model.ID)
	displayName := generateDisplayName(model.ID)

	return YUNAIModel{
		InternalKey:       generateInternalKey(model.ID),
		DisplayName:       displayName,
		CustomDisplayName: displayName, // 默认与原名相同，后台可修改
		Provider:          "siliconflow",
		ModelType:         modelType,
		Category:          category,
		Description:       fmt.Sprintf("SiliconFlow提供的%s，来自%s", displayName, category),
		Capabilities:      generateCapabilities(modelType),
		ParamsSchema:      generateParamsSchema(modelType),
		SystemPrompt:      generateSystemPrompt(modelType),
		Pricing:           `{"input_token_price": 0.014, "output_token_price": 0.014, "unit": "1k_tokens", "currency": "CNY"}`,
		MaxTokens:         8192,
		SupportStreaming:  modelType == "chat",
		IsActive:          true,
		IsFeatured:        sortOrder <= 50, // 前50个设为推荐
		Weight:            200 - sortOrder,
	}
}

func generateCapabilities(modelType string) string {
	switch modelType {
	case "chat":
		return `["chat", "text_generation", "conversation", "reasoning"]`
	case "embedding":
		return `["embedding", "semantic_search", "similarity", "vector_retrieval"]`
	case "image":
		return `["image_generation", "text_to_image", "creative_art", "design"]`
	case "audio":
		return `["audio_processing", "speech_synthesis", "voice_generation", "tts", "asr"]`
	case "video":
		return `["video_generation", "text_to_video", "animation"]`
	case "reranking":
		return `["reranking", "search_optimization", "relevance_scoring"]`
	default:
		return `["general", "ai_assistant"]`
	}
}

func generateParamsSchema(modelType string) string {
	switch modelType {
	case "chat":
		return `{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 2000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}`
	case "embedding":
		return `{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}`
	case "image":
		return `{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 20}}`
	case "audio":
		return `{"voice_id": {"type": "string", "default": "default"}, "speed": {"type": "number", "default": 1.0}}`
	case "video":
		return `{"duration": {"type": "integer", "default": 5}, "fps": {"type": "integer", "default": 24}}`
	default:
		return `{"temperature": {"type": "number", "default": 0.7}}`
	}
}

func generateSystemPrompt(modelType string) string {
	if modelType == "chat" {
		return "你是一个有用的AI助手，请提供准确、有帮助的回答。"
	}
	return fmt.Sprintf("这是一个%s模型，专门用于相关任务处理。", modelType)
}

func insertOrUpdateModel(db *sql.DB, model YUNAIModel) error {
	query := `
		INSERT INTO ai_models (
			internal_key, display_name, custom_display_name, provider, model_type, category, description,
			capabilities, params_schema, model_system_prompt, pricing, max_tokens,
			support_streaming, is_active, is_featured, weight, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
		) ON CONFLICT (internal_key) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			custom_display_name = COALESCE(ai_models.custom_display_name, EXCLUDED.custom_display_name),
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
		model.InternalKey, model.DisplayName, model.CustomDisplayName, model.Provider, model.ModelType,
		model.Category, model.Description, model.Capabilities, model.ParamsSchema,
		model.SystemPrompt, model.Pricing, model.MaxTokens, model.SupportStreaming,
		model.IsActive, model.IsFeatured, model.Weight,
	)
	
	return err
}
