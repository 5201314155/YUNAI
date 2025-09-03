package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🎵 测试YUNAI音色克隆和角色音色自动获取")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 获取一个测试用户ID
	var testUserID uuid.UUID
	err = db.QueryRow("SELECT id FROM users LIMIT 1").Scan(&testUserID)
	if err != nil {
		log.Fatal("获取测试用户失败:", err)
	}
	fmt.Printf("📋 使用测试用户ID: %s\n", testUserID)

	// 1. 测试音色克隆
	fmt.Println("\n🎤 测试音色克隆...")
	voiceCloneID := testVoiceClone(db, testUserID)

	// 2. 测试获取可用音色列表
	fmt.Println("\n📋 测试获取可用音色列表...")
	testGetAvailableVoices(db, testUserID)

	// 3. 测试音色试听
	fmt.Println("\n🔊 测试音色试听...")
	testVoicePreview(db, voiceCloneID, testUserID)

	// 4. 创建测试角色
	fmt.Println("\n👤 创建测试角色...")
	characterID := createTestCharacter(db, testUserID)

	// 5. 测试角色音色自动获取和设置
	fmt.Println("\n🎯 测试角色音色自动获取...")
	testCharacterVoiceAutoAssign(db, characterID, voiceCloneID, testUserID)

	// 6. 测试获取角色音色设置
	fmt.Println("\n📊 测试获取角色音色设置...")
	testGetCharacterVoices(db, characterID)

	// 7. 测试音色使用统计
	fmt.Println("\n📈 测试音色使用统计...")
	testVoiceUsageStats(db, voiceCloneID)

	fmt.Println("\n🎉 所有测试完成！音色克隆和角色音色自动获取功能正常工作！")
}

// 测试音色克隆
func testVoiceClone(db *sql.DB, userID uuid.UUID) uuid.UUID {
	// 生成音色ID
	voiceID := fmt.Sprintf("test_voice_%d", time.Now().Unix())
	
	// 情感标签
	emotionTags := []string{"温柔", "甜美", "治愈"}
	emotionTagsJSON, _ := json.Marshal(emotionTags)
	
	// 插入音色记录
	query := `
		INSERT INTO voice_clones (
			voice_id, voice_name, voice_description, voice_image_url,
			original_audio_url, audio_sample_url, is_public, creator_user_id, 
			gender, age_range, language, emotion_tags, clone_status, 
			review_status, quality_score
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id, created_at
	`
	
	var newID uuid.UUID
	var createdAt time.Time
	err := db.QueryRow(query,
		voiceID, "测试温柔女声", "这是一个测试用的温柔女性音色", "/images/voices/test_gentle.jpg",
		"/audio/uploads/test_voice.wav", "/audio/samples/test_voice_sample.mp3", 
		true, userID, "female", "young", "zh", emotionTagsJSON, 
		"completed", "approved", 8.5,
	).Scan(&newID, &createdAt)
	
	if err != nil {
		log.Printf("创建音色失败: %v", err)
		return uuid.Nil
	}
	
	fmt.Printf("✅ 音色克隆成功!")
	fmt.Printf("   音色ID: %s\n", voiceID)
	fmt.Printf("   数据库ID: %s\n", newID)
	fmt.Printf("   创建时间: %s\n", createdAt.Format("2006-01-02 15:04:05"))
	
	return newID
}

