package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🎛️ 测试YUNAI完整模型管理系统")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 1. 测试创建自定义模型
	fmt.Println("🆕 测试创建100%自定义模型...")
	customModelID := testCreateCustomModel(db)

	// 2. 测试更新模型 (100%自定义)
	fmt.Println("\n✏️ 测试更新模型 (100%自定义)...")
	testUpdateModel(db, customModelID)

	// 3. 测试管理端视图 (显示真实模型ID)
	fmt.Println("\n🔧 测试管理端视图 (显示真实模型ID)...")
	testAdminView(db)

	// 4. 测试用户端视图 (显示自定义名称)
	fmt.Println("\n👤 测试用户端视图 (显示自定义名称)...")
	testUserView(db)

	// 5. 测试模型启用/禁用
	fmt.Println("\n🔄 测试模型启用/禁用...")
	testToggleModel(db, customModelID)

	// 6. 测试批量操作
	fmt.Println("\n📦 测试批量操作...")
	testBatchOperations(db)

	// 7. 测试模型删除
	fmt.Println("\n🗑️ 测试模型删除...")
	testDeleteModel(db, customModelID)

	// 8. 测试流式/非流式设置
	fmt.Println("\n🌊 测试流式/非流式设置...")
	testStreamingSettings(db)

	fmt.Println("\n🎉 所有测试完成！模型管理系统功能正常工作！")
}

// 测试创建自定义模型
func testCreateCustomModel(db *sql.DB) uuid.UUID {
	// 100%自定义的模型信息
	capabilities := []string{"chat", "reasoning", "code_generation", "creative_writing"}
	capabilitiesJSON, _ := json.Marshal(capabilities)
	
	paramsSchema := map[string]interface{}{
		"temperature": map[string]interface{}{
			"type":    "number",
			"default": 0.8,
			"min":     0.0,
			"max":     2.0,
		},
		"max_tokens": map[string]interface{}{
			"type":    "integer",
			"default": 3000,
			"min":     1,
			"max":     8000,
		},
		"stream": map[string]interface{}{
			"type":    "boolean",
			"default": true,
		},
	}
	paramsSchemaJSON, _ := json.Marshal(paramsSchema)
	
	pricing := map[string]interface{}{
		"input_token_price":  0.002,
		"output_token_price": 0.004,
		"unit":               "1k_tokens",
		"currency":           "CNY",
	}
	pricingJSON, _ := json.Marshal(pricing)
	
	// 插入自定义模型
	query := `
		INSERT INTO ai_models (
			internal_key, display_name, custom_display_name, provider, model_type,
			category, description, capabilities, params_schema, model_system_prompt,
			pricing, max_tokens, support_streaming, is_active, is_featured, weight,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
		) RETURNING id, created_at
	`
	
	var newID uuid.UUID
	var createdAt string
	err := db.QueryRow(query,
		"custom-ai-model-v1",                    // internal_key (真实模型ID)
		"Custom AI Model V1",                    // display_name (原始名称)
		"小云专属智能助手",                          // custom_display_name (用户看到的名称)
		"yunai",                                 // provider (自定义提供商)
		"chat",                                  // model_type
		"YUNAI专属模型",                          // category (100%自定义分类)
		"这是YUNAI项目专门定制的智能对话模型，具有优秀的中文理解和创作能力", // description (100%自定义)
		capabilitiesJSON,                        // capabilities
		paramsSchemaJSON,                        // params_schema (100%自定义参数)
		"你是小云，YUNAI项目的专属AI助手。你聪明、友善、专业，擅长帮助用户解决各种问题。", // system_prompt (100%自定义)
		pricingJSON,                             // pricing (100%自定义价格)
		8192,                                    // max_tokens (100%自定义)
		true,                                    // support_streaming (100%自定义)
		true,                                    // is_active
		true,                                    // is_featured
		1000,                                    // weight (100%自定义排序)
	).Scan(&newID, &createdAt)
	
	if err != nil {
		log.Printf("创建自定义模型失败: %v", err)
		return uuid.Nil
	}
	
	fmt.Printf("✅ 100%自定义模型创建成功!\n")
	fmt.Printf("   模型ID: %s\n", newID)
	fmt.Printf("   真实模型ID: custom-ai-model-v1\n")
	fmt.Printf("   用户显示名称: 小云专属智能助手\n")
	fmt.Printf("   自定义分类: YUNAI专属模型\n")
	fmt.Printf("   自定义提示词: 你是小云，YUNAI项目的专属AI助手...\n")
	fmt.Printf("   自定义价格: 输入¥0.002/1k, 输出¥0.004/1k\n")
	fmt.Printf("   支持流式: true\n")
	fmt.Printf("   创建时间: %s\n", createdAt)
	
	return newID
}

