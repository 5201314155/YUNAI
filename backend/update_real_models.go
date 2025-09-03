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
	fmt.Println("🔄 更新真实的SiliconFlow模型列表")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// SiliconFlow API配置
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	// 1. 获取真实的模型列表
	fmt.Println("📡 正在获取SiliconFlow真实模型列表...")
	realModels, err := getRealModels(apiKey)
	if err != nil {
		log.Fatal("获取模型失败:", err)
	}

	fmt.Printf("✅ 成功获取到 %d 个真实模型\n", len(realModels))

	// 2. 清理旧的SiliconFlow模型
	fmt.Println("\n🗑️ 清理旧的模型数据...")
	_, err = db.Exec("DELETE FROM ai_models WHERE provider = 'siliconflow'")
	if err != nil {
		log.Printf("清理旧模型失败: %v", err)
	}

	// 3. 插入真实模型到数据库
	fmt.Println("\n💾 插入真实模型到数据库...")
	insertedCount := 0
	chatModelCount := 0

	for i, model := range realModels {
		// 判断模型类型
		modelType := getModelType(model.ID)
		category := categorizeModel(model.ID)
		displayName := generateDisplayName(model.ID)
		customDisplayName := generateCustomDisplayName(model.ID) // 管理端可自定义的用户友好名称
		
		// 判断是否支持流式
		supportStreaming := modelType == "chat"
		
		yunaiModel := YUNAIModel{
			InternalKey:       model.ID, // 使用真实的模型ID
			DisplayName:       displayName,
			CustomDisplayName: customDisplayName, // 用户端显示的名称
			Provider:          "siliconflow",
			ModelType:         modelType,
			Category:          category,
			Description:       fmt.Sprintf("SiliconFlow提供的%s，来自%s", displayName, category),
			Capabilities:      generateCapabilities(modelType),
			ParamsSchema:      generateParamsSchema(modelType),
			SystemPrompt:      generateSystemPrompt(modelType, model.ID),
			Pricing:           generatePricing(modelType),
			MaxTokens:         getMaxTokens(model.ID),
			SupportStreaming:  supportStreaming,
			IsActive:          true,
			IsFeatured:        i < 20, // 前20个设为推荐
			Weight:            1000 - i,
		}
		
		err := insertModel(db, yunaiModel)
		if err != nil {
			log.Printf("插入模型 %s 失败: %v", model.ID, err)
			continue
		}
		
		insertedCount++
		if modelType == "chat" {
			chatModelCount++
		}
		
		if insertedCount%10 == 0 || insertedCount <= 5 {
			fmt.Printf("✅ [%d/%d] 已添加: %s -> %s\n", insertedCount, len(realModels), model.ID, customDisplayName)
		}
	}

	fmt.Printf("\n🎉 成功添加 %d 个真实SiliconFlow模型！\n", insertedCount)
	fmt.Printf("   其中对话模型: %d 个 (支持流式输出)\n", chatModelCount)

	// 4. 显示对话模型列表用于测试
	fmt.Println("\n🤖 可用于测试的对话模型:")
	showChatModels(db)
}

