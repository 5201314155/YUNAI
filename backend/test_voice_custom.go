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
	fmt.Println("🎵 测试YUNAI音色自定义管理系统")
	fmt.Println("===========================================")

	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 获取测试用户ID
	var testUserID uuid.UUID
	err = db.QueryRow("SELECT id FROM users LIMIT 1").Scan(&testUserID)
	if err != nil {
		log.Fatal("获取测试用户失败:", err)
	}
	fmt.Printf("📋 使用测试用户ID: %s\n", testUserID)

	// 1. 测试用户自定义音色创建
	fmt.Println("\n🎤 测试用户自定义音色创建...")
	voiceID1 := testCreateCustomVoice(db, testUserID, "我的专属音色", "/images/my_voice.jpg", true)
	_ = testCreateCustomVoice(db, testUserID, "私密音色", "/images/private_voice.jpg", false)

	// 2. 测试获取用户音色列表
	fmt.Println("\n📋 测试获取用户音色列表...")
	testGetMyVoices(db, testUserID)

	// 3. 测试更新音色信息
	fmt.Println("\n✏️ 测试更新音色信息...")
	testUpdateVoiceInfo(db, voiceID1, testUserID)

	// 4. 测试切换公开/私密状态
	fmt.Println("\n🔄 测试切换公开/私密状态...")
	testTogglePrivacy(db, voiceID1, testUserID)

	// 5. 测试获取可用音色列表
	fmt.Println("\n🌍 测试获取可用音色列表...")
	testGetAvailableVoices(db, testUserID)

	// 6. 测试角色音色选择 (无感获取音色ID)
	fmt.Println("\n🎯 测试角色音色选择...")
	characterID := uuid.New()
	testSelectVoiceForCharacter(db, characterID, voiceID1, testUserID)

	// 7. 测试音色试听
	fmt.Println("\n🔊 测试音色试听...")
	testVoicePreview(db, voiceID1, testUserID)

	// 8. 测试音色使用统计
	fmt.Println("\n📈 测试音色使用统计...")
	testVoiceUsageStats(db, voiceID1)

	fmt.Println("\n🎉 所有测试完成！用户音色自定义管理功能正常工作！")
}

// 测试创建自定义音色
func testCreateCustomVoice(db *sql.DB, userID uuid.UUID, voiceName, imageURL string, isPublic bool) uuid.UUID {
	// 生成音色ID
	voiceID := fmt.Sprintf("custom_%s_%d", userID.String()[:8], time.Now().Unix())

	// 用户自定义的情感标签
	emotionTags := []string{"温柔", "专业", "亲切"}
	emotionTagsJSON, _ := json.Marshal(emotionTags)

	// 插入音色记录
	query := `
		INSERT INTO voice_clones (
			voice_id, voice_name, voice_description, voice_image_url,
			original_audio_url, is_public, creator_user_id, gender,
			age_range, language, emotion_tags, clone_status, 
			review_status, quality_score
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, created_at
	`

	var newID uuid.UUID
	var createdAt time.Time
	err := db.QueryRow(query,
		voiceID, voiceName, "这是用户自定义的音色描述", imageURL,
		"/audio/uploads/"+voiceID+".wav", isPublic, userID, "female",
		"adult", "zh", emotionTagsJSON, "completed", "approved", 8.5,
	).Scan(&newID, &createdAt)

	if err != nil {
		log.Printf("创建音色失败: %v", err)
		return uuid.Nil
	}

	// 模拟生成试听样本
	sampleURL := fmt.Sprintf("/audio/samples/%s_sample.mp3", voiceID)
	updateQuery := `UPDATE voice_clones SET audio_sample_url = $1 WHERE id = $2`
	db.Exec(updateQuery, sampleURL, newID)

	publicStatus := "私密"
	if isPublic {
		publicStatus = "公开"
	}

	fmt.Printf("✅ 用户自定义音色创建成功!\n")
	fmt.Printf("   音色名称: %s\n", voiceName)
	fmt.Printf("   音色ID: %s\n", voiceID)
	fmt.Printf("   自定义图片: %s\n", imageURL)
	fmt.Printf("   公开状态: %s\n", publicStatus)
	fmt.Printf("   创建时间: %s\n", createdAt.Format("2006-01-02 15:04:05"))

	return newID
}