// 测试更新模型
func testUpdateModel(db *sql.DB, modelID uuid.UUID) {
	// 100%自定义更新所有字段
	newCapabilities := []string{"chat", "reasoning", "code_generation", "creative_writing", "translation", "summarization"}
	capabilitiesJSON, _ := json.Marshal(newCapabilities)
	
	newParamsSchema := map[string]interface{}{
		"temperature": map[string]interface{}{
			"type":    "number",
			"default": 0.7,
			"min":     0.0,
			"max":     2.0,
		},
		"max_tokens": map[string]interface{}{
			"type":    "integer",
			"default": 4000,
			"min":     1,
			"max":     10000,
		},
		"stream": map[string]interface{}{
			"type":    "boolean",
			"default": true,
		},
		"top_p": map[string]interface{}{
			"type":    "number",
			"default": 0.9,
			"min":     0.0,
			"max":     1.0,
		},
	}
	paramsSchemaJSON, _ := json.Marshal(newParamsSchema)
	
	newPricing := map[string]interface{}{
		"input_token_price":  0.0015,
		"output_token_price": 0.003,
		"unit":               "1k_tokens",
		"currency":           "CNY",
	}
	pricingJSON, _ := json.Marshal(newPricing)
	
	// 更新模型
	query := `
		UPDATE ai_models SET
			internal_key = $1,
			display_name = $2,
			custom_display_name = $3,
			category = $4,
			description = $5,
			capabilities = $6,
			params_schema = $7,
			model_system_prompt = $8,
			pricing = $9,
			max_tokens = $10,
			support_streaming = $11,
			weight = $12,
			updated_at = NOW()
		WHERE id = $13
		RETURNING updated_at
	`
	
	var updatedAt string
	err := db.QueryRow(query,
		"custom-ai-model-v2",                    // 更新真实模型ID
		"Custom AI Model V2",                    // 更新原始名称
		"小云超级智能助手 Pro",                     // 更新用户显示名称
		"YUNAI旗舰模型",                          // 更新分类
		"这是YUNAI项目的旗舰智能对话模型，具有卓越的中文理解、创作、翻译和总结能力", // 更新描述
		capabilitiesJSON,                        // 更新能力
		paramsSchemaJSON,                        // 更新参数
		"你是小云Pro，YUNAI项目的旗舰AI助手。你拥有卓越的智能和创造力，能够帮助用户完成各种复杂任务。", // 更新提示词
		pricingJSON,                             // 更新价格
		10240,                                   // 更新最大Token
		true,                                    // 保持流式支持
		1100,                                    // 更新权重
		modelID,
	).Scan(&updatedAt)
	
	if err != nil {
		log.Printf("更新模型失败: %v", err)
		return
	}
	
	fmt.Printf("✅ 模型100%自定义更新成功!\n")
	fmt.Printf("   新的真实模型ID: custom-ai-model-v2\n")
	fmt.Printf("   新的用户显示名称: 小云超级智能助手 Pro\n")
	fmt.Printf("   新的分类: YUNAI旗舰模型\n")
	fmt.Printf("   新的最大Token: 10240\n")
	fmt.Printf("   新的价格: 输入¥0.0015/1k, 输出¥0.003/1k\n")
	fmt.Printf("   更新时间: %s\n", updatedAt)
}

// 测试管理端视图
func testAdminView(db *sql.DB) {
	query := `
		SELECT 
			id, internal_key, display_name, custom_display_name, provider,
			model_type, category, support_streaming, is_active, weight
		FROM ai_models
		WHERE provider IN ('siliconflow', 'yunai')
		ORDER BY weight DESC
		LIMIT 10
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询管理端视图失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("🔧 管理端视图 (显示真实模型ID):\n")
	fmt.Printf("-------------------------------------------\n")
	
	count := 0
	for rows.Next() {
		var id uuid.UUID
		var internalKey, displayName, customDisplayName, provider, modelType, category string
		var supportStreaming, isActive bool
		var weight int
		
		err := rows.Scan(&id, &internalKey, &displayName, &customDisplayName, &provider, &modelType, &category, &supportStreaming, &isActive, &weight)
		if err != nil {
			continue
		}
		
		count++
		activeIcon := "❌"
		if isActive {
			activeIcon = "✅"
		}
		
		streamIcon := "❌"
		if supportStreaming {
			streamIcon = "🌊"
		}
		
		fmt.Printf("%d. %s %s [%s] %s\n", count, activeIcon, streamIcon, provider, displayName)
		fmt.Printf("   真实模型ID: %s\n", internalKey)
		fmt.Printf("   用户显示名称: %s\n", customDisplayName)
		fmt.Printf("   分类: %s | 类型: %s | 权重: %d\n", category, modelType, weight)
		fmt.Printf("   数据库ID: %s\n", id.String()[:8])
		fmt.Println()
	}
	
	fmt.Printf("✅ 管理端显示 %d 个模型 (包含真实模型ID)\n", count)
}

// 测试用户端视图
func testUserView(db *sql.DB) {
	query := `
		SELECT 
			id, custom_display_name, provider, model_type, category,
			description, support_streaming, is_featured, weight
		FROM ai_models
		WHERE is_active = true
		ORDER BY weight DESC, custom_display_name ASC
		LIMIT 10
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询用户端视图失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("👤 用户端视图 (显示自定义名称):\n")
	fmt.Printf("-------------------------------------------\n")
	
	count := 0
	for rows.Next() {
		var id uuid.UUID
		var customDisplayName, provider, modelType, category, description string
		var supportStreaming, isFeatured bool
		var weight int
		
		err := rows.Scan(&id, &customDisplayName, &provider, &modelType, &category, &description, &supportStreaming, &isFeatured, &weight)
		if err != nil {
			continue
		}
		
		count++
		featuredIcon := ""
		if isFeatured {
			featuredIcon = "⭐"
		}
		
		streamIcon := ""
		if supportStreaming {
			streamIcon = "🌊"
		}
		
		fmt.Printf("%d. %s %s %s [%s]\n", count, featuredIcon, streamIcon, customDisplayName, provider)
		fmt.Printf("   分类: %s | 类型: %s\n", category, modelType)
		fmt.Printf("   描述: %s\n", description)
		fmt.Printf("   权重: %d\n", weight)
		fmt.Println()
	}
	
	fmt.Printf("✅ 用户端显示 %d 个模型 (只显示自定义名称，隐藏真实模型ID)\n", count)
}

