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

// SiliconFlow API响应结构 (根据官方文档)
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
	fmt.Println("🚀 YUNAI SiliconFlow模型管理系统")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// SiliconFlow API配置
	apiURL := "https://api.siliconflow.cn/v1/models"
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	fmt.Println("📡 正在获取SiliconFlow模型列表...")

	// 获取所有类型的模型
	modelTypes := []string{"", "text", "image", "audio", "video"}
	allModels := make([]Model, 0)

	for _, modelType := range modelTypes {
		models, err := fetchModels(apiURL, apiKey, modelType)
		if err != nil {
			log.Printf("获取%s类型模型失败: %v", modelType, err)
			continue
		}
		allModels = append(allModels, models...)
		fmt.Printf("✅ 获取%s类型模型: %d个\n", getTypeDisplayName(modelType), len(models))
	}

	// 去重
	uniqueModels := removeDuplicates(allModels)
	fmt.Printf("\n📊 总计获取到 %d 个唯一模型\n", len(uniqueModels))

	// 转换并插入数据库
	fmt.Println("\n💾 正在添加模型到数据库...")
	insertedCount := 0
	for i, model := range uniqueModels {
		yunaiModel := convertToYUNAIModel(model, i+1)
		
		err := insertOrUpdateModel(db, yunaiModel)
		if err != nil {
			log.Printf("插入模型 %s 失败: %v", yunaiModel.InternalKey, err)
			continue
		}
		
		insertedCount++
		fmt.Printf("✅ [%d/%d] %s - %s\n", insertedCount, len(uniqueModels), yunaiModel.InternalKey, yunaiModel.DisplayName)
	}

	fmt.Printf("\n🎉 成功添加/更新 %d 个SiliconFlow模型到数据库！\n", insertedCount)
	
	// 显示分类统计
	showCategoryStats(db)

	// 显示功能特性
	showFeatures()
}

func fetchModels(apiURL, apiKey, modelType string) ([]Model, error) {
	url := apiURL
	if modelType != "" {
		url += "?type=" + modelType
	}

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

func removeDuplicates(models []Model) []Model {
	seen := make(map[string]bool)
	unique := make([]Model, 0)
	
	for _, model := range models {
		if !seen[model.ID] {
			seen[model.ID] = true
			unique = append(unique, model)
		}
	}
	
	return unique
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
		IsFeatured:        sortOrder <= 20, // 前20个设为推荐
		Weight:            120 - sortOrder,  // 权重递减
	}
}

func categorizeModel(modelID string) string {
	modelID = strings.ToLower(modelID)
	
	// 根据模型ID进行智能分类
	if strings.Contains(modelID, "gpt") {
		return "OpenAI系列"
	} else if strings.Contains(modelID, "claude") {
		return "Anthropic系列"
	} else if strings.Contains(modelID, "qwen") {
		return "阿里通义"
	} else if strings.Contains(modelID, "glm") || strings.Contains(modelID, "chatglm") {
		return "清华GLM"
	} else if strings.Contains(modelID, "deepseek") {
		return "深度求索"
	} else if strings.Contains(modelID, "llama") {
		return "Meta开源"
	} else if strings.Contains(modelID, "baichuan") {
		return "百川智能"
	} else if strings.Contains(modelID, "embedding") || strings.Contains(modelID, "bge") {
		return "嵌入模型"
	} else if strings.Contains(modelID, "flux") || strings.Contains(modelID, "stable") || strings.Contains(modelID, "kolors") {
		return "图像生成"
	} else if strings.Contains(modelID, "fish") || strings.Contains(modelID, "cosyvoice") || strings.Contains(modelID, "sensevoice") {
		return "语音模型"
	} else if strings.Contains(modelID, "wan") || strings.Contains(modelID, "video") {
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
	// 生成友好的显示名称
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
	// 根据模型ID和分类生成描述
	return fmt.Sprintf("SiliconFlow提供的%s，来自%s分类，适用于相关AI应用场景", generateDisplayName(modelID), category)
}

func generateCapabilities(modelID string, modelType string) string {
	// 根据模型类型生成能力列表
	switch modelType {
	case "chat":
		return `["chat", "text_generation", "conversation", "reasoning"]`
	case "embedding":
		return `["embedding", "semantic_search", "similarity", "vector_retrieval"]`
	case "image":
		return `["image_generation", "text_to_image", "creative_art", "design"]`
	case "audio":
		return `["audio_processing", "speech_synthesis", "voice_generation"]`
	case "video":
		return `["video_generation", "text_to_video", "animation"]`
	case "reranking":
		return `["reranking", "search_optimization", "relevance_scoring"]`
	default:
		return `["general", "ai_assistant"]`
	}
}

func generateParamsSchema(modelType string) string {
	// 根据模型类型生成参数模式
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

func generateSystemPrompt(modelID string, modelType string) string {
	if modelType == "chat" {
		return "你是一个有用的AI助手，请提供准确、有帮助的回答。"
	}
	return fmt.Sprintf("这是一个%s模型，专门用于相关任务处理。", modelType)
}

func generatePricing(modelID string) string {
	// 根据模型复杂度设置价格
	return `{"input_token_price": 0.014, "output_token_price": 0.014, "unit": "1k_tokens", "currency": "CNY"}`
}

func getMaxTokens(modelID string) int {
	// 根据模型ID推断最大token数
	if strings.Contains(strings.ToLower(modelID), "32k") {
		return 32768
	} else if strings.Contains(strings.ToLower(modelID), "128k") {
		return 131072
	} else {
		return 8192
	}
}

func getSupportStreaming(modelType string) bool {
	return modelType == "chat"
}

func generateInternalKey(modelID string) string {
	// 生成内部键名
	key := strings.ReplaceAll(modelID, "/", "_")
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ToLower(key)
	return "sf_" + key
}

func insertOrUpdateModel(db *sql.DB, model YUNAIModel) error {
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

func showFeatures() {
	fmt.Println("\n🎯 YUNAI模型管理系统功能特性:")
	fmt.Println("-------------------------------------------")
	fmt.Println("✅ 模型分类管理 - 智能分类，便于管理")
	fmt.Println("✅ 模型禁用功能 - 可随时启用/禁用模型")
	fmt.Println("✅ 模型微调支持 - 支持自定义参数调整")
	fmt.Println("✅ 自定义提示词 - 为每个模型设置专用提示词")
	fmt.Println("✅ 流式开关控制 - 支持流式/非流式输出切换")
	fmt.Println("✅ 模型测试功能 - 一键测试模型可用性")
	fmt.Println("✅ 智能扣费系统 - 精确计算使用成本")
	fmt.Println("✅ 前端自由切换 - 用户可自由选择模型")
	fmt.Println("✅ 性能监控 - 实时监控模型响应时间和成功率")
	fmt.Println("✅ 备用模型机制 - 主模型失败时自动切换备用模型")
}

func getTypeDisplayName(modelType string) string {
	switch modelType {
	case "":
		return "全部"
	case "text":
		return "文本"
	case "image":
		return "图像"
	case "audio":
		return "语音"
	case "video":
		return "视频"
	default:
		return modelType
	}
}
