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
	fmt.Println("🎵 测试YUNAI音色管理API")
	fmt.Println("===========================================")

	baseURL := "http://localhost:8080/api/v1"
	testUserID := "550e8400-e29b-41d4-a716-446655440000"

	// 等待API服务启动
	fmt.Println("⏳ 等待API服务启动...")
	time.Sleep(2 * time.Second)

	// 1. 测试获取可用音色列表
	fmt.Println("\n📋 测试获取可用音色列表...")
	testGetAvailableVoices(baseURL, testUserID)

	// 2. 测试音色克隆
	fmt.Println("\n🎤 测试音色克隆...")
	voiceID := testCloneVoice(baseURL, testUserID)

	// 3. 测试音色试听
	fmt.Println("\n🔊 测试音色试听...")
	testVoicePreview(baseURL, voiceID, testUserID)

	// 4. 测试切换音色公开状态
	fmt.Println("\n🔄 测试切换音色公开状态...")
	testToggleVoicePrivacy(baseURL, voiceID, testUserID)

	// 5. 测试角色音色设置
	fmt.Println("\n🎯 测试角色音色设置...")
	characterID := "832fbefb-ba9f-4412-b1b4-f5fc756142ee" // 使用之前创建的角色ID
	testUseVoiceForCharacter(baseURL, voiceID, characterID, testUserID)

	// 6. 测试获取角色音色
	fmt.Println("\n📊 测试获取角色音色...")
	testGetCharacterVoices(baseURL, characterID)

	fmt.Println("\n🎉 所有API测试完成！")
}

func testGetAvailableVoices(baseURL, userID string) {
	url := fmt.Sprintf("%s/voices/available/%s", baseURL, userID)
	
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 状态码: %d\n", resp.StatusCode)
	fmt.Printf("✅ 响应: %s\n", result["message"])
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		if voices, ok := data["voices"].([]interface{}); ok {
			fmt.Printf("✅ 找到 %d 个可用音色\n", len(voices))
			for i, voice := range voices {
				if v, ok := voice.(map[string]interface{}); ok {
					fmt.Printf("   %d. %s (%s)\n", i+1, v["voice_name"], v["voice_id"])
				}
			}
		}
	}
}

func testCloneVoice(baseURL, userID string) string {
	url := fmt.Sprintf("%s/voices/clone", baseURL)
	
	payload := map[string]interface{}{
		"voice_name":         "API测试音色",
		"voice_description":  "通过API创建的测试音色",
		"voice_image_url":    "/images/voices/api_test.jpg",
		"original_audio_url": "/audio/uploads/api_test.wav",
		"is_public":          true,
		"creator_user_id":    userID,
		"gender":             "female",
		"age_range":          "adult",
		"language":           "zh",
		"emotion_tags":       []string{"专业", "清晰", "标准"},
	}

	jsonData, _ := json.Marshal(payload)
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return ""
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return ""
	}

	fmt.Printf("✅ 状态码: %d\n", resp.StatusCode)
	fmt.Printf("✅ 响应: %s\n", result["message"])
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		if voiceID, ok := data["voice_id"].(string); ok {
			fmt.Printf("✅ 音色ID: %s\n", voiceID)
			return voiceID
		}
	}
	
	return ""
}

func testVoicePreview(baseURL, voiceID, userID string) {
	url := fmt.Sprintf("%s/voices/%s/preview", baseURL, voiceID)
	
	payload := map[string]interface{}{
		"user_id": userID,
		"text":    "这是一个音色试听测试，请听听我的声音如何。",
	}

	jsonData, _ := json.Marshal(payload)
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 状态码: %d\n", resp.StatusCode)
	fmt.Printf("✅ 响应: %s\n", result["message"])
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		if previewURL, ok := data["preview_url"].(string); ok {
			fmt.Printf("✅ 试听地址: %s\n", previewURL)
		}
	}
}

func testToggleVoicePrivacy(baseURL, voiceID, userID string) {
	url := fmt.Sprintf("%s/voices/%s/privacy", baseURL, voiceID)
	
	payload := map[string]interface{}{
		"user_id":   userID,
		"is_public": false, // 切换为私密
	}

	jsonData, _ := json.Marshal(payload)
	
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 状态码: %d\n", resp.StatusCode)
	fmt.Printf("✅ 响应: %s\n", result["message"])
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		if isPublic, ok := data["is_public"].(bool); ok {
			status := "私密"
			if isPublic {
				status = "公开"
			}
			fmt.Printf("✅ 音色状态: %s\n", status)
		}
	}
}

func testUseVoiceForCharacter(baseURL, voiceID, characterID, userID string) {
	url := fmt.Sprintf("%s/voices/%s/use", baseURL, voiceID)
	
	payload := map[string]interface{}{
		"user_id":      userID,
		"character_id": characterID,
		"is_primary":   true,
		"settings": map[string]interface{}{
			"speed":    1.2,
			"pitch":    1.0,
			"emotion":  "friendly",
			"volume":   0.9,
		},
	}

	jsonData, _ := json.Marshal(payload)
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 状态码: %d\n", resp.StatusCode)
	fmt.Printf("✅ 响应: %s\n", result["message"])
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		fmt.Printf("✅ 角色ID: %s\n", data["character_id"])
		fmt.Printf("✅ 音色ID: %s\n", data["voice_id"])
		fmt.Printf("✅ 主要音色: %t\n", data["is_primary"])
	}
}

func testGetCharacterVoices(baseURL, characterID string) {
	url := fmt.Sprintf("%s/characters/%s/voices", baseURL, characterID)
	
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 状态码: %d\n", resp.StatusCode)
	fmt.Printf("✅ 响应: %s\n", result["message"])
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		if voices, ok := data["voices"].([]interface{}); ok {
			fmt.Printf("✅ 角色共有 %d 个音色\n", len(voices))
			for i, voice := range voices {
				if v, ok := voice.(map[string]interface{}); ok {
					isPrimary := ""
					if primary, ok := v["is_primary"].(bool); ok && primary {
						isPrimary = " (主要音色)"
					}
					fmt.Printf("   %d. %s%s\n", i+1, v["voice_name"], isPrimary)
					fmt.Printf("      设置: %s\n", v["voice_settings"])
				}
			}
		}
	}
}
