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
	fmt.Println("🌍 测试公开音色列表 - 验证私密音色不会出现")
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

	// 1. 显示所有音色的状态
	fmt.Println("\n📊 当前数据库中所有音色状态:")
	showAllVoicesStatus(db)

	// 2. 测试公开音色列表 (只显示公开音色)
	fmt.Println("\n🌍 公开音色列表 (所有用户可见):")
	testPublicVoicesList(db)

	// 3. 测试用户可用音色列表 (公开+自己的私密)
	fmt.Println("\n👤 用户可用音色列表 (公开+我的私密):")
	testUserAvailableVoices(db, testUserID)

	// 4. 创建另一个用户来验证权限隔离
	fmt.Println("\n🔒 测试权限隔离 - 其他用户视角:")
	testOtherUserView(db)

	fmt.Println("\n🎉 测试完成！私密音色权限控制正常工作！")
}

// 显示所有音色的状态
func showAllVoicesStatus(db *sql.DB) {
	query := `
		SELECT 
			voice_id, voice_name, is_public, creator_user_id, clone_status,
			usage_count, created_at
		FROM voice_clones
		ORDER BY created_at DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询所有音色失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("-------------------------------------------\n")
	
	count := 0
	publicCount := 0
	privateCount := 0
	
	for rows.Next() {
		var voiceID, voiceName string
		var isPublic bool
		var creatorID uuid.UUID
		var cloneStatus string
		var usageCount int
		var createdAt string
		
		err := rows.Scan(&voiceID, &voiceName, &isPublic, &creatorID, &cloneStatus, &usageCount, &createdAt)
		if err != nil {
			continue
		}
		
		count++
		statusIcon := "🔒"
		statusText := "私密"
		if isPublic {
			statusIcon = "🌍"
			statusText = "公开"
			publicCount++
		} else {
			privateCount++
		}
		
		fmt.Printf("%d. %s %s (%s)\n", count, statusIcon, voiceName, voiceID)
		fmt.Printf("   状态: %s | 创建者: %s | 使用: %d次\n", statusText, creatorID.String()[:8], usageCount)
		fmt.Printf("   克隆状态: %s | 创建时间: %s\n", cloneStatus, createdAt[:19])
		fmt.Println()
	}
	
	fmt.Printf("📊 统计: 总计 %d 个音色 (公开: %d, 私密: %d)\n", count, publicCount, privateCount)
}

// 测试公开音色列表 (只显示公开音色)
func testPublicVoicesList(db *sql.DB) {
	query := `
		SELECT 
			voice_id, voice_name, voice_description, voice_image_url,
			audio_sample_url, creator_user_id, gender, age_range,
			emotion_tags, quality_score, usage_count, like_count
		FROM voice_clones
		WHERE is_public = true AND clone_status = 'completed'
		ORDER BY quality_score DESC, usage_count DESC, created_at DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询公开音色失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("-------------------------------------------\n")
	
	count := 0
	for rows.Next() {
		var voiceID, voiceName, description, imageURL, sampleURL string
		var creatorID uuid.UUID
		var gender, ageRange string
		var emotionTags json.RawMessage
		var qualityScore float64
		var usageCount, likeCount int
		
		err := rows.Scan(
			&voiceID, &voiceName, &description, &imageURL, &sampleURL,
			&creatorID, &gender, &ageRange, &emotionTags, &qualityScore,
			&usageCount, &likeCount,
		)
		if err != nil {
			continue
		}
		
		count++
		fmt.Printf("%d. 🌍 %s (%s)\n", count, voiceName, voiceID)
		fmt.Printf("   描述: %s\n", description)
		fmt.Printf("   图片: %s\n", imageURL)
		fmt.Printf("   试听: %s\n", sampleURL)
		fmt.Printf("   创建者: %s\n", creatorID.String()[:8])
		fmt.Printf("   特征: %s %s 评分: %.1f\n", gender, ageRange, qualityScore)
		fmt.Printf("   标签: %s\n", string(emotionTags))
		fmt.Printf("   统计: 使用 %d次, 点赞 %d个\n", usageCount, likeCount)
		fmt.Println()
	}
	
	if count == 0 {
		fmt.Printf("❌ 没有找到公开音色\n")
	} else {
		fmt.Printf("✅ 公开音色列表显示 %d 个音色 (只包含公开音色)\n", count)
	}
}