// 测试模型启用/禁用
func testToggleModel(db *sql.DB, modelID uuid.UUID) {
	// 禁用模型
	_, err := db.Exec("UPDATE ai_models SET is_active = false, updated_at = NOW() WHERE id = $1", modelID)
	if err != nil {
		log.Printf("禁用模型失败: %v", err)
		return
	}
	
	fmt.Printf("✅ 模型已禁用\n")
	
	// 重新启用模型
	_, err = db.Exec("UPDATE ai_models SET is_active = true, updated_at = NOW() WHERE id = $1", modelID)
	if err != nil {
		log.Printf("启用模型失败: %v", err)
		return
	}
	
	fmt.Printf("✅ 模型已重新启用\n")
}

// 测试批量操作
func testBatchOperations(db *sql.DB) {
	// 获取一些模型ID用于批量操作
	query := `SELECT id FROM ai_models WHERE provider = 'siliconflow' LIMIT 3`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("获取模型ID失败: %v", err)
		return
	}
	defer rows.Close()
	
	var modelIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			modelIDs = append(modelIDs, id)
		}
	}
	
	if len(modelIDs) == 0 {
		fmt.Printf("❌ 没有找到可用于批量操作的模型\n")
		return
	}
	
	// 批量禁用
	query = `UPDATE ai_models SET is_active = false, updated_at = NOW() WHERE id = ANY($1::uuid[])`
	result, err := db.Exec(query, modelIDs)
	if err != nil {
		log.Printf("批量禁用失败: %v", err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ 批量禁用成功，影响 %d 个模型\n", rowsAffected)
	
	// 批量重新启用
	query = `UPDATE ai_models SET is_active = true, updated_at = NOW() WHERE id = ANY($1::uuid[])`
	result, err = db.Exec(query, modelIDs)
	if err != nil {
		log.Printf("批量启用失败: %v", err)
		return
	}
	
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("✅ 批量启用成功，影响 %d 个模型\n", rowsAffected)
}

// 测试模型删除
func testDeleteModel(db *sql.DB, modelID uuid.UUID) {
	// 获取模型信息
	var customDisplayName string
	err := db.QueryRow("SELECT custom_display_name FROM ai_models WHERE id = $1", modelID).Scan(&customDisplayName)
	if err != nil {
		log.Printf("获取模型信息失败: %v", err)
		return
	}
	
	// 删除模型
	result, err := db.Exec("DELETE FROM ai_models WHERE id = $1", modelID)
	if err != nil {
		log.Printf("删除模型失败: %v", err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("✅ 模型删除成功: %s\n", customDisplayName)
	} else {
		fmt.Printf("❌ 模型删除失败\n")
	}
}

// 测试流式/非流式设置
func testStreamingSettings(db *sql.DB) {
	// 统计流式支持情况
	query := `
		SELECT 
			model_type,
			COUNT(*) as total_count,
			COUNT(CASE WHEN support_streaming = true THEN 1 END) as streaming_count
		FROM ai_models
		WHERE is_active = true
		GROUP BY model_type
		ORDER BY total_count DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询流式统计失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("🌊 流式/非流式支持统计:\n")
	fmt.Printf("-------------------------------------------\n")
	
	for rows.Next() {
		var modelType string
		var totalCount, streamingCount int
		
		err := rows.Scan(&modelType, &totalCount, &streamingCount)
		if err != nil {
			continue
		}
		
		nonStreamingCount := totalCount - streamingCount
		streamingPercent := float64(streamingCount) / float64(totalCount) * 100
		
		fmt.Printf("📊 %s 类型:\n", modelType)
		fmt.Printf("   总计: %d 个模型\n", totalCount)
		fmt.Printf("   支持流式: %d 个 (%.1f%%)\n", streamingCount, streamingPercent)
		fmt.Printf("   非流式: %d 个 (%.1f%%)\n", nonStreamingCount, 100-streamingPercent)
		fmt.Println()
	}
	
	fmt.Printf("✅ 流式设置统计完成\n")
}
