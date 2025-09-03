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

func main() {
	fmt.Println("🔄 获取真实SiliconFlow模型列表并更新数据库")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// SiliconFlow API配置
	apiKey := "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"

	// 1. 获取真实的模型列表
	fmt.Println("📡 正在获取SiliconFlow真实模型列表...")
	realModels, err := getRealModels(apiKey)
	if err != nil {
		log.Fatal("获取模型失败:", err)
	}

	fmt.Printf("✅ 成功获取到 %d 个真实模型\n", len(realModels))

	// 2. 按类型分类
	chatModels := []Model{}
	embeddingModels := []Model{}
	imageModels := []Model{}
	audioModels := []Model{}
	videoModels := []Model{}

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
		} else {
			// 默认为对话模型
			chatModels = append(chatModels, model)
		}
	}

	fmt.Printf("📊 模型分类统计:\n")
	fmt.Printf("   🤖 对话模型: %d 个\n", len(chatModels))
	fmt.Printf("   🧠 嵌入模型: %d 个\n", len(embeddingModels))
	fmt.Printf("   🎨 图像模型: %d 个\n", len(imageModels))
	fmt.Printf("   🎵 音频模型: %d 个\n", len(audioModels))
	fmt.Printf("   🎬 视频模型: %d 个\n", len(videoModels))

	// 3. 更新对话模型
	fmt.Println("\n🤖 更新对话模型...")
	updateChatModels(db, chatModels)

	// 4. 更新嵌入模型
	fmt.Println("\n🧠 更新嵌入模型...")
	updateEmbeddingModels(db, embeddingModels)

	// 5. 更新图像模型
	fmt.Println("\n🎨 更新图像模型...")
	updateImageModels(db, imageModels)

	// 6. 更新音频模型
	fmt.Println("\n🎵 更新音频模型...")
	updateAudioModels(db, audioModels)

	// 7. 更新视频模型
	fmt.Println("\n🎬 更新视频模型...")
	updateVideoModels(db, videoModels)

	// 8. 显示更新后的统计
	fmt.Println("\n📊 更新后的模型统计:")
	showModelStats(db)

	fmt.Println("\n🎉 所有模型真实名称更新完成！")
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

func updateChatModels(db *sql.DB, models []Model) {
	updatedCount := 0
	
	for _, model := range models {
		// 尝试匹配数据库中的模型
		var dbID string
		var currentInternalKey string
		
		// 通过模型名称的关键词匹配
		modelLower := strings.ToLower(model.ID)
		var matchQuery string
		
		if strings.Contains(modelLower, "qwen") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'chat' AND (internal_key ILIKE '%qwen%' OR display_name ILIKE '%qwen%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "glm") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'chat' AND (internal_key ILIKE '%glm%' OR display_name ILIKE '%glm%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "deepseek") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'chat' AND (internal_key ILIKE '%deepseek%' OR display_name ILIKE '%deepseek%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "llama") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'chat' AND (internal_key ILIKE '%llama%' OR display_name ILIKE '%llama%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "internlm") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'chat' AND (internal_key ILIKE '%internlm%' OR display_name ILIKE '%internlm%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "yi") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'chat' AND (internal_key ILIKE '%yi%' OR display_name ILIKE '%yi%') AND internal_key != $1 LIMIT 1"
		} else {
			continue // 跳过无法匹配的模型
		}
		
		err := db.QueryRow(matchQuery, model.ID).Scan(&dbID, &currentInternalKey)
		if err != nil {
			continue // 没有找到匹配的模型
		}
		
		// 更新模型的真实名称
		_, err = db.Exec(
			"UPDATE ai_models SET internal_key = $1, updated_at = NOW() WHERE id = $2",
			model.ID, dbID,
		)
		
		if err != nil {
			log.Printf("更新对话模型失败 %s: %v", model.ID, err)
			continue
		}
		
		updatedCount++
		fmt.Printf("✅ %s -> %s\n", currentInternalKey, model.ID)
	}
	
	fmt.Printf("🎯 成功更新 %d 个对话模型\n", updatedCount)
}

func updateEmbeddingModels(db *sql.DB, models []Model) {
	updatedCount := 0
	
	for _, model := range models {
		var dbID string
		var currentInternalKey string
		
		// 匹配嵌入模型
		modelLower := strings.ToLower(model.ID)
		var matchQuery string
		
		if strings.Contains(modelLower, "bge") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'embedding' AND (internal_key ILIKE '%bge%' OR display_name ILIKE '%bge%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "embedding") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'embedding' AND (internal_key ILIKE '%embedding%' OR display_name ILIKE '%embedding%') AND internal_key != $1 LIMIT 1"
		} else {
			continue
		}
		
		err := db.QueryRow(matchQuery, model.ID).Scan(&dbID, &currentInternalKey)
		if err != nil {
			continue
		}
		
		_, err = db.Exec(
			"UPDATE ai_models SET internal_key = $1, updated_at = NOW() WHERE id = $2",
			model.ID, dbID,
		)
		
		if err != nil {
			log.Printf("更新嵌入模型失败 %s: %v", model.ID, err)
			continue
		}
		
		updatedCount++
		fmt.Printf("✅ %s -> %s\n", currentInternalKey, model.ID)
	}
	
	fmt.Printf("🎯 成功更新 %d 个嵌入模型\n", updatedCount)
}