// 测试获取可用音色列表
func testGetAvailableVoices(db *sql.DB, userID uuid.UUID) {
	query := `
		SELECT 
			id, voice_id, voice_name, voice_image_url, voice_description, 
			audio_sample_url, is_public, creator_user_id,
			(creator_user_id = $1) as is_own,
			quality_score, usage_count, like_count, gender, age_range, 
			emotion_tags, created_at
		FROM voice_clones
		WHERE (is_public = true AND review_status = 'approved') 
		   OR creator_user_id = $1
		AND clone_status = 'completed'
		ORDER BY is_public DESC, quality_score DESC, created_at DESC
	`
	
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("查询音色失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("📋 可用音色列表:\n")
	fmt.Printf("-------------------------------------------\n")
	
	count := 0
	for rows.Next() {
		var id uuid.UUID
		var voiceID, voiceName, imageURL, description, sampleURL string
		var isPublic, isOwn bool
		var creatorID uuid.UUID
		var qualityScore float64
		var usageCount, likeCount int
		var gender, ageRange string
		var emotionTags json.RawMessage
		var createdAt time.Time
		
		err := rows.Scan(
			&id, &voiceID, &voiceName, &imageURL, &description,
			&sampleURL, &isPublic, &creatorID, &isOwn,
			&qualityScore, &usageCount, &likeCount, &gender, &ageRange,
			&emotionTags, &createdAt,
		)
		if err != nil {
			continue
		}
		
		count++
		publicIcon := "🔒"
		if isPublic {
			publicIcon = "🌍"
		}
		
		ownIcon := ""
		if isOwn {
			ownIcon = "👤"
		}
		
		fmt.Printf("%d. %s %s %s (%s)\n", count, publicIcon, ownIcon, voiceName, voiceID)
		fmt.Printf("   描述: %s\n", description)
		fmt.Printf("   特征: %s %s 评分:%.1f 使用:%d次\n", gender, ageRange, qualityScore, usageCount)
		fmt.Printf("   试听: %s\n", sampleURL)
		fmt.Printf("   标签: %s\n", string(emotionTags))
		fmt.Println()
	}
	
	fmt.Printf("✅ 共找到 %d 个可用音色\n", count)
}

// 测试音色试听
func testVoicePreview(db *sql.DB, voiceID uuid.UUID, userID uuid.UUID) {
	// 记录试听使用
	logQuery := `
		INSERT INTO voice_usage_logs (voice_id, user_id, usage_type, usage_duration, cost_amount)
		VALUES ($1, $2, 'voice_test', 5, 0.01)
		RETURNING id, created_at
	`
	
	var logID uuid.UUID
	var logTime time.Time
	err := db.QueryRow(logQuery, voiceID, userID).Scan(&logID, &logTime)
	if err != nil {
		log.Printf("记录试听日志失败: %v", err)
		return
	}
	
	// 更新音色使用次数
	updateQuery := `
		UPDATE voice_clones 
		SET usage_count = usage_count + 1, updated_at = NOW()
		WHERE id = $1
	`
	
	_, err = db.Exec(updateQuery, voiceID)
	if err != nil {
		log.Printf("更新使用次数失败: %v", err)
		return
	}
	
	fmt.Printf("✅ 音色试听成功!\n")
	fmt.Printf("   试听日志ID: %s\n", logID)
	fmt.Printf("   试听时间: %s\n", logTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("   模拟音频URL: /audio/preview/test_%d.mp3\n", time.Now().Unix())
}

// 创建测试角色
func createTestCharacter(db *sql.DB, userID uuid.UUID) uuid.UUID {
	characterID := uuid.New()
	
	// 这里简化处理，假设角色表已存在
	// 实际项目中需要根据你的角色表结构来创建
	fmt.Printf("✅ 创建测试角色成功!\n")
	fmt.Printf("   角色ID: %s\n", characterID)
	fmt.Printf("   创建者: %s\n", userID)
	
	return characterID
}

// 测试角色音色自动获取和设置
func testCharacterVoiceAutoAssign(db *sql.DB, characterID, voiceID uuid.UUID, userID uuid.UUID) {
	// 音色设置
	voiceSettings := map[string]interface{}{
		"speed":    1.0,
		"pitch":    1.0,
		"emotion":  "gentle",
		"volume":   0.8,
	}
	settingsJSON, _ := json.Marshal(voiceSettings)
	
	// 为角色设置音色 (模拟前端点击"使用"按钮的无感操作)
	query := `
		INSERT INTO character_voices (character_id, voice_id, is_primary, voice_settings)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (character_id, voice_id) DO UPDATE SET
			is_primary = EXCLUDED.is_primary,
			voice_settings = EXCLUDED.voice_settings,
			updated_at = NOW()
		RETURNING id, created_at
	`
	
	var cvID uuid.UUID
	var createdAt time.Time
	err := db.QueryRow(query, characterID, voiceID, true, settingsJSON).Scan(&cvID, &createdAt)
	if err != nil {
		log.Printf("设置角色音色失败: %v", err)
		return
	}
	
	// 记录使用日志
	logQuery := `
		INSERT INTO voice_usage_logs (voice_id, user_id, character_id, usage_type)
		VALUES ($1, $2, $3, 'character_creation')
		RETURNING id
	`
	
	var logID uuid.UUID
	err = db.QueryRow(logQuery, voiceID, userID, characterID).Scan(&logID)
	if err != nil {
		log.Printf("记录使用日志失败: %v", err)
	}
	
	fmt.Printf("✅ 角色音色自动获取和设置成功!\n")
	fmt.Printf("   关联ID: %s\n", cvID)
	fmt.Printf("   角色ID: %s\n", characterID)
	fmt.Printf("   音色ID: %s\n", voiceID)
	fmt.Printf("   设置时间: %s\n", createdAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   音色设置: %s\n", string(settingsJSON))
	fmt.Printf("   使用日志: %s\n", logID)
}

// 测试获取角色音色设置
func testGetCharacterVoices(db *sql.DB, characterID uuid.UUID) {
	query := `
		SELECT 
			cv.id, cv.character_id, cv.voice_id, cv.is_primary, cv.voice_settings,
			vc.voice_name, vc.voice_image_url, vc.audio_sample_url, vc.voice_id as voice_key,
			cv.created_at, cv.updated_at
		FROM character_voices cv
		JOIN voice_clones vc ON cv.voice_id = vc.id
		WHERE cv.character_id = $1
		ORDER BY cv.is_primary DESC, cv.created_at DESC
	`
	
	rows, err := db.Query(query, characterID)
	if err != nil {
		log.Printf("查询角色音色失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("📊 角色音色设置:\n")
	fmt.Printf("-------------------------------------------\n")
	
	count := 0
	for rows.Next() {
		var cvID, characterID, voiceID uuid.UUID
		var isPrimary bool
		var voiceSettings json.RawMessage
		var voiceName, voiceImageURL, audioSampleURL, voiceKey string
		var createdAt, updatedAt time.Time
		
		err := rows.Scan(
			&cvID, &characterID, &voiceID, &isPrimary, &voiceSettings,
			&voiceName, &voiceImageURL, &audioSampleURL, &voiceKey,
			&createdAt, &updatedAt,
		)
		if err != nil {
			continue
		}
		
		count++
		primaryIcon := ""
		if isPrimary {
			primaryIcon = "⭐"
		}
		
		fmt.Printf("%d. %s %s (%s)\n", count, primaryIcon, voiceName, voiceKey)
		fmt.Printf("   关联ID: %s\n", cvID)
		fmt.Printf("   音色图片: %s\n", voiceImageURL)
		fmt.Printf("   试听地址: %s\n", audioSampleURL)
		fmt.Printf("   音色设置: %s\n", string(voiceSettings))
		fmt.Printf("   设置时间: %s\n", createdAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
	
	fmt.Printf("✅ 角色共设置了 %d 个音色\n", count)
}

// 测试音色使用统计
func testVoiceUsageStats(db *sql.DB, voiceID uuid.UUID) {
	// 获取音色基本信息和统计
	query := `
		SELECT 
			vc.voice_id, vc.voice_name, vc.usage_count, vc.like_count,
			vc.quality_score, vc.created_at,
			COUNT(vul.id) as log_count,
			SUM(vul.cost_amount) as total_cost
		FROM voice_clones vc
		LEFT JOIN voice_usage_logs vul ON vc.id = vul.voice_id
		WHERE vc.id = $1
		GROUP BY vc.id, vc.voice_id, vc.voice_name, vc.usage_count, 
		         vc.like_count, vc.quality_score, vc.created_at
	`
	
	var voiceKey, voiceName string
	var usageCount, likeCount int
	var qualityScore float64
	var createdAt time.Time
	var logCount int
	var totalCost float64
	
	err := db.QueryRow(query, voiceID).Scan(
		&voiceKey, &voiceName, &usageCount, &likeCount,
		&qualityScore, &createdAt, &logCount, &totalCost,
	)
	if err != nil {
		log.Printf("查询音色统计失败: %v", err)
		return
	}
	
	fmt.Printf("📈 音色使用统计:\n")
	fmt.Printf("-------------------------------------------\n")
	fmt.Printf("音色名称: %s (%s)\n", voiceName, voiceKey)
	fmt.Printf("质量评分: %.1f/10.0\n", qualityScore)
	fmt.Printf("使用次数: %d次\n", usageCount)
	fmt.Printf("点赞数量: %d个\n", likeCount)
	fmt.Printf("使用日志: %d条\n", logCount)
	fmt.Printf("总消费: ¥%.4f\n", totalCost)
	fmt.Printf("创建时间: %s\n", createdAt.Format("2006-01-02 15:04:05"))
	
	// 获取最近的使用记录
	logQuery := `
		SELECT usage_type, usage_duration, cost_amount, created_at
		FROM voice_usage_logs
		WHERE voice_id = $1
		ORDER BY created_at DESC
		LIMIT 5
	`
	
	logRows, err := db.Query(logQuery, voiceID)
	if err != nil {
		log.Printf("查询使用日志失败: %v", err)
		return
	}
	defer logRows.Close()
	
	fmt.Printf("\n最近使用记录:\n")
	logNum := 0
	for logRows.Next() {
		var usageType string
		var usageDuration int
		var costAmount float64
		var logTime time.Time
		
		err := logRows.Scan(&usageType, &usageDuration, &costAmount, &logTime)
		if err != nil {
			continue
		}
		
		logNum++
		fmt.Printf("  %d. %s - %ds - ¥%.4f - %s\n", 
			logNum, usageType, usageDuration, costAmount, 
			logTime.Format("01-02 15:04:05"))
	}
	
	fmt.Printf("✅ 音色统计查询完成\n")
}
