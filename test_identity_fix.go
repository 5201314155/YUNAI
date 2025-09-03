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
	fmt.Println("🔧 测试身份识别系统修复")
	
	baseURL := "http://localhost:8081"
	
	// 1. 先注册一个用户
	fmt.Println("1. 注册测试用户...")
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
		fmt.Printf("❌ 注册失败: %v\n", err)
		return
	}

	var registerResult map[string]interface{}
	json.Unmarshal(resp, &registerResult)
	
	user := registerResult["user"].(map[string]interface{})
	userID := user["id"].(string)
	fmt.Printf("✅ 用户注册成功: %s\n", userID)

	// 2. 创建一个角色
	fmt.Println("2. 创建测试角色...")
	characterData := map[string]interface{}{
		"name":        "小雨",
		"description": "温柔的图书管理员",
		"personality": map[string]interface{}{
			"description": "温柔善良",
		},
	}

	resp, err = makeRequest("POST", baseURL+"/api/v1/characters", characterData, "temp_token")
	if err != nil {
		fmt.Printf("❌ 创建角色失败: %v\n", err)
		return
	}

	var characterResult map[string]interface{}
	json.Unmarshal(resp, &characterResult)
	
	character := characterResult["character"].(map[string]interface{})
	characterID := character["id"].(string)
	fmt.Printf("✅ 角色创建成功: %s\n", characterID)

	// 3. 测试身份识别
	fmt.Println("3. 测试身份识别...")
	extractData := map[string]interface{}{
		"user_id":      userID,
		"context_type": "character",
		"context_id":   characterID,
		"setting_text": "在这个魔法学院中，用户（小明）是一个勇敢的战士学生，你小雨是图书管理员。",
	}

	resp, err = makeRequest("POST", baseURL+"/api/v1/identity/extract", extractData, "temp_token")
	if err != nil {
		fmt.Printf("❌ 身份提取失败: %v\n", err)
		return
	}

	var identityResult map[string]interface{}
	json.Unmarshal(resp, &identityResult)
	
	fmt.Printf("✅ 身份识别响应: %s\n", string(resp))

	if data, ok := identityResult["data"].(map[string]interface{}); ok {
		if identityContext, ok := data["identity_context"].(map[string]interface{}); ok {
			displayName := identityContext["display_name"].(string)
			identityType := identityContext["identity_type"].(string)
			
			fmt.Printf("✅ 身份识别成功!\n")
			fmt.Printf("   提取的身份: %s\n", displayName)
			fmt.Printf("   身份类型: %s\n", identityType)
			
			if displayName == "小明" {
				fmt.Println("🎉 身份识别修复成功！")
			} else {
				fmt.Printf("⚠️ 身份提取不准确，期望'小明'，实际'%s'\n", displayName)
			}
		}
	}
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
