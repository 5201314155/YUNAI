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

// 用户信息API响应结构
type UserInfoResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Status  bool     `json:"status"`
	Data    UserData `json:"data"`
}

type UserData struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Image         string `json:"image"`
	Email         string `json:"email"`
	IsAdmin       bool   `json:"isAdmin"`
	Balance       string `json:"balance"`        // 可用余额
	Status        string `json:"status"`
	Introduction  string `json:"introduction"`
	Role          string `json:"role"`
	ChargeBalance string `json:"chargeBalance"` // 充值余额
	TotalBalance  string `json:"totalBalance"`  // 总余额
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
	TestStatus        string `json:"test_status"`        // 测试状态
	LastTestTime      string `json:"last_test_time"`     // 最后测试时间
	ResponseTime      int    `json:"response_time"`      // 响应时间(ms)
	SuccessRate       float64 `json:"success_rate"`      // 成功率
	TotalCost         float64 `json:"total_cost"`        // 总消费
	UsageCount        int64   `json:"usage_count"`       // 使用次数
}

func main() {
	fmt.Println("🚀 YUNAI完整SiliconFlow管理系统")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// SiliconFlow API配置
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	// 1. 获取用户信息
	fmt.Println("👤 获取SiliconFlow用户信息...")
	userInfo, err := getUserInfo(apiKey)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
	} else {
		displayUserInfo(userInfo)
	}

	// 2. 获取所有类型的模型
	fmt.Println("\n📡 获取SiliconFlow所有模型...")
	allModels := make([]Model, 0)
	
	// 支持的模型类型
	modelTypes := []string{"", "text", "image", "audio", "video"}
	subTypes := []string{"", "chat", "embedding", "reranker", "text-to-image", "image-to-image", "speech-to-text", "text-to-video"}
	
	for _, modelType := range modelTypes {
		for _, subType := range subTypes {
			if modelType == "" && subType == "" {
				continue // 避免重复获取
			}
			
			models, err := getModels(apiKey, modelType, subType)
			if err != nil {
				log.Printf("获取模型失败 (type:%s, subtype:%s): %v", modelType, subType, err)
				continue
			}
			
			if len(models) > 0 {
				allModels = append(allModels, models...)
				fmt.Printf("✅ %s/%s: %d个模型\n", getDisplayName(modelType), getDisplayName(subType), len(models))
			}
		}
	}

	// 去重
	uniqueModels := removeDuplicates(allModels)
	fmt.Printf("\n📊 总计获取到 %d 个唯一模型\n", len(uniqueModels))

	// 3. 添加模型到数据库
	fmt.Println("\n💾 添加模型到数据库...")
	insertedCount := 0
	for i, model := range uniqueModels {
		yunaiModel := convertToYUNAIModel(model, i+1)
		
		err := insertOrUpdateModel(db, yunaiModel)
		if err != nil {
			log.Printf("插入模型 %s 失败: %v", yunaiModel.InternalKey, err)
			continue
		}
		
		insertedCount++
		if insertedCount <= 10 || insertedCount%10 == 0 {
			fmt.Printf("✅ [%d/%d] %s - %s\n", insertedCount, len(uniqueModels), yunaiModel.InternalKey, yunaiModel.DisplayName)
		}
	}

	fmt.Printf("\n🎉 成功添加/更新 %d 个SiliconFlow模型！\n", insertedCount)
	
	// 4. 显示统计信息
	showDetailedStats(db)
	
	// 5. 显示管理功能
	showManagementFeatures()
}

func getUserInfo(apiKey string) (*UserData, error) {
	url := "https://api.siliconflow.cn/v1/user/info"
	
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

	var userResponse UserInfoResponse
	if err := json.Unmarshal(body, &userResponse); err != nil {
		return nil, err
	}

	if userResponse.Code != 20000 {
		return nil, fmt.Errorf("API返回错误: %s", userResponse.Message)
	}

	return &userResponse.Data, nil
}

