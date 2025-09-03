package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func main() {
	fmt.Println("🔍 获取真实的SiliconFlow模型名称")
	fmt.Println("===========================================")

	// SiliconFlow API配置
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	// 获取真实的模型列表
	fmt.Println("📡 正在获取SiliconFlow真实模型列表...")
	realModels, err := getRealModels(apiKey)
	if err != nil {
		log.Fatal("获取模型失败:", err)
	}

	fmt.Printf("✅ 成功获取到 %d 个真实模型\n\n", len(realModels))

	// 按类型分类显示
	chatModels := []Model{}
	embeddingModels := []Model{}
	imageModels := []Model{}
	audioModels := []Model{}
	videoModels := []Model{}
	otherModels := []Model{}

	for _, model := range realModels {
		modelID := strings.ToLower(model.ID)
		
		if strings.Contains(modelID, "embedding") || strings.Contains(modelID, "bge") {
			embeddingModels = append(embeddingModels, model)
		} else if strings.Contains(modelID, "flux") || strings.Contains(modelID, "stable") || strings.Contains(modelID, "kolors") {
			imageModels = append(imageModels, model)
		} else if strings.Contains(modelID, "fish") || strings.Contains(modelID, "cosyvoice") || strings.Contains(modelID, "sensevoice") {
			audioModels = append(audioModels, model)
		} else if strings.Contains(modelID, "wan") || strings.Contains(modelID, "video") {
			videoModels = append(videoModels, model)
		} else if strings.Contains(modelID, "chat") || strings.Contains(modelID, "instruct") || 
				  strings.Contains(modelID, "qwen") || strings.Contains(modelID, "glm") || 
				  strings.Contains(modelID, "deepseek") || strings.Contains(modelID, "llama") {
			chatModels = append(chatModels, model)
		} else {
			otherModels = append(otherModels, model)
		}
	}

	// 显示对话模型 (重点)
	fmt.Printf("🤖 对话模型 (%d个):\n", len(chatModels))
	fmt.Println("-------------------------------------------")
	for i, model := range chatModels {
		fmt.Printf("%d. %s\n", i+1, model.ID)
		if i >= 9 { // 只显示前10个
			fmt.Printf("   ... 还有 %d 个对话模型\n", len(chatModels)-10)
			break
		}
	}

	// 显示其他类型模型
	fmt.Printf("\n🧠 嵌入模型 (%d个):\n", len(embeddingModels))
	for i, model := range embeddingModels {
		if i < 3 {
			fmt.Printf("   %s\n", model.ID)
		}
	}
	if len(embeddingModels) > 3 {
		fmt.Printf("   ... 还有 %d 个\n", len(embeddingModels)-3)
	}

	fmt.Printf("\n🎨 图像模型 (%d个):\n", len(imageModels))
	for i, model := range imageModels {
		if i < 3 {
			fmt.Printf("   %s\n", model.ID)
		}
	}
	if len(imageModels) > 3 {
		fmt.Printf("   ... 还有 %d 个\n", len(imageModels)-3)
	}

	fmt.Printf("\n🎵 音频模型 (%d个):\n", len(audioModels))
	for _, model := range audioModels {
		fmt.Printf("   %s\n", model.ID)
	}

	fmt.Printf("\n🎬 视频模型 (%d个):\n", len(videoModels))
	for _, model := range videoModels {
		fmt.Printf("   %s\n", model.ID)
	}

	fmt.Printf("\n❓ 其他模型 (%d个):\n", len(otherModels))
	for i, model := range otherModels {
		if i < 5 {
			fmt.Printf("   %s\n", model.ID)
		}
	}
	if len(otherModels) > 5 {
		fmt.Printf("   ... 还有 %d 个\n", len(otherModels)-5)
	}

	// 推荐用于测试的对话模型
	fmt.Println("\n🎯 推荐用于测试的对话模型:")
	fmt.Println("-------------------------------------------")
	
	recommendedModels := []string{}
	for _, model := range chatModels {
		modelID := strings.ToLower(model.ID)
		if strings.Contains(modelID, "qwen") || strings.Contains(modelID, "glm") || 
		   strings.Contains(modelID, "deepseek") || strings.Contains(modelID, "llama") {
			recommendedModels = append(recommendedModels, model.ID)
			if len(recommendedModels) >= 5 {
				break
			}
		}
	}

	for i, modelID := range recommendedModels {
		fmt.Printf("%d. %s\n", i+1, modelID)
	}

	fmt.Println("\n✅ 请使用上述真实模型ID进行对话测试")
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
