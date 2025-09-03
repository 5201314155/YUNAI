package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const BASE_URL = "http://localhost:8080/api/v1"

func main() {
	fmt.Println("🧪 测试YUNAI模型API获取功能")
	fmt.Println("===========================================")

	// 等待API服务启动
	fmt.Println("⏳ 等待API服务启动...")
	time.Sleep(2 * time.Second)

	// 1. 测试系统信息API
	fmt.Println("📊 测试系统信息API...")
	testSystemInfo()

	// 2. 测试管理员模型API (显示真实模型名称)
	fmt.Println("\n🔧 测试管理员模型API (显示真实模型名称)...")
	testAdminModelsAPI()

	// 3. 测试用户模型API (显示友好名称)
	fmt.Println("\n👤 测试用户模型API (显示友好名称)...")
	testUserModelsAPI()

	// 4. 测试不同类型的模型
	fmt.Println("\n🎨 测试获取图像模型...")
	testModelsByType("image")

	fmt.Println("\n🧠 测试获取嵌入模型...")
	testModelsByType("embedding")

	fmt.Println("\n🎵 测试获取音频模型...")
	testModelsByType("audio")

	fmt.Println("\n🎬 测试获取视频模型...")
	testModelsByType("video")

	fmt.Println("\n🎉 模型API测试完成！")
}

// 测试系统信息
func testSystemInfo() {
	resp, err := http.Get(BASE_URL + "/system/info")
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ API请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("   响应: %s\n", string(body))
		return
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if system, ok := data["system"].(map[string]interface{}); ok {
			fmt.Printf("✅ 系统: %s v%s (%s)\n", system["name"], system["version"], system["status"])
		}
		if models, ok := data["models"].(map[string]interface{}); ok {
			fmt.Printf("✅ 模型统计: 总计%v个, 活跃%v个, 对话%v个, 流式%v个\n", 
				models["total"], models["active"], models["chat"], models["streaming"])
		}
	}
}

// 测试管理员模型API
func testAdminModelsAPI() {
	resp, err := http.Get(BASE_URL + "/admin/models?type=chat&limit=5")
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ API请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("   响应: %s\n", string(body))
		return
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if modelsData, ok := data["models"].([]interface{}); ok {
			fmt.Printf("✅ 管理员视图: 找到 %d 个对话模型\n", len(modelsData))
			
			for i, modelData := range modelsData {
				if i >= 5 { // 只显示前5个
					break
				}
				
				if modelMap, ok := modelData.(map[string]interface{}); ok {
					internalKey := getString(modelMap, "internal_key")
					customName := getString(modelMap, "custom_display_name")
					provider := getString(modelMap, "provider")
					weight := getFloat(modelMap, "weight")
					
					streamIcon := "❌"
					if getBool(modelMap, "support_streaming") {
						streamIcon = "🌊"
					}
					
					fmt.Printf("   %d. %s %s [%s] (权重: %.0f)\n", i+1, streamIcon, customName, provider, weight)
					fmt.Printf("      🔑 真实模型ID: %s\n", internalKey)
				}
			}
		}
		
		if total, ok := data["total"].(float64); ok {
			fmt.Printf("📊 总计: %.0f 个模型\n", total)
		}
	}
}

// 测试用户模型API
func testUserModelsAPI() {
	resp, err := http.Get(BASE_URL + "/models?type=chat")
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ API请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("   响应: %s\n", string(body))
		return
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if modelsData, ok := data["models"].([]interface{}); ok {
			fmt.Printf("✅ 用户视图: 找到 %d 个对话模型\n", len(modelsData))
			
			for i, modelData := range modelsData {
				if i >= 5 { // 只显示前5个
					break
				}
				
				if modelMap, ok := modelData.(map[string]interface{}); ok {
					name := getString(modelMap, "name")
					provider := getString(modelMap, "provider")
					category := getString(modelMap, "category")
					maxTokens := getFloat(modelMap, "max_tokens")
					
					streamIcon := "❌"
					if getBool(modelMap, "support_streaming") {
						streamIcon = "🌊"
					}
					
					featuredIcon := ""
					if getBool(modelMap, "is_featured") {
						featuredIcon = "⭐"
					}
					
					fmt.Printf("   %d. %s %s %s [%s]\n", i+1, streamIcon, featuredIcon, name, provider)
					fmt.Printf("      📂 分类: %s | 🔢 最大Token: %.0f\n", category, maxTokens)
					fmt.Printf("      🔒 真实模型ID对用户隐藏\n")
				}
			}
		}
		
		if total, ok := data["total"].(float64); ok {
			fmt.Printf("📊 总计: %.0f 个模型\n", total)
		}
	}
}

// 测试按类型获取模型
func testModelsByType(modelType string) {
	url := fmt.Sprintf("%s/models?type=%s", BASE_URL, modelType)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ API请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("   响应: %s\n", string(body))
		return
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if modelsData, ok := data["models"].([]interface{}); ok {
			fmt.Printf("✅ %s模型: 找到 %d 个\n", getTypeIcon(modelType), len(modelsData))
			
			for i, modelData := range modelsData {
				if i >= 3 { // 只显示前3个
					fmt.Printf("   ... 还有 %d 个%s模型\n", len(modelsData)-3, modelType)
					break
				}
				
				if modelMap, ok := modelData.(map[string]interface{}); ok {
					name := getString(modelMap, "name")
					provider := getString(modelMap, "provider")
					category := getString(modelMap, "category")
					
					fmt.Printf("   %d. %s [%s] - %s\n", i+1, name, provider, category)
				}
			}
		}
		
		if total, ok := data["total"].(float64); ok {
			fmt.Printf("📊 %s模型总计: %.0f 个\n", modelType, total)
		}
	}
}

// 获取类型图标
func getTypeIcon(modelType string) string {
	switch modelType {
	case "chat":
		return "🤖"
	case "embedding":
		return "🧠"
	case "image":
		return "🎨"
	case "audio":
		return "🎵"
	case "video":
		return "🎬"
	default:
		return "❓"
	}
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}