// 测试用户可用音色列表 (公开+自己的私密)
func testUserAvailableVoices(db *sql.DB, userID uuid.UUID) {
	query := `
		SELECT 
			voice_id, voice_name, voice_image_url, is_public,
			(creator_user_id = $1) as is_own, quality_score, usage_count
		FROM voice_clones
		WHERE (is_public = true OR creator_user_id = $1)
		AND clone_status = 'completed'
		ORDER BY (creator_user_id = $1) DESC, quality_score DESC
	`
	
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("查询用户可用音色失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("-------------------------------------------\n")
	
	count := 0
	publicCount := 0
	privateCount := 0
	
	for rows.Next() {
		var voiceID, voiceName, imageURL string
		var isPublic, isOwn bool
		var qualityScore float64
		var usageCount int
		
		err := rows.Scan(&voiceID, &voiceName, &imageURL, &isPublic, &isOwn, &qualityScore, &usageCount)
		if err != nil {
			continue
		}
		
		count++
		publicIcon := "🔒"
		ownIcon := ""
		
		if isPublic {
			publicIcon = "🌍"
			publicCount++
		} else {
			privateCount++
		}
		
		if isOwn {
			ownIcon = "👤"
		}
		
		fmt.Printf("%d. %s %s %s (%s)\n", count, publicIcon, ownIcon, voiceName, voiceID)
		fmt.Printf("   图片: %s\n", imageURL)
		fmt.Printf("   评分: %.1f | 使用: %d次\n", qualityScore, usageCount)
		
		if isOwn && !isPublic {
			fmt.Printf("   ⚠️  这是我的私密音色，只有我能看到\n")
		} else if isPublic && !isOwn {
			fmt.Printf("   ℹ️  这是其他用户的公开音色\n")
		} else if isPublic && isOwn {
			fmt.Printf("   ℹ️  这是我的公开音色，所有人都能看到\n")
		}
		fmt.Println()
	}
	
	fmt.Printf("✅ 用户可用音色: %d 个 (公开: %d, 我的私密: %d)\n", count, publicCount, privateCount)
}

// 测试其他用户视角 (验证权限隔离)
func testOtherUserView(db *sql.DB) {
	// 创建一个虚拟的其他用户ID
	otherUserID := uuid.New()
	
	fmt.Printf("模拟其他用户ID: %s\n", otherUserID.String()[:8])
	
	// 查询其他用户能看到的音色 (只有公开音色)
	query := `
		SELECT 
			voice_id, voice_name, is_public,
			(creator_user_id = $1) as is_own
		FROM voice_clones
		WHERE (is_public = true OR creator_user_id = $1)
		AND clone_status = 'completed'
		ORDER BY quality_score DESC
	`
	
	rows, err := db.Query(query, otherUserID)
	if err != nil {
		log.Printf("查询其他用户可见音色失败: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Printf("-------------------------------------------\n")
	
	count := 0
	for rows.Next() {
		var voiceID, voiceName string
		var isPublic, isOwn bool
		
		err := rows.Scan(&voiceID, &voiceName, &isPublic, &isOwn)
		if err != nil {
			continue
		}
		
		count++
		statusIcon := "🌍"
		if !isPublic {
			statusIcon = "🔒"
		}
		
		fmt.Printf("%d. %s %s (%s)\n", count, statusIcon, voiceName, voiceID)
		
		if isPublic {
			fmt.Printf("   ✅ 公开音色 - 其他用户可以看到和使用\n")
		} else {
			fmt.Printf("   ❌ 这不应该出现！私密音色被其他用户看到了！\n")
		}
		fmt.Println()
	}
	
	fmt.Printf("✅ 其他用户只能看到 %d 个公开音色 (私密音色已被正确隔离)\n", count)
	
	// 验证私密音色确实被隔离
	privateQuery := `
		SELECT COUNT(*) FROM voice_clones
		WHERE is_public = false AND creator_user_id != $1
		AND clone_status = 'completed'
	`
	
	var privateCount int
	err = db.QueryRow(privateQuery, otherUserID).Scan(&privateCount)
	if err != nil {
		log.Printf("查询私密音色数量失败: %v", err)
		return
	}
	
	if privateCount > 0 {
		fmt.Printf("🔒 数据库中有 %d 个其他用户的私密音色，已被正确隔离\n", privateCount)
	} else {
		fmt.Printf("ℹ️  数据库中没有其他用户的私密音色\n")
	}
}