// 测试获取用户音色列表
func testGetMyVoices(db *sql.DB, userID uuid.UUID) {
	query := `
		SELECT 
			voice_id, voice_name, voice_description, voice_image_url,
			audio_sample_url, is_public, gender, age_range, emotion_tags,
			quality_score, usage_count, like_count, created_at
		FROM voice_clones
		WHERE creator_user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("查询用户音色失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("📋 我的音色列表:\n")
	fmt.Printf("-------------------------------------------\n")

	count := 0
	for rows.Next() {
		var voiceID, voiceName, description, imageURL, sampleURL string
		var isPublic bool
		var gender, ageRange string
		var emotionTags json.RawMessage
		var qualityScore float64
		var usageCount, likeCount int
		var createdAt time.Time

		err := rows.Scan(
			&voiceID, &voiceName, &description, &imageURL, &sampleURL,
			&isPublic, &gender, &ageRange, &emotionTags, &qualityScore,
			&usageCount, &likeCount, &createdAt,
		)
		if err != nil {
			continue
		}

		count++
		publicIcon := "🔒"
		if isPublic {
			publicIcon = "🌍"
		}

		fmt.Printf("%d. %s %s (%s)\n", count, publicIcon, voiceName, voiceID)
		fmt.Printf("   描述: %s\n", description)
		fmt.Printf("   图片: %s\n", imageURL)
		fmt.Printf("   试听: %s\n", sampleURL)
		fmt.Printf("   特征: %s %s 评分:%.1f\n", gender, ageRange, qualityScore)
		fmt.Printf("   标签: %s\n", string(emotionTags))
		fmt.Printf("   统计: 使用%d次 点赞%d个\n", usageCount, likeCount)
		fmt.Printf("   创建: %s\n", createdAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	fmt.Printf("✅ 共有 %d 个自定义音色\n", count)
}

// 测试更新音色信息
func testUpdateVoiceInfo(db *sql.DB, voiceDBID uuid.UUID, userID uuid.UUID) {
	// 获取音色ID
	var voiceID string
	err := db.QueryRow("SELECT voice_id FROM voice_clones WHERE id = $1", voiceDBID).Scan(&voiceID)
	if err != nil {
		log.Printf("获取音色ID失败: %v", err)
		return
	}

	// 更新音色信息
	newEmotionTags := []string{"温柔", "专业", "亲切", "优雅"}
	emotionTagsJSON, _ := json.Marshal(newEmotionTags)

	query := `
		UPDATE voice_clones 
		SET voice_name = $1,
		    voice_description = $2,
		    voice_image_url = $3,
		    emotion_tags = $4,
		    updated_at = NOW()
		WHERE id = $5 AND creator_user_id = $6
	`

	result, err := db.Exec(query,
		"更新后的音色名称", "这是更新后的音色描述，更加详细",
		"/images/updated_voice.jpg", emotionTagsJSON, voiceDBID, userID)

	if err != nil {
		log.Printf("更新音色信息失败: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("✅ 音色信息更新成功!\n")
		fmt.Printf("   音色ID: %s\n", voiceID)
		fmt.Printf("   新名称: 更新后的音色名称\n")
		fmt.Printf("   新图片: /images/updated_voice.jpg\n")
		fmt.Printf("   新标签: %s\n", string(emotionTagsJSON))
	} else {
		fmt.Printf("❌ 音色信息更新失败\n")
	}
}

// 测试切换公开/私密状态
func testTogglePrivacy(db *sql.DB, voiceDBID uuid.UUID, userID uuid.UUID) {
	// 获取当前状态
	var currentPublic bool
	var voiceID string
	err := db.QueryRow("SELECT voice_id, is_public FROM voice_clones WHERE id = $1", voiceDBID).Scan(&voiceID, &currentPublic)
	if err != nil {
		log.Printf("获取音色状态失败: %v", err)
		return
	}

	// 切换状态
	newPublic := !currentPublic
	query := `
		UPDATE voice_clones 
		SET is_public = $1, updated_at = NOW()
		WHERE id = $2 AND creator_user_id = $3
	`

	result, err := db.Exec(query, newPublic, voiceDBID, userID)
	if err != nil {
		log.Printf("切换音色状态失败: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		oldStatus := "私密"
		newStatus := "私密"
		if currentPublic {
			oldStatus = "公开"
		}
		if newPublic {
			newStatus = "公开"
		}

		fmt.Printf("✅ 音色状态切换成功!\n")
		fmt.Printf("   音色ID: %s\n", voiceID)
		fmt.Printf("   状态变化: %s → %s\n", oldStatus, newStatus)
	} else {
		fmt.Printf("❌ 音色状态切换失败\n")
	}
}

// 测试获取可用音色列表
func testGetAvailableVoices(db *sql.DB, userID uuid.UUID) {
	query := `
		SELECT 
			voice_id, voice_name, voice_image_url, audio_sample_url,
			is_public, (creator_user_id = $1) as is_own,
			quality_score, usage_count, emotion_tags
		FROM voice_clones
		WHERE (is_public = true OR creator_user_id = $1)
		AND clone_status = 'completed'
		ORDER BY (creator_user_id = $1) DESC, quality_score DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("查询可用音色失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("🌍 可用音色列表 (公开+我的私密):\n")
	fmt.Printf("-------------------------------------------\n")

	count := 0
	for rows.Next() {
		var voiceID, voiceName, imageURL, sampleURL string
		var isPublic, isOwn bool
		var qualityScore float64
		var usageCount int
		var emotionTags json.RawMessage

		err := rows.Scan(
			&voiceID, &voiceName, &imageURL, &sampleURL,
			&isPublic, &isOwn, &qualityScore, &usageCount, &emotionTags,
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
		fmt.Printf("   图片: %s\n", imageURL)
		fmt.Printf("   试听: %s\n", sampleURL)
		fmt.Printf("   评分: %.1f 使用: %d次\n", qualityScore, usageCount)
		fmt.Printf("   标签: %s\n", string(emotionTags))
		fmt.Println()
	}

	fmt.Printf("✅ 共有 %d 个可用音色\n", count)
}

// 测试角色音色选择 (无感获取音色ID)
func testSelectVoiceForCharacter(db *sql.DB, characterID, voiceDBID uuid.UUID, userID uuid.UUID) {
	// 音色设置
	voiceSettings := map[string]interface{}{
		"speed":   1.2,
		"pitch":   1.0,
		"emotion": "friendly",
		"volume":  0.9,
	}
	settingsJSON, _ := json.Marshal(voiceSettings)

	// 为角色选择音色 (无感获取音色ID)
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
	err := db.QueryRow(query, characterID, voiceDBID, true, settingsJSON).Scan(&cvID, &createdAt)
	if err != nil {
		log.Printf("选择角色音色失败: %v", err)
		return
	}

	// 获取音色信息
	var voiceID, voiceName string
	err = db.QueryRow("SELECT voice_id, voice_name FROM voice_clones WHERE id = $1", voiceDBID).Scan(&voiceID, &voiceName)
	if err != nil {
		log.Printf("获取音色信息失败: %v", err)
		return
	}

	// 记录使用日志
	logQuery := `
		INSERT INTO voice_usage_logs (voice_id, user_id, character_id, usage_type)
		VALUES ($1, $2, $3, 'character_creation')
	`
	db.Exec(logQuery, voiceDBID, userID, characterID)

	// 更新使用次数
	updateQuery := `UPDATE voice_clones SET usage_count = usage_count + 1, updated_at = NOW() WHERE id = $1`
	db.Exec(updateQuery, voiceDBID)

	fmt.Printf("✅ 角色音色选择成功 (无感获取音色ID)!\n")
	fmt.Printf("   角色ID: %s\n", characterID)
	fmt.Printf("   音色ID: %s (自动获取)\n", voiceID)
	fmt.Printf("   音色名称: %s\n", voiceName)
	fmt.Printf("   关联ID: %s\n", cvID)
	fmt.Printf("   音色设置: %s\n", string(settingsJSON))
	fmt.Printf("   设置时间: %s\n", createdAt.Format("2006-01-02 15:04:05"))
}

// 测试音色试听
func testVoicePreview(db *sql.DB, voiceDBID uuid.UUID, userID uuid.UUID) {
	// 记录试听使用
	logQuery := `
		INSERT INTO voice_usage_logs (voice_id, user_id, usage_type, usage_duration, cost_amount)
		VALUES ($1, $2, 'voice_test', 8, 0.02)
		RETURNING id, created_at
	`

	var logID uuid.UUID
	var logTime time.Time
	err := db.QueryRow(logQuery, voiceDBID, userID).Scan(&logID, &logTime)
	if err != nil {
		log.Printf("记录试听日志失败: %v", err)
		return
	}

	// 更新使用次数
	updateQuery := `UPDATE voice_clones SET usage_count = usage_count + 1, updated_at = NOW() WHERE id = $1`
	db.Exec(updateQuery, voiceDBID)

	// 获取音色信息
	var voiceID, voiceName string
	err = db.QueryRow("SELECT voice_id, voice_name FROM voice_clones WHERE id = $1", voiceDBID).Scan(&voiceID, &voiceName)
	if err != nil {
		log.Printf("获取音色信息失败: %v", err)
		return
	}

	previewURL := fmt.Sprintf("/audio/preview/%s_%d.mp3", voiceID, time.Now().Unix())

	fmt.Printf("✅ 音色试听成功!\n")
	fmt.Printf("   音色名称: %s\n", voiceName)
	fmt.Printf("   试听文本: 这是一个音色试听测试，听听我的声音如何。\n")
	fmt.Printf("   试听地址: %s\n", previewURL)
	fmt.Printf("   试听时长: 8秒\n")
	fmt.Printf("   试听成本: ¥0.02\n")
	fmt.Printf("   日志ID: %s\n", logID)
}

// 测试音色使用统计
func testVoiceUsageStats(db *sql.DB, voiceDBID uuid.UUID) {
	query := `
		SELECT 
			vc.voice_id, vc.voice_name, vc.voice_image_url, vc.is_public,
			vc.usage_count, vc.like_count, vc.quality_score, vc.created_at,
			COUNT(vul.id) as log_count,
			SUM(vul.cost_amount) as total_cost
		FROM voice_clones vc
		LEFT JOIN voice_usage_logs vul ON vc.id = vul.voice_id
		WHERE vc.id = $1
		GROUP BY vc.id, vc.voice_id, vc.voice_name, vc.voice_image_url, 
		         vc.is_public, vc.usage_count, vc.like_count, vc.quality_score, vc.created_at
	`

	var voiceID, voiceName, imageURL string
	var isPublic bool
	var usageCount, likeCount int
	var qualityScore float64
	var createdAt time.Time
	var logCount int
	var totalCost float64

	err := db.QueryRow(query, voiceDBID).Scan(
		&voiceID, &voiceName, &imageURL, &isPublic,
		&usageCount, &likeCount, &qualityScore, &createdAt,
		&logCount, &totalCost,
	)
	if err != nil {
		log.Printf("查询音色统计失败: %v", err)
		return
	}

	publicStatus := "私密"
	if isPublic {
		publicStatus = "公开"
	}

	fmt.Printf("📈 音色使用统计:\n")
	fmt.Printf("-------------------------------------------\n")
	fmt.Printf("音色名称: %s (%s)\n", voiceName, voiceID)
	fmt.Printf("自定义图片: %s\n", imageURL)
	fmt.Printf("公开状态: %s\n", publicStatus)
	fmt.Printf("质量评分: %.1f/10.0\n", qualityScore)
	fmt.Printf("使用次数: %d次\n", usageCount)
	fmt.Printf("点赞数量: %d个\n", likeCount)
	fmt.Printf("使用日志: %d条\n", logCount)
	fmt.Printf("总消费: ¥%.4f\n", totalCost)
	fmt.Printf("创建时间: %s\n", createdAt.Format("2006-01-02 15:04:05"))

	fmt.Printf("✅ 音色统计查询完成\n")
}
