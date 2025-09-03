package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🚀 YUNAI完整系统功能测试")
	fmt.Println("===============================================================================")

	baseURL := "http://localhost:8081"

	// 测试结果统计
	testResults := make(map[string]bool)

	// 1. 用户认证系统测试
	fmt.Println("\n1. 🔐 用户认证系统测试")
	userID, token := testUserAuth(baseURL)
	testResults["用户认证"] = userID != ""

	if userID == "" {
		fmt.Println("❌ 用户认证失败，无法继续测试")
		return
	}

	// 2. AI角色管理系统测试
	fmt.Println("\n2. 🤖 AI角色管理系统测试")
	characterID := testCharacterManagement(baseURL, token)
	testResults["角色管理"] = characterID != ""

	// 3. 用户身份识别系统测试
	fmt.Println("\n3. 👤 用户身份识别系统测试")
	testResults["身份识别"] = testUserIdentity(baseURL, token, userID, characterID)

	// 4. 全局欺诈提示词系统测试
	fmt.Println("\n4. 🧠 全局欺诈提示词系统测试")
	testResults["全局提示词"] = testGlobalPrompt(baseURL, token, userID, characterID)

	// 5. 朋友圈系统测试
	fmt.Println("\n5. 📱 朋友圈系统测试")
	testResults["朋友圈"] = testMoments(baseURL, token, userID, characterID)

	// 6. 钱包系统测试
	fmt.Println("\n6. 💰 钱包系统测试")
	testResults["钱包系统"] = testWallet(baseURL, token, userID)

	// 7. 世界观设定系统测试
	fmt.Println("\n7. 🌍 世界观设定系统测试")
	testResults["世界观设定"] = testWorldSetting(baseURL, token, userID)

	// 8. 语音通话系统测试
	fmt.Println("\n8. 🎙️ 语音通话系统测试")
	testResults["语音通话"] = testVoiceCall(baseURL, token, userID, characterID)

	// 输出测试结果
	fmt.Println("\n===============================================================================")
	fmt.Println("🎯 YUNAI系统功能测试结果")
	fmt.Println("===============================================================================")

	successCount := 0
	totalCount := len(testResults)

	for feature, success := range testResults {
		status := "❌ 失败"
		if success {
			status = "✅ 成功"
			successCount++
		}
		fmt.Printf("%-15s: %s\n", feature, status)
	}

	fmt.Printf("\n📊 测试总结: %d/%d 功能测试通过 (%.1f%%)\n",
		successCount, totalCount, float64(successCount)/float64(totalCount)*100)

	if successCount == totalCount {
		fmt.Println("🎉 所有功能测试通过！YUNAI系统运行正常！")
	} else {
		fmt.Println("⚠️ 部分功能测试失败，请检查相关模块")
	}
}

