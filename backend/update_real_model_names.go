package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔄 更新数据库模型名称为真实SiliconFlow名称")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 模型名称映射表 (从我们的命名 -> 真实SiliconFlow名称)
	modelMappings := map[string]string{
		// GLM系列
		"sf_thudm_glm_4_9b_chat":      "THUDM/glm-4-9b-chat",
		"sf_pro_thudm_glm_4_9b_chat":  "Pro/THUDM/glm-4-9b-chat",
		
		// Qwen系列
		"sf_qwen_qwen2_7b_instruct":     "Qwen/Qwen2-7B-Instruct",
		"sf_pro_qwen_qwen2_7b_instruct": "Pro/Qwen/Qwen2-7B-Instruct",
		"sf_qwen_qwen2_5_7b_instruct":   "Qwen/Qwen2.5-7B-Instruct",
		"sf_qwen_qwen2_5_72b_instruct":  "Qwen/Qwen2.5-72B-Instruct",
		
		// InternLM系列
		"sf_internlm_internlm2_5_7b_chat": "internlm/internlm2_5-7b-chat",
		
		// DeepSeek系列
		"sf_deepseek_deepseek_chat":       "deepseek-ai/deepseek-chat",
		"sf_deepseek_deepseek_coder":      "deepseek-ai/deepseek-coder",
		"sf_deepseek_deepseek_v3":         "deepseek-ai/DeepSeek-V3",
		
		// Llama系列
		"sf_meta_llama_3_1_8b_instruct":  "meta-llama/Meta-Llama-3.1-8B-Instruct",
		"sf_meta_llama_3_1_70b_instruct": "meta-llama/Meta-Llama-3.1-70B-Instruct",
		"sf_meta_llama_3_2_3b_instruct":  "meta-llama/Llama-3.2-3B-Instruct",
		
		// Yi系列
		"sf_01_ai_yi_1_5_9b_chat": "01-ai/Yi-1.5-9B-Chat",
		
		// 其他常见模型
		"sf_mistral_mistral_7b_instruct": "mistralai/Mistral-7B-Instruct-v0.3",
		"sf_google_gemma_2_9b_it":        "google/gemma-2-9b-it",
	}

	fmt.Printf("📋 准备更新 %d 个模型的真实名称\n", len(modelMappings))

	// 批量更新模型名称
	updatedCount := 0
	for oldName, newName := range modelMappings {
		result, err := db.Exec(
			"UPDATE ai_models SET internal_key = $1, updated_at = NOW() WHERE internal_key = $2",
			newName, oldName,
		)
		
		if err != nil {
			log.Printf("更新模型 %s -> %s 失败: %v", oldName, newName, err)
			continue
		}
		
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			updatedCount++
			fmt.Printf("✅ %s -> %s\n", oldName, newName)
		}
	}

	fmt.Printf("\n🎉 成功更新 %d 个模型名称！\n", updatedCount)

	// 显示更新后的对话模型列表
	fmt.Println("\n📋 更新后的对话模型列表:")
	fmt.Println("-------------------------------------------")
	
	query := `
		SELECT id, internal_key, custom_display_name, support_streaming
		FROM ai_models 
		WHERE model_type = 'chat' AND provider = 'siliconflow' AND is_active = true
		ORDER BY weight DESC
		LIMIT 10
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询更新后的模型失败: %v", err)
		return
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		var id, internalKey, customDisplayName string
		var supportStreaming bool
		
		err := rows.Scan(&id, &internalKey, &customDisplayName, &supportStreaming)
		if err != nil {
			continue
		}
		
		count++
		streamIcon := "❌"
		if supportStreaming {
			streamIcon = "🌊"
		}
		
		fmt.Printf("%d. %s %s\n", count, streamIcon, customDisplayName)
		fmt.Printf("   真实模型名称: %s\n", internalKey)
		fmt.Printf("   数据库ID: %s\n", id[:8])
		fmt.Println()
	}
	
	fmt.Printf("✅ 显示了前 %d 个对话模型\n", count)
	fmt.Println("\n🚀 现在可以使用真实的SiliconFlow模型名称进行对话了！")
}
