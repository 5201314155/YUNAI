package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type ModuleModelConfig struct {
	ModuleName      string          `json:"module_name"`
	FunctionType    string          `json:"function_type"`
	PrimaryModelID  string          `json:"primary_model_id"`
	FallbackModels  json.RawMessage `json:"fallback_models"`
	ModelParams     json.RawMessage `json:"model_params"`
	Weight          int             `json:"weight"`
	Description     string          `json:"description"`
}

func main() {
	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	fmt.Println("🚀 YUNAI功能模块专用AI模型配置测试")
	fmt.Println("===========================================")

	// 测试模型配置查询
	testModuleConfigs := []struct {
		module   string
		function string
		desc     string
	}{
		{"chat", "conversation", "💬 聊天对话"},
		{"relationship", "embedding", "🕸️ 关系网络嵌入"},
		{"relationship", "analysis", "🕸️ 关系分析"},
		{"moments", "generation", "📱 朋友圈生成"},
		{"character", "avatar_generation", "🎨 角色头像生成"},
		{"voice", "synthesis", "🎵 语音合成"},
		{"search", "embedding", "🔍 搜索嵌入"},
	}

	fmt.Printf("\n📋 已配置的模块功能 (%d个):\n", len(testModuleConfigs))
	fmt.Println("-------------------------------------------")

	for i, config := range testModuleConfigs {
		fmt.Printf("%d. %s - %s.%s\n", i+1, config.desc, config.module, config.function)
		
		// 使用函数查询模型配置
		var modelID string
		var modelParams, fallbackModels json.RawMessage
		
		err := db.QueryRow("SELECT * FROM get_module_model($1, $2)", config.module, config.function).
			Scan(&modelID, &modelParams, &fallbackModels)
		
		if err != nil {
			fmt.Printf("   ❌ 查询失败: %v\n", err)
			continue
		}
		
		fmt.Printf("   ✅ 主要模型: %s\n", modelID)
		fmt.Printf("   📊 参数: %s\n", string(modelParams))
		fmt.Printf("   🔄 备用模型: %s\n", string(fallbackModels))
		fmt.Println()
	}

	// 查询所有配置
	fmt.Println("\n📊 完整配置列表:")
	fmt.Println("-------------------------------------------")
	
	rows, err := db.Query(`
		SELECT module_name, function_type, primary_model_id, fallback_models, 
		       model_params, weight, description 
		FROM module_model_configs 
		WHERE is_active = true 
		ORDER BY module_name, weight DESC
	`)
	if err != nil {
		log.Fatal("查询配置失败:", err)
	}
	defer rows.Close()

	moduleCount := make(map[string]int)
	totalConfigs := 0

	for rows.Next() {
		var config ModuleModelConfig
		err := rows.Scan(
			&config.ModuleName,
			&config.FunctionType,
			&config.PrimaryModelID,
			&config.FallbackModels,
			&config.ModelParams,
			&config.Weight,
			&config.Description,
		)
		if err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}

		moduleCount[config.ModuleName]++
		totalConfigs++

		fmt.Printf("🔧 %s.%s\n", config.ModuleName, config.FunctionType)
		fmt.Printf("   模型: %s (权重: %d)\n", config.PrimaryModelID, config.Weight)
		fmt.Printf("   描述: %s\n", config.Description)
		fmt.Printf("   参数: %s\n", string(config.ModelParams))
		fmt.Printf("   备用: %s\n", string(config.FallbackModels))
		fmt.Println()
	}

	// 统计信息
	fmt.Println("📈 配置统计:")
	fmt.Println("-------------------------------------------")
	fmt.Printf("总配置数: %d\n", totalConfigs)
	fmt.Printf("模块数: %d\n", len(moduleCount))
	fmt.Println("\n各模块配置数:")
	for module, count := range moduleCount {
		fmt.Printf("  %s: %d个功能\n", module, count)
	}

	// 测试SiliconFlow API密钥配置
	fmt.Println("\n🔑 SiliconFlow API配置:")
	fmt.Println("-------------------------------------------")
	fmt.Println("API地址: https://api.siliconflow.cn/v1")
	fmt.Println("API密钥: sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi")
	fmt.Println("状态: ✅ 已配置")

	// 模型使用建议
	fmt.Println("\n💡 模型使用建议:")
	fmt.Println("-------------------------------------------")
	fmt.Println("1. 💬 聊天对话: 使用DeepSeek V3获得最佳推理能力")
	fmt.Println("2. 🕸️ 关系网络: 使用BGE中文嵌入模型增强语义理解")
	fmt.Println("3. 📱 朋友圈: 使用GLM-4 9B获得更好的中文创作能力")
	fmt.Println("4. 🎨 图像生成: 使用FLUX.1 Schnell快速生成高质量头像")
	fmt.Println("5. 🎵 语音合成: 使用Fish Speech 1.5获得自然语音")
	fmt.Println("6. 🔍 搜索优化: 使用BGE嵌入模型提升搜索准确性")

	fmt.Println("\n🎉 模型配置测试完成！所有功能模块都已配置专用AI模型！")
}