// 测试用户认证
func testUserAuth(baseURL string) (string, string) {
	timestamp := time.Now().Unix()
	username := fmt.Sprintf("testuser_%d", timestamp)
	email := fmt.Sprintf("test_%d@yunai.com", timestamp)

	registerData := map[string]interface{}{
		"username": username,
		"email":    email,
		"password": "test123456",
		"nickname": "测试用户小明",
	}

	resp, err := makeRequest("POST", baseURL+"/api/v1/auth/register", registerData, "")
	if err != nil {
		fmt.Printf("   ❌ 注册失败: %v\n", err)
		return "", ""
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if user, ok := result["user"].(map[string]interface{}); ok {
		userID := user["id"].(string)
		fmt.Printf("   ✅ 用户注册成功: %s\n", userID)

		// 测试登录
		loginData := map[string]interface{}{
			"email":    email,
			"password": "test123456",
		}

		_, err = makeRequest("POST", baseURL+"/api/v1/auth/login", loginData, "")
		if err != nil {
			fmt.Printf("   ❌ 登录失败: %v\n", err)
			return userID, ""
		}

		fmt.Printf("   ✅ 登录成功\n")
		return userID, "temp_token"
	}

	return "", ""
}

// 测试角色管理
func testCharacterManagement(baseURL, token string) string {
	characterData := map[string]interface{}{
		"name":        "小雨",
		"description": "温柔善良的图书管理员",
		"personality": map[string]interface{}{
			"description": "温柔、善良、文静",
		},
	}

	resp, err := makeRequest("POST", baseURL+"/api/v1/characters", characterData, token)
	if err != nil {
		fmt.Printf("   ❌ 创建角色失败: %v\n", err)
		return ""
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if character, ok := result["character"].(map[string]interface{}); ok {
		characterID := character["id"].(string)
		fmt.Printf("   ✅ 角色创建成功: %s\n", characterID)
		return characterID
	}

	return ""
}

// 测试用户身份识别
func testUserIdentity(baseURL, token, userID, characterID string) bool {
	extractData := map[string]interface{}{
		"user_id":      userID,
		"context_type": "character",
		"context_id":   characterID,
		"setting_text": "在这个魔法学院中，用户（小明）是一个勇敢的战士学生，你小雨是图书管理员。",
	}

	resp, err := makeRequest("POST", baseURL+"/api/v1/identity/extract", extractData, token)
	if err != nil {
		fmt.Printf("   ❌ 身份提取失败: %v\n", err)
		return false
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		if identityContext, ok := data["identity_context"].(map[string]interface{}); ok {
			displayName := identityContext["display_name"].(string)
			fmt.Printf("   ✅ 身份识别成功: %s\n", displayName)
			return displayName == "小明"
		}
	}

	return false
}

// 测试全局提示词
func testGlobalPrompt(baseURL, token, userID, characterID string) bool {
	promptData := map[string]interface{}{
		"user_id":       userID,
		"character_id":  characterID,
		"context_type":  "chat",
		"world_setting": "魔法学院，充满神秘和冒险",
		"relationship":  "朋友关系，互相信任",
	}

	resp, err := makeRequest("POST", baseURL+"/api/v1/global-prompt/generate", promptData, token)
	if err != nil {
		fmt.Printf("   ❌ 全局提示词生成失败: %v\n", err)
		return false
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		if globalPrompt, ok := data["global_prompt"].(string); ok {
			fmt.Printf("   ✅ 全局提示词生成成功 (长度: %d)\n", len(globalPrompt))
			return len(globalPrompt) > 50
		}
	}

	return false
}

// 测试朋友圈
func testMoments(baseURL, token, userID, characterID string) bool {
	momentData := map[string]interface{}{
		"user_id":      userID,
		"character_id": characterID,
		"content":      "今天在图书馆整理了很多新书，感觉很充实！",
		"content_type": "text",
		"mood":         "happy",
		"tags":         []string{"日常", "工作"},
		"visibility":   "friends",
	}

	resp, err := makeRequest("POST", baseURL+"/api/v1/moments", momentData, token)
	if err != nil {
		fmt.Printf("   ❌ 创建朋友圈失败: %v\n", err)
		return false
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	// 检查朋友圈创建响应格式
	if moment, ok := result["moment"].(map[string]interface{}); ok {
		momentID := moment["id"].(string)
		fmt.Printf("   ✅ 朋友圈创建成功: %s\n", momentID)
		return true
	} else if message, ok := result["message"].(string); ok && message == "朋友圈创建成功" {
		fmt.Printf("   ✅ 朋友圈创建成功\n")
		return true
	}

	fmt.Printf("   ❌ 朋友圈响应格式错误: %s\n", string(resp))
	return false
}

// 测试钱包系统
func testWallet(baseURL, token, userID string) bool {
	resp, err := makeRequest("GET", baseURL+"/api/v1/wallet/balance?user_id="+userID, nil, token)
	if err != nil {
		fmt.Printf("   ❌ 钱包查询失败: %v\n", err)
		return false
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		balance := data["balance"]
		fmt.Printf("   ✅ 钱包查询成功，余额: %v\n", balance)
		return true
	}

	return false
}

// 测试世界观设定
func testWorldSetting(baseURL, token, userID string) bool {
	worldData := map[string]interface{}{
		"user_id":         userID,
		"name":            "魔法学院世界",
		"description":     "充满魔法和冒险的学院世界",
		"setting_content": "这是一个魔法学院，学生们在这里学习各种魔法技能。",
	}

	_, err := makeRequest("POST", baseURL+"/api/v1/world-setting/create", worldData, token)
	if err != nil {
		fmt.Printf("   ❌ 创建世界观设定失败: %v\n", err)
		return false
	}

	fmt.Printf("   ✅ 世界观设定创建成功\n")
	return true
}

// 测试语音通话
func testVoiceCall(baseURL, token, userID, characterID string) bool {
	callData := map[string]interface{}{
		"user_id":      userID,
		"character_id": characterID,
		"call_type":    "voice",
	}

	_, err := makeRequest("POST", baseURL+"/api/v1/voice-call/start", callData, token)
	if err != nil {
		fmt.Printf("   ❌ 开始语音通话失败: %v\n", err)
		return false
	}

	fmt.Printf("   ✅ 语音通话开始成功\n")
	return true
}

// HTTP请求辅助函数
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
