package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8081"

func main() {
	fmt.Println("🧪 YUNAI扩展功能系统测试")
	fmt.Println("===============================================================================")

	// 测试结果统计
	totalTests := 0
	passedTests := 0

	// 1. 测试AI模型管理系统
	fmt.Println("\n1. 🤖 测试AI模型管理系统...")
	if testModelManagement() {
		fmt.Println("   ✅ AI模型管理系统测试通过")
		passedTests++
	} else {
		fmt.Println("   ❌ AI模型管理系统测试失败")
	}
	totalTests++

	// 2. 测试AI调用计费系统
	fmt.Println("\n2. 💰 测试AI调用计费系统...")
	if testAIBilling() {
		fmt.Println("   ✅ AI调用计费系统测试通过")
		passedTests++
	} else {
		fmt.Println("   ❌ AI调用计费系统测试失败")
	}
	totalTests++

	// 3. 测试支付卡片系统
	fmt.Println("\n3. 💳 测试支付卡片系统...")
	if testPaymentCards() {
		fmt.Println("   ✅ 支付卡片系统测试通过")
		passedTests++
	} else {
		fmt.Println("   ❌ 支付卡片系统测试失败")
	}
	totalTests++

	// 输出测试结果
	fmt.Println("\n===============================================================================")
	fmt.Printf("📊 扩展功能测试结果: %d/%d 功能测试通过 (%.1f%%)\n", 
		passedTests, totalTests, float64(passedTests)/float64(totalTests)*100)
	fmt.Println("===============================================================================")

	if passedTests == totalTests {
		fmt.Println("🎉 所有扩展功能测试通过！")
	} else {
		fmt.Printf("⚠️ %d个功能需要修复\n", totalTests-passedTests)
	}
}

