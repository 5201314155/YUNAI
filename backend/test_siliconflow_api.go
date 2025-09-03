package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	fmt.Println("🧪 测试SiliconFlow API连接")
	fmt.Println("===========================================")

	// SiliconFlow API配置
	apiURL := "https://api.siliconflow.cn/v1/models"
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	fmt.Printf("📡 API地址: %s\n", apiURL)
	fmt.Printf("🔑 API密钥: %s...\n", apiKey[:20])

	// 测试基本连接
	fmt.Println("\n🔍 测试基本API连接...")
	models, err := fetchModels(apiURL, apiKey, "")
	if err != nil {
		fmt.Printf("❌ API连接失败: %v\n", err)
		return
	}

	fmt.Printf("✅ API连接成功！获取到 %d 个模型\n", len(models))

	// 显示前10个模型
	fmt.Println("\n📋 前10个模型:")
	fmt.Println("-------------------------------------------")
	for i, model := range models {
		if i >= 10 {
			break
		}
		fmt.Printf("%d. %s (%s)\n", i+1, model.ID, model.Object)
	}

	// 测试不同类型的模型
	fmt.Println("\n🔍 测试不同类型模型获取:")
	fmt.Println("-------------------------------------------")
	
	types := []string{"text", "image", "audio", "video"}
	for _, modelType := range types {
		fmt.Printf("正在获取%s类型模型...", modelType)
		typeModels, err := fetchModels(apiURL, apiKey, modelType)
		if err != nil {
			fmt.Printf(" ❌ 失败: %v\n", err)
			continue
		}
		fmt.Printf(" ✅ 成功获取 %d 个模型\n", len(typeModels))
	}

	fmt.Println("\n🎉 SiliconFlow API测试完成！")
	fmt.Println("✅ API连接正常")
	fmt.Println("✅ 模型获取功能正常")
	fmt.Println("✅ 可以开始正式添加模型到数据库")
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