func getModels(apiKey, modelType, subType string) ([]Model, error) {
	url := "https://api.siliconflow.cn/v1/models"
	params := make([]string, 0)
	
	if modelType != "" {
		params = append(params, "type="+modelType)
	}
	if subType != "" {
		params = append(params, "sub_type="+subType)
	}
	
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
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
		return nil, fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	var sfResponse SiliconFlowResponse
	if err := json.Unmarshal(body, &sfResponse); err != nil {
		return nil, err
	}

	return sfResponse.Data, nil
}

func displayUserInfo(userInfo *UserData) {
	fmt.Println("💰 SiliconFlow账户信息:")
	fmt.Println("-------------------------------------------")
	fmt.Printf("👤 用户ID: %s\n", userInfo.ID)
	fmt.Printf("📧 邮箱: %s\n", userInfo.Email)
	fmt.Printf("🏷️ 状态: %s\n", userInfo.Status)
	fmt.Printf("👑 管理员: %t\n", userInfo.IsAdmin)
	fmt.Printf("💰 可用余额: ¥%s\n", userInfo.Balance)
	fmt.Printf("💳 充值余额: ¥%s\n", userInfo.ChargeBalance)
	fmt.Printf("💎 总余额: ¥%s\n", userInfo.TotalBalance)
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
		IsFeatured:        sortOrder <= 20,
		Weight:            120 - sortOrder,
		TestStatus:        "未测试",
		LastTestTime:      "",
		ResponseTime:      0,
		SuccessRate:       0.0,
		TotalCost:         0.0,
		UsageCount:        0,
	}
}

// 智能分类函数
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

func generateDescription(modelID string, category string) string {
	return fmt.Sprintf("SiliconFlow提供的%s，来自%s分类，支持完整的API调用和参数配置", generateDisplayName(modelID), category)
}

func generateCapabilities(modelID string, modelType string) string {
	switch modelType {
	case "chat":
		return `["chat", "text_generation", "conversation", "reasoning", "role_play"]`
	case "embedding":
		return `["embedding", "semantic_search", "similarity", "vector_retrieval", "text_analysis"]`
	case "image":
		return `["image_generation", "text_to_image", "creative_art", "design", "visual_content"]`
	case "audio":
		return `["audio_processing", "speech_synthesis", "voice_generation", "tts", "asr"]`
	case "video":
		return `["video_generation", "text_to_video", "animation", "visual_storytelling"]`
	case "reranking":
		return `["reranking", "search_optimization", "relevance_scoring", "information_retrieval"]`
	default:
		return `["general", "ai_assistant", "multi_purpose"]`
	}
}

func generateParamsSchema(modelType string) string {
	switch modelType {
	case "chat":
		return `{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2, "description": "控制输出随机性"}, "max_tokens": {"type": "integer", "default": 2000, "min": 1, "max": 8000, "description": "最大输出长度"}, "stream": {"type": "boolean", "default": false, "description": "是否流式输出"}, "top_p": {"type": "number", "default": 0.9, "min": 0, "max": 1, "description": "核采样参数"}}`
	case "embedding":
		return `{"dimensions": {"type": "integer", "default": 1024, "description": "向量维度"}, "normalize": {"type": "boolean", "default": true, "description": "是否归一化"}}`
	case "image":
		return `{"width": {"type": "integer", "default": 1024, "min": 256, "max": 2048, "description": "图像宽度"}, "height": {"type": "integer", "default": 1024, "min": 256, "max": 2048, "description": "图像高度"}, "steps": {"type": "integer", "default": 20, "min": 1, "max": 100, "description": "生成步数"}}`
	case "audio":
		return `{"voice_id": {"type": "string", "default": "default", "description": "声音ID"}, "speed": {"type": "number", "default": 1.0, "min": 0.5, "max": 2.0, "description": "语速"}}`
	case "video":
		return `{"duration": {"type": "integer", "default": 5, "min": 1, "max": 30, "description": "视频时长"}, "fps": {"type": "integer", "default": 24, "min": 12, "max": 60, "description": "帧率"}}`
	default:
		return `{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2, "description": "控制参数"}}`
	}
}