func getRealModels(apiKey string) ([]Model, error) {
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

func generateDisplayName(modelID string) string {
	// 保持原始模型ID作为显示名称（管理端用）
	return modelID
}

func generateCustomDisplayName(modelID string) string {
	// 生成用户友好的显示名称（用户端显示）
	modelID = strings.ToLower(modelID)
	
	// 特殊模型名称映射
	nameMap := map[string]string{
		"deepseek-chat":           "DeepSeek 对话模型",
		"deepseek-coder":          "DeepSeek 代码助手",
		"qwen/qwen2.5-72b-instruct": "通义千问 2.5 (72B)",
		"qwen/qwen2-7b-instruct":    "通义千问 2.0 (7B)",
		"zhipuai/glm-4-9b-chat":     "智谱GLM-4 (9B)",
		"meta-llama/llama-3.1-8b-instruct": "Llama 3.1 (8B)",
		"meta-llama/llama-3.1-70b-instruct": "Llama 3.1 (70B)",
		"internlm/internlm2_5-7b-chat": "书生浦语 2.5 (7B)",
		"01-ai/yi-1.5-9b-chat":      "零一万物 Yi 1.5 (9B)",
	}
	
	// 检查是否有预定义的友好名称
	for key, friendlyName := range nameMap {
		if strings.Contains(modelID, strings.ToLower(key)) {
			return friendlyName
		}
	}
	
	// 如果没有预定义，生成通用友好名称
	parts := strings.Split(modelID, "/")
	if len(parts) > 1 {
		name := parts[len(parts)-1]
		name = strings.ReplaceAll(name, "-", " ")
		name = strings.ReplaceAll(name, "_", " ")
		name = strings.Title(name)
		return name
	}
	
	return strings.Title(strings.ReplaceAll(modelID, "-", " "))
}

func generateCapabilities(modelType string) string {
	switch modelType {
	case "chat":
		return `["chat", "text_generation", "conversation", "reasoning", "code_generation"]`
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

func generateSystemPrompt(modelType, modelID string) string {
	if modelType == "chat" {
		if strings.Contains(strings.ToLower(modelID), "deepseek") {
			return "你是DeepSeek AI助手，擅长逻辑推理和代码生成。请提供准确、有用的回答。"
		} else if strings.Contains(strings.ToLower(modelID), "qwen") {
			return "你是通义千问，阿里巴巴开发的AI助手。请用中文回答，提供准确、有帮助的信息。"
		} else if strings.Contains(strings.ToLower(modelID), "glm") {
			return "你是GLM AI助手，由清华大学开发。请提供专业、准确的回答。"
		}
		return "你是一个有用的AI助手，请提供准确、有帮助的回答。"
	}
	return fmt.Sprintf("这是一个%s模型，专门用于相关任务处理。", modelType)
}

func generatePricing(modelType string) string {
	switch modelType {
	case "chat":
		return `{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}`
	case "embedding":
		return `{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}`
	case "image":
		return `{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}`
	case "audio":
		return `{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}`
	case "video":
		return `{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}`
	default:
		return `{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}`
	}
}

func getMaxTokens(modelID string) int {
	modelID = strings.ToLower(modelID)
	
	if strings.Contains(modelID, "72b") || strings.Contains(modelID, "70b") {
		return 32768
	} else if strings.Contains(modelID, "kimi") {
		return 200000
	} else if strings.Contains(modelID, "longcontext") {
		return 128000
	}
	
	return 8192
}

type YUNAIModel struct {
	InternalKey       string
	DisplayName       string
	CustomDisplayName string
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

func insertModel(db *sql.DB, model YUNAIModel) error {
	query := `
		INSERT INTO ai_models (
			internal_key, display_name, custom_display_name, provider, model_type, category, description,
			capabilities, params_schema, model_system_prompt, pricing, max_tokens,
			support_streaming, is_active, is_featured, weight, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
		)
	`
	
	_, err := db.Exec(query,
		model.InternalKey, model.DisplayName, model.CustomDisplayName, model.Provider, model.ModelType,
		model.Category, model.Description, model.Capabilities, model.ParamsSchema,
		model.SystemPrompt, model.Pricing, model.MaxTokens, model.SupportStreaming,
		model.IsActive, model.IsFeatured, model.Weight,
	)
	
	return err
}

func showChatModels(db *sql.DB) {
	query := `
		SELECT internal_key, custom_display_name, support_streaming, max_tokens
		FROM ai_models 
		WHERE provider = 'siliconflow' 
		AND model_type = 'chat'
		AND is_active = true
		ORDER BY weight DESC
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询对话模型失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("-------------------------------------------\n")
	count := 0
	for rows.Next() {
		var internalKey, customDisplayName string
		var supportStreaming bool
		var maxTokens int
		
		err := rows.Scan(&internalKey, &customDisplayName, &supportStreaming, &maxTokens)
		if err != nil {
			continue
		}
		
		count++
		streamIcon := "❌"
		if supportStreaming {
			streamIcon = "✅"
		}
		
		fmt.Printf("%d. %s\n", count, customDisplayName)
		fmt.Printf("   真实模型ID: %s\n", internalKey)
		fmt.Printf("   支持流式: %s | 最大Token: %d\n", streamIcon, maxTokens)
		fmt.Println()
	}
	
	fmt.Printf("✅ 共有 %d 个对话模型可用于测试\n", count)
}
