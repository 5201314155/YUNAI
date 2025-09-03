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
	fmt.Println("🔧 调试全局提示词系统")
	
	baseURL := "http://localhost:8081"
	
	// 使用已知存在的角色ID
	characterID := "353110dc-28f3-4285-b4f4-91a9e787e632"
	userID := "35c4c7db-9e4e-4e94-bddd-3f025d1705cc"
	
	fmt.Printf("测试角色ID: %s\n", characterID)
	fmt.Printf("测试用户ID: %s\n", userID)
	
	// 测试全局提示词生成
	promptData := map[string]interface{}{
		"user_id":      userID,
		"character_id": characterID,
		"context_type": "chat",
		"world_setting": "魔法学院，充满神秘和冒险",
		"relationship":  "朋友关系，互相信任",
	}

	fmt.Println("发送请求到: /api/v1/global-prompt/generate")
	fmt.Printf("请求数据: %+v\n", promptData)

	resp, err := makeRequest("POST", baseURL+"/api/v1/global-prompt/generate", promptData, "temp_token")
	if err != nil {
		fmt.Printf("❌ 全局提示词生成失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 响应: %s\n", string(resp))
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