func updateImageModels(db *sql.DB, models []Model) {
	updatedCount := 0
	
	for _, model := range models {
		var dbID string
		var currentInternalKey string
		
		// 匹配图像模型
		modelLower := strings.ToLower(model.ID)
		var matchQuery string
		
		if strings.Contains(modelLower, "flux") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'image' AND (internal_key ILIKE '%flux%' OR display_name ILIKE '%flux%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "stable") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'image' AND (internal_key ILIKE '%stable%' OR display_name ILIKE '%stable%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "kolors") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'image' AND (internal_key ILIKE '%kolors%' OR display_name ILIKE '%kolors%') AND internal_key != $1 LIMIT 1"
		} else {
			continue
		}
		
		err := db.QueryRow(matchQuery, model.ID).Scan(&dbID, &currentInternalKey)
		if err != nil {
			continue
		}
		
		_, err = db.Exec(
			"UPDATE ai_models SET internal_key = $1, updated_at = NOW() WHERE id = $2",
			model.ID, dbID,
		)
		
		if err != nil {
			log.Printf("更新图像模型失败 %s: %v", model.ID, err)
			continue
		}
		
		updatedCount++
		fmt.Printf("✅ %s -> %s\n", currentInternalKey, model.ID)
	}
	
	fmt.Printf("🎯 成功更新 %d 个图像模型\n", updatedCount)
}

func updateAudioModels(db *sql.DB, models []Model) {
	updatedCount := 0
	
	for _, model := range models {
		var dbID string
		var currentInternalKey string
		
		// 匹配音频模型
		modelLower := strings.ToLower(model.ID)
		var matchQuery string
		
		if strings.Contains(modelLower, "fish") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'audio' AND (internal_key ILIKE '%fish%' OR display_name ILIKE '%fish%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "cosyvoice") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'audio' AND (internal_key ILIKE '%cosyvoice%' OR display_name ILIKE '%cosyvoice%') AND internal_key != $1 LIMIT 1"
		} else if strings.Contains(modelLower, "sensevoice") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'audio' AND (internal_key ILIKE '%sensevoice%' OR display_name ILIKE '%sensevoice%') AND internal_key != $1 LIMIT 1"
		} else {
			continue
		}
		
		err := db.QueryRow(matchQuery, model.ID).Scan(&dbID, &currentInternalKey)
		if err != nil {
			continue
		}
		
		_, err = db.Exec(
			"UPDATE ai_models SET internal_key = $1, updated_at = NOW() WHERE id = $2",
			model.ID, dbID,
		)
		
		if err != nil {
			log.Printf("更新音频模型失败 %s: %v", model.ID, err)
			continue
		}
		
		updatedCount++
		fmt.Printf("✅ %s -> %s\n", currentInternalKey, model.ID)
	}
	
	fmt.Printf("🎯 成功更新 %d 个音频模型\n", updatedCount)
}

func updateVideoModels(db *sql.DB, models []Model) {
	updatedCount := 0
	
	for _, model := range models {
		var dbID string
		var currentInternalKey string
		
		// 匹配视频模型
		modelLower := strings.ToLower(model.ID)
		var matchQuery string
		
		if strings.Contains(modelLower, "wan") || strings.Contains(modelLower, "video") {
			matchQuery = "SELECT id, internal_key FROM ai_models WHERE model_type = 'video' AND (internal_key ILIKE '%wan%' OR internal_key ILIKE '%video%' OR display_name ILIKE '%wan%' OR display_name ILIKE '%video%') AND internal_key != $1 LIMIT 1"
		} else {
			continue
		}
		
		err := db.QueryRow(matchQuery, model.ID).Scan(&dbID, &currentInternalKey)
		if err != nil {
			continue
		}
		
		_, err = db.Exec(
			"UPDATE ai_models SET internal_key = $1, updated_at = NOW() WHERE id = $2",
			model.ID, dbID,
		)
		
		if err != nil {
			log.Printf("更新视频模型失败 %s: %v", model.ID, err)
			continue
		}
		
		updatedCount++
		fmt.Printf("✅ %s -> %s\n", currentInternalKey, model.ID)
	}
	
	fmt.Printf("🎯 成功更新 %d 个视频模型\n", updatedCount)
}

func showModelStats(db *sql.DB) {
	query := `
		SELECT 
			model_type,
			COUNT(*) as total_count,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_count
		FROM ai_models
		WHERE provider = 'siliconflow'
		GROUP BY model_type
		ORDER BY total_count DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询模型统计失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Println("-------------------------------------------")
	for rows.Next() {
		var modelType string
		var totalCount, activeCount int
		
		err := rows.Scan(&modelType, &totalCount, &activeCount)
		if err != nil {
			continue
		}
		
		var icon string
		switch modelType {
		case "chat":
			icon = "🤖"
		case "embedding":
			icon = "🧠"
		case "image":
			icon = "🎨"
		case "audio":
			icon = "🎵"
		case "video":
			icon = "🎬"
		default:
			icon = "❓"
		}
		
		fmt.Printf("%s %s: %d个模型 (活跃: %d)\n", icon, modelType, totalCount, activeCount)
	}
}