func generateSystemPrompt(modelID string, modelType string) string {
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

func generatePricing(modelID string) string {
	return `{"input_token_price": 0.014, "output_token_price": 0.014, "unit": "1k_tokens", "currency": "CNY"}`
}

func getMaxTokens(modelID string) int {
	modelID = strings.ToLower(modelID)
	if strings.Contains(modelID, "128k") {
		return 131072
	} else if strings.Contains(modelID, "32k") {
		return 32768
	} else if strings.Contains(modelID, "16k") {
		return 16384
	} else {
		return 8192
	}
}

func getSupportStreaming(modelType string) bool {
	return modelType == "chat"
}

func generateInternalKey(modelID string) string {
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

func showDetailedStats(db *sql.DB) {
	fmt.Println("\n📊 详细模型统计:")
	fmt.Println("===========================================")
	
	// 按分类统计
	query := `
		SELECT category, model_type, COUNT(*) as count, 
		       COUNT(CASE WHEN is_active = true THEN 1 END) as active_count,
		       COUNT(CASE WHEN is_featured = true THEN 1 END) as featured_count
		FROM ai_models 
		WHERE provider = 'siliconflow'
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
	totalActive := 0
	totalFeatured := 0
	
	for rows.Next() {
		var category, modelType string
		var count, activeCount, featuredCount int
		
		if err := rows.Scan(&category, &modelType, &count, &activeCount, &featuredCount); err != nil {
			continue
		}
		
		if category != currentCategory {
			if currentCategory != "" {
				fmt.Println()
			}
			fmt.Printf("📁 %s:\n", category)
			currentCategory = category
		}
		
		fmt.Printf("   %s: %d个 (活跃:%d, 推荐:%d)\n", modelType, count, activeCount, featuredCount)
		totalCount += count
		totalActive += activeCount
		totalFeatured += featuredCount
	}
	
	fmt.Printf("\n🎯 总计: %d个模型 (活跃:%d, 推荐:%d)\n", totalCount, totalActive, totalFeatured)
}

func showManagementFeatures() {
	fmt.Println("\n🎛️ YUNAI SiliconFlow管理功能:")
	fmt.Println("===========================================")
	fmt.Println("✅ 用户信息管理 - 实时查看余额、额度、使用情况")
	fmt.Println("✅ 全类型模型支持 - 文本、图像、音频、视频、嵌入、重排序")
	fmt.Println("✅ 智能分类管理 - 按厂商、类型自动分类")
	fmt.Println("✅ 模型启用/禁用 - 灵活控制模型可用性")
	fmt.Println("✅ 自定义提示词 - 为每个模型设置专用提示词")
	fmt.Println("✅ 参数配置管理 - 完整的参数模式定义")
	fmt.Println("✅ 流式输出控制 - 支持流式/非流式切换")
	fmt.Println("✅ 模型测试功能 - 一键测试模型可用性和响应时间")
	fmt.Println("✅ 使用统计监控 - 实时监控使用次数和成功率")
	fmt.Println("✅ 成本计算系统 - 精确计算和统计使用成本")
	fmt.Println("✅ 前端自由切换 - 用户可在前端自由选择模型")
	fmt.Println("✅ 备用模型机制 - 主模型失败时自动切换")
	fmt.Println("✅ 批量管理操作 - 支持批量启用/禁用/测试")
	fmt.Println("✅ 实时同步更新 - 定期同步SiliconFlow最新模型")
}

func getDisplayName(name string) string {
	if name == "" {
		return "全部"
	}
	displayNames := map[string]string{
		"text":            "文本",
		"image":           "图像",
		"audio":           "音频",
		"video":           "视频",
		"chat":            "对话",
		"embedding":       "嵌入",
		"reranker":        "重排序",
		"text-to-image":   "文本转图像",
		"image-to-image":  "图像转图像",
		"speech-to-text":  "语音转文本",
		"text-to-video":   "文本转视频",
	}
	
	if display, exists := displayNames[name]; exists {
		return display
	}
	return name
}