// testModelManagement 测试AI模型管理系统
func testModelManagement() bool {
	// 1. 创建AI模型
	modelData := map[string]interface{}{
		"internal_key":   "test-model-" + fmt.Sprintf("%d", time.Now().Unix()),
		"display_name":   "测试模型",
		"provider":       "deepseek",
		"model_type":     "chat",
		"capabilities":   json.RawMessage(`{"chat": true, "image": false}`),
		"params_schema":  json.RawMessage(`{"temperature": {"type": "float", "min": 0, "max": 2}}`),
		"pricing":        json.RawMessage(`{"input_token_price": 0.001, "output_token_price": 0.002}`),
		"visibility":     "public",
		"weight":         100,
	}

	resp, err := makeRequest("POST", baseURL+"/api/v1/models", modelData, "")
	if err != nil {
		fmt.Printf("   ❌ 创建模型失败: %v\n", err)
		return false
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if success, ok := result["success"].(bool); !ok || !success {
		fmt.Printf("   ❌ 创建模型响应错误: %s\n", string(resp))
		return false
	}

	modelData_result, ok := result["model"].(map[string]interface{})
	if !ok {
		fmt.Printf("   ❌ 模型数据格式错误\n")
		return false
	}

	modelID := modelData_result["id"].(string)
	fmt.Printf("   ✅ 模型创建成功: %s\n", modelID)

	// 2. 获取模型列表
	resp, err = makeRequest("GET", baseURL+"/api/v1/models", nil, "")
	if err != nil {
		fmt.Printf("   ❌ 获取模型列表失败: %v\n", err)
		return false
	}

	json.Unmarshal(resp, &result)
	if success, ok := result["success"].(bool); !ok || !success {
		fmt.Printf("   ❌ 获取模型列表响应错误\n")
		return false
	}

	fmt.Printf("   ✅ 模型列表获取成功\n")

	// 3. 测试模型连接
	resp, err = makeRequest("POST", baseURL+"/api/v1/models/"+modelID+"/test", nil, "")
	if err != nil {
		fmt.Printf("   ❌ 模型连接测试失败: %v\n", err)
		return false
	}

	json.Unmarshal(resp, &result)
	if success, ok := result["success"].(bool); !ok || !success {
		fmt.Printf("   ❌ 模型连接测试响应错误\n")
		return false
	}

	fmt.Printf("   ✅ 模型连接测试成功\n")
	return true
}

// testAIBilling 测试AI调用计费系统
func testAIBilling() bool {
	// 首先需要有用户和模型
	// 这里使用模拟的ID，实际应该从前面的测试中获取
	userID := "b35325dd-e28f-4111-a973-334cf4964efd" // 模拟用户ID
	modelID := "288813ef-6767-4737-b74d-1a103d0de299" // 模拟模型ID

	// 1. 测试AI调用计费
	chargeData := map[string]interface{}{
		"user_id":       userID,
		"model_id":      modelID,
		"session_id":    "test-session-" + fmt.Sprintf("%d", time.Now().Unix()),
		"call_type":     "chat",
		"input_tokens":  100,
		"output_tokens": 200,
	}

	resp, err := makeRequest("POST", baseURL+"/api/v1/billing/charge", chargeData, "")
	if err != nil {
		fmt.Printf("   ❌ AI调用计费失败: %v\n", err)
		return false
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if success, ok := result["success"].(bool); ok && success {
		fmt.Printf("   ✅ AI调用计费成功\n")
	} else {
		fmt.Printf("   ❌ AI调用计费响应错误: %s\n", string(resp))
		return false
	}

	// 2. 获取用户计费历史
	resp, err = makeRequest("GET", baseURL+"/api/v1/billing/history/"+userID, nil, "")
	if err != nil {
		fmt.Printf("   ❌ 获取计费历史失败: %v\n", err)
		return false
	}

	json.Unmarshal(resp, &result)
	if success, ok := result["success"].(bool); ok && success {
		fmt.Printf("   ✅ 计费历史获取成功\n")
		return true
	}

	fmt.Printf("   ❌ 计费历史响应错误\n")
	return false
}

// testPaymentCards 测试支付卡片系统
func testPaymentCards() bool {
	userID := "b35325dd-e28f-4111-a973-334cf4964efd" // 模拟用户ID

	// 1. 创建支付卡片
	cardData := map[string]interface{}{
		"user_id":    userID,
		"card_name":  "测试卡片",
		"card_type":  "standard",
		"balance":    100.0,
		"is_default": true,
	}

	resp, err := makeRequest("POST", baseURL+"/api/v1/payment-cards", cardData, "")
	if err != nil {
		fmt.Printf("   ❌ 创建支付卡片失败: %v\n", err)
		return false
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if success, ok := result["success"].(bool); !ok || !success {
		fmt.Printf("   ❌ 创建支付卡片响应错误: %s\n", string(resp))
		return false
	}

	cardData_result, ok := result["card"].(map[string]interface{})
	if !ok {
		fmt.Printf("   ❌ 卡片数据格式错误\n")
		return false
	}

	cardID := cardData_result["id"].(string)
	fmt.Printf("   ✅ 支付卡片创建成功: %s\n", cardID)

	// 2. 获取用户卡片列表
	resp, err = makeRequest("GET", baseURL+"/api/v1/payment-cards/user/"+userID, nil, "")
	if err != nil {
		fmt.Printf("   ❌ 获取卡片列表失败: %v\n", err)
		return false
	}

	json.Unmarshal(resp, &result)
	if success, ok := result["success"].(bool); !ok || !success {
		fmt.Printf("   ❌ 获取卡片列表响应错误\n")
		return false
	}

	fmt.Printf("   ✅ 卡片列表获取成功\n")

	// 3. 获取卡片类型
	resp, err = makeRequest("GET", baseURL+"/api/v1/payment-cards/types", nil, "")
	if err != nil {
		fmt.Printf("   ❌ 获取卡片类型失败: %v\n", err)
		return false
	}

	json.Unmarshal(resp, &result)
	if success, ok := result["success"].(bool); ok && success {
		fmt.Printf("   ✅ 卡片类型获取成功\n")
		return true
	}

	fmt.Printf("   ❌ 卡片类型响应错误\n")
	return false
}

// makeRequest 发送HTTP请求的辅助函数
func makeRequest(method, url string, data interface{}, token string) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
