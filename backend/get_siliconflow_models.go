package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Provider     string `json:"provider"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	MaxTokens    int    `json:"max_tokens"`
	InputPrice   string `json:"input_price"`
	OutputPrice  string `json:"output_price"`
	IsActive     bool   `json:"is_active"`
	SortOrder    int    `json:"sort_order"`
}

func main() {
	// SiliconFlow API配置
	apiURL := "https://api.siliconflow.cn/v1/models"
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	// 创建HTTP请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("API请求失败，状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
		return
	}

	// 解析响应
	var sfResponse SiliconFlowResponse
	if err := json.Unmarshal(body, &sfResponse); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return
	}

	fmt.Printf("🚀 成功获取到 %d 个SiliconFlow模型:\n\n", len(sfResponse.Data))

	// 转换为YUNAI模型格式
	yunaiModels := make([]YUNAIModel, 0, len(sfResponse.Data))
	
	for i, model := range sfResponse.Data {
		yunaiModel := convertToYUNAIModel(model, i+1)
		yunaiModels = append(yunaiModels, yunaiModel)
		
		fmt.Printf("📋 模型 %d:\n", i+1)
		fmt.Printf("   ID: %s\n", yunaiModel.ID)
		fmt.Printf("   显示名称: %s\n", yunaiModel.DisplayName)
		fmt.Printf("   分类: %s\n", yunaiModel.Category)
		fmt.Printf("   描述: %s\n", yunaiModel.Description)
		fmt.Printf("   最大Token: %d\n", yunaiModel.MaxTokens)
		fmt.Printf("   输入价格: %s\n", yunaiModel.InputPrice)
		fmt.Printf("   输出价格: %s\n", yunaiModel.OutputPrice)
		fmt.Printf("   ----------------------------------------\n")
	}

	// 生成SQL插入语句
	generateSQL(yunaiModels)
}

func convertToYUNAIModel(model Model, sortOrder int) YUNAIModel {
	// 根据模型ID判断分类和设置参数
	category := categorizeModel(model.ID)
	displayName := generateDisplayName(model.ID)
	description := generateDescription(model.ID, category)
	maxTokens := getMaxTokens(model.ID)
	inputPrice, outputPrice := getPricing(model.ID)

	return YUNAIModel{
		ID:          model.ID,
		Name:        model.ID,
		DisplayName: displayName,
		Provider:    "SiliconFlow",
		Category:    category,
		Description: description,
		MaxTokens:   maxTokens,
		InputPrice:  inputPrice,
		OutputPrice: outputPrice,
		IsActive:    true,
		SortOrder:   sortOrder,
	}
}

func categorizeModel(modelID string) string {
	modelID = strings.ToLower(modelID)
	
	if strings.Contains(modelID, "gpt") || strings.Contains(modelID, "chatgpt") {
		return "对话模型"
	} else if strings.Contains(modelID, "claude") {
		return "对话模型"
	} else if strings.Contains(modelID, "llama") {
		return "开源模型"
	} else if strings.Contains(modelID, "qwen") || strings.Contains(modelID, "通义") {
		return "中文模型"
	} else if strings.Contains(modelID, "baichuan") || strings.Contains(modelID, "百川") {
		return "中文模型"
	} else if strings.Contains(modelID, "chatglm") || strings.Contains(modelID, "glm") {
		return "中文模型"
	} else if strings.Contains(modelID, "deepseek") {
		return "代码模型"
	} else if strings.Contains(modelID, "code") || strings.Contains(modelID, "coder") {
		return "代码模型"
	} else if strings.Contains(modelID, "embedding") || strings.Contains(modelID, "embed") {
		return "嵌入模型"
	} else if strings.Contains(modelID, "vision") || strings.Contains(modelID, "image") {
		return "多模态模型"
	} else if strings.Contains(modelID, "instruct") || strings.Contains(modelID, "chat") {
		return "对话模型"
	} else {
		return "通用模型"
	}
}

func generateDisplayName(modelID string) string {
	// 生成友好的显示名称
	parts := strings.Split(modelID, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return modelID
}

func generateDescription(modelID string, category string) string {
	modelID = strings.ToLower(modelID)
	
	descriptions := map[string]string{
		"gpt":       "OpenAI GPT系列模型，具有强大的对话和推理能力",
		"claude":    "Anthropic Claude系列模型，注重安全性和有用性",
		"llama":     "Meta开源的大语言模型，性能优异且可商用",
		"qwen":      "阿里巴巴通义千问系列，中文理解能力突出",
		"baichuan":  "百川智能开源模型，专为中文优化",
		"chatglm":   "清华大学开源的中英双语模型",
		"deepseek":  "深度求索代码模型，专业的编程助手",
		"embedding": "文本嵌入模型，用于语义搜索和相似度计算",
		"vision":    "多模态模型，支持图像理解和生成",
	}
	
	for key, desc := range descriptions {
		if strings.Contains(modelID, key) {
			return desc
		}
	}
	
	return fmt.Sprintf("SiliconFlow提供的%s，适用于各种AI应用场景", category)
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
	} else {
		return 8192 // 默认值
	}
}

func getPricing(modelID string) (string, string) {
	// SiliconFlow的大概定价（需要根据实际情况调整）
	modelID = strings.ToLower(modelID)
	
	if strings.Contains(modelID, "gpt-4") {
		return "¥0.105/1K tokens", "¥0.315/1K tokens"
	} else if strings.Contains(modelID, "gpt-3.5") {
		return "¥0.0105/1K tokens", "¥0.021/1K tokens"
	} else if strings.Contains(modelID, "claude") {
		return "¥0.084/1K tokens", "¥0.252/1K tokens"
	} else if strings.Contains(modelID, "llama") {
		return "¥0.007/1K tokens", "¥0.007/1K tokens"
	} else {
		return "¥0.014/1K tokens", "¥0.014/1K tokens"
	}
}

func generateSQL(models []YUNAIModel) {
	fmt.Printf("\n🔧 生成SQL插入语句:\n\n")
	
	for _, model := range models {
		sql := fmt.Sprintf(`INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), '%s', '%s', '%s', '%s', '%s', %d, '%s', '%s', %t, %d, NOW(), NOW());`,
			strings.ReplaceAll(model.Name, "'", "''"),
			strings.ReplaceAll(model.DisplayName, "'", "''"),
			model.Provider,
			model.Category,
			strings.ReplaceAll(model.Description, "'", "''"),
			model.MaxTokens,
			model.InputPrice,
			model.OutputPrice,
			model.IsActive,
			model.SortOrder,
		)
		fmt.Println(sql)
	}
}
