package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// 音色自定义结构
type VoiceCustom struct {
	ID               uuid.UUID `json:"id" db:"id"`
	VoiceID          string    `json:"voice_id" db:"voice_id"`
	VoiceName        string    `json:"voice_name" db:"voice_name"`           // 用户自定义名称
	VoiceImageURL    string    `json:"voice_image_url" db:"voice_image_url"` // 用户上传图片
	VoiceDescription string    `json:"voice_description" db:"voice_description"`
	AudioSampleURL   string    `json:"audio_sample_url" db:"audio_sample_url"`
	OriginalAudioURL string    `json:"original_audio_url" db:"original_audio_url"`
	
	// 用户自定义权限
	IsPublic        bool      `json:"is_public" db:"is_public"`         // 用户设置公开/私密
	CreatorUserID   uuid.UUID `json:"creator_user_id" db:"creator_user_id"`
	IsOwn           bool      `json:"is_own" db:"is_own"`
	
	// 音色特征 (用户可选择)
	Gender      string          `json:"gender" db:"gender"`
	AgeRange    string          `json:"age_range" db:"age_range"`
	Language    string          `json:"language" db:"language"`
	EmotionTags json.RawMessage `json:"emotion_tags" db:"emotion_tags"`
	
	// 系统信息
	QualityScore  float64 `json:"quality_score" db:"quality_score"`
	CloneStatus   string  `json:"clone_status" db:"clone_status"`
	UsageCount    int     `json:"usage_count" db:"usage_count"`
	LikeCount     int     `json:"like_count" db:"like_count"`
	
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var db *sql.DB

func main() {
	// 初始化数据库连接
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=5201314hdz dbname=yunai sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 初始化Gin路由
	r := gin.Default()

	// 音色自定义管理API
	voiceAPI := r.Group("/api/v1/voice-custom")
	{
		// 创建自定义音色 (用户上传音频+自定义信息)
		voiceAPI.POST("/create", createCustomVoice)
		
		// 更新音色自定义信息 (名称、图片、描述等)
		voiceAPI.PUT("/:voice_id/update", updateCustomVoice)
		
		// 上传/更新音色图片
		voiceAPI.POST("/:voice_id/upload-image", uploadVoiceImage)
		
		// 切换音色公开/私密状态
		voiceAPI.PUT("/:voice_id/privacy", toggleVoicePrivacy)
		
		// 获取用户的音色列表 (可编辑)
		voiceAPI.GET("/my/:user_id", getMyCustomVoices)
		
		// 获取公开音色列表 (所有用户可见)
		voiceAPI.GET("/public", getPublicVoices)
		
		// 获取用户可用音色 (公开+自己的私密)
		voiceAPI.GET("/available/:user_id", getAvailableVoices)
		
		// 删除音色
		voiceAPI.DELETE("/:voice_id", deleteCustomVoice)
		
		// 音色试听
		voiceAPI.POST("/:voice_id/preview", previewVoice)
		
		// 收藏/取消收藏音色
		voiceAPI.POST("/:voice_id/favorite", toggleFavorite)
		
		// 点赞/取消点赞音色
		voiceAPI.POST("/:voice_id/like", toggleLike)
	}

	// 角色音色选择API
	characterAPI := r.Group("/api/v1/character-voice")
	{
		// 为角色选择音色 (从音色列表中选择)
		characterAPI.POST("/select", selectVoiceForCharacter)
		
		// 获取角色当前音色
		characterAPI.GET("/:character_id", getCharacterCurrentVoice)
		
		// 更新角色音色设置
		characterAPI.PUT("/:character_id/settings", updateCharacterVoiceSettings)
	}

	fmt.Println("🎵 YUNAI音色自定义管理系统")
	fmt.Println("===========================================")
	fmt.Println("📡 服务地址: http://localhost:8080")
	fmt.Println("📋 主要功能:")
	fmt.Println("  ✅ 用户自定义音色名称")
	fmt.Println("  ✅ 用户上传音色图片")
	fmt.Println("  ✅ 用户设置公开/私密")
	fmt.Println("  ✅ 音色试听和选择")
	fmt.Println("  ✅ 角色音色自动获取")
	fmt.Println("\n🚀 服务已启动，等待请求...")

	r.Run(":8080")
}

// 创建自定义音色
func createCustomVoice(c *gin.Context) {
	var req struct {
		// 用户自定义信息
		VoiceName        string   `json:"voice_name" binding:"required"`        // 用户自定义名称
		VoiceDescription string   `json:"voice_description"`                    // 用户描述
		VoiceImageURL    string   `json:"voice_image_url"`                      // 用户上传的图片
		IsPublic         bool     `json:"is_public"`                            // 用户设置公开/私密
		
		// 音频文件
		OriginalAudioURL string   `json:"original_audio_url" binding:"required"` // 用户上传的音频
		
		// 用户选择的特征
		Gender       string   `json:"gender"`       // 用户选择性别
		AgeRange     string   `json:"age_range"`    // 用户选择年龄段
		Language     string   `json:"language"`     // 用户选择语言
		EmotionTags  []string `json:"emotion_tags"` // 用户选择情感标签
		
		// 系统信息
		CreatorUserID string `json:"creator_user_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 生成音色ID
	voiceID := fmt.Sprintf("custom_%s_%d", req.CreatorUserID[:8], time.Now().Unix())
	
	// 转换情感标签为JSON
	emotionTagsJSON, _ := json.Marshal(req.EmotionTags)
	
	// 插入音色记录
	query := `
		INSERT INTO voice_clones (
			voice_id, voice_name, voice_description, voice_image_url,
			original_audio_url, is_public, creator_user_id, gender,
			age_range, language, emotion_tags, clone_status, 
			review_status, quality_score
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7::uuid, $8, $9, $10, $11, 
			'processing', 'approved', 0.0
		) RETURNING id, created_at
	`
	
	var newID uuid.UUID
	var createdAt time.Time
	err := db.QueryRow(query,
		voiceID, req.VoiceName, req.VoiceDescription, req.VoiceImageURL,
		req.OriginalAudioURL, req.IsPublic, req.CreatorUserID, req.Gender,
		req.AgeRange, req.Language, emotionTagsJSON,
	).Scan(&newID, &createdAt)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "创建音色失败: " + err.Error(),
		})
		return
	}
	
	// TODO: 这里应该调用SiliconFlow API进行实际的音色克隆
	// 模拟处理完成
	go func() {
		time.Sleep(5 * time.Second) // 模拟处理时间
		
		// 更新状态为完成，并生成试听样本
		sampleURL := fmt.Sprintf("/audio/samples/%s_sample.mp3", voiceID)
		updateQuery := `
			UPDATE voice_clones 
			SET clone_status = 'completed', 
			    audio_sample_url = $1,
			    quality_score = $2,
			    updated_at = NOW()
			WHERE id = $3
		`
		db.Exec(updateQuery, sampleURL, 8.0, newID)
	}()
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "音色创建成功，正在处理中...",
		Data: map[string]interface{}{
			"voice_id":     voiceID,
			"id":           newID,
			"voice_name":   req.VoiceName,
			"is_public":    req.IsPublic,
			"status":       "processing",
			"created_at":   createdAt,
		},
	})
}

// 更新音色自定义信息
func updateCustomVoice(c *gin.Context) {
	voiceID := c.Param("voice_id")
	
	var req struct {
		UserID           string   `json:"user_id" binding:"required"`
		VoiceName        string   `json:"voice_name"`        // 用户可修改名称
		VoiceDescription string   `json:"voice_description"` // 用户可修改描述
		VoiceImageURL    string   `json:"voice_image_url"`   // 用户可更换图片
		Gender           string   `json:"gender"`            // 用户可修改性别
		AgeRange         string   `json:"age_range"`         // 用户可修改年龄段
		Language         string   `json:"language"`          // 用户可修改语言
		EmotionTags      []string `json:"emotion_tags"`      // 用户可修改标签
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 转换情感标签为JSON
	emotionTagsJSON, _ := json.Marshal(req.EmotionTags)
	
	// 更新音色信息
	query := `
		UPDATE voice_clones 
		SET voice_name = COALESCE(NULLIF($1, ''), voice_name),
		    voice_description = COALESCE(NULLIF($2, ''), voice_description),
		    voice_image_url = COALESCE(NULLIF($3, ''), voice_image_url),
		    gender = COALESCE(NULLIF($4, ''), gender),
		    age_range = COALESCE(NULLIF($5, ''), age_range),
		    language = COALESCE(NULLIF($6, ''), language),
		    emotion_tags = COALESCE($7, emotion_tags),
		    updated_at = NOW()
		WHERE voice_id = $8 AND creator_user_id = $9::uuid
	`
	
	result, err := db.Exec(query, 
		req.VoiceName, req.VoiceDescription, req.VoiceImageURL,
		req.Gender, req.AgeRange, req.Language, emotionTagsJSON,
		voiceID, req.UserID)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "更新音色失败: " + err.Error(),
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, APIResponse{
			Code:    404,
			Message: "音色不存在或无权限操作",
		})
		return
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "音色信息更新成功",
		Data: map[string]interface{}{
			"voice_id":   voiceID,
			"updated_at": time.Now(),
		},
	})
}

// 切换音色公开/私密状态
func toggleVoicePrivacy(c *gin.Context) {
	voiceID := c.Param("voice_id")
	
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		IsPublic bool   `json:"is_public"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 更新音色公开状态
	query := `
		UPDATE voice_clones 
		SET is_public = $1, updated_at = NOW()
		WHERE voice_id = $2 AND creator_user_id = $3::uuid
	`
	
	result, err := db.Exec(query, req.IsPublic, voiceID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "更新音色状态失败: " + err.Error(),
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, APIResponse{
			Code:    404,
			Message: "音色不存在或无权限操作",
		})
		return
	}
	
	statusText := "私密"
	if req.IsPublic {
		statusText = "公开"
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: fmt.Sprintf("音色已设置为%s", statusText),
		Data: map[string]interface{}{
			"voice_id":  voiceID,
			"is_public": req.IsPublic,
			"status":    statusText,
		},
	})
}

// 获取用户的音色列表 (可编辑)
func getMyCustomVoices(c *gin.Context) {
	userID := c.Param("user_id")
	
	query := `
		SELECT 
			id, voice_id, voice_name, voice_description, voice_image_url,
			audio_sample_url, original_audio_url, is_public, gender,
			age_range, language, emotion_tags, quality_score, clone_status,
			usage_count, like_count, created_at, updated_at
		FROM voice_clones
		WHERE creator_user_id = $1::uuid
		ORDER BY created_at DESC
	`
	
	rows, err := db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询音色失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()
	
	var voices []VoiceCustom
	for rows.Next() {
		var voice VoiceCustom
		err := rows.Scan(
			&voice.ID, &voice.VoiceID, &voice.VoiceName, &voice.VoiceDescription,
			&voice.VoiceImageURL, &voice.AudioSampleURL, &voice.OriginalAudioURL,
			&voice.IsPublic, &voice.Gender, &voice.AgeRange, &voice.Language,
			&voice.EmotionTags, &voice.QualityScore, &voice.CloneStatus,
			&voice.UsageCount, &voice.LikeCount, &voice.CreatedAt, &voice.UpdatedAt,
		)
		if err != nil {
			continue
		}
		voice.IsOwn = true // 都是用户自己的音色
		voices = append(voices, voice)
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取我的音色列表成功",
		Data: map[string]interface{}{
			"voices": voices,
			"total":  len(voices),
		},
	})
}

// 获取公开音色列表
func getPublicVoices(c *gin.Context) {
	query := `
		SELECT 
			id, voice_id, voice_name, voice_description, voice_image_url,
			audio_sample_url, creator_user_id, gender, age_range, language,
			emotion_tags, quality_score, usage_count, like_count, created_at
		FROM voice_clones
		WHERE is_public = true AND clone_status = 'completed'
		ORDER BY quality_score DESC, usage_count DESC, created_at DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询公开音色失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()
	
	var voices []VoiceCustom
	for rows.Next() {
		var voice VoiceCustom
		err := rows.Scan(
			&voice.ID, &voice.VoiceID, &voice.VoiceName, &voice.VoiceDescription,
			&voice.VoiceImageURL, &voice.AudioSampleURL, &voice.CreatorUserID,
			&voice.Gender, &voice.AgeRange, &voice.Language, &voice.EmotionTags,
			&voice.QualityScore, &voice.UsageCount, &voice.LikeCount, &voice.CreatedAt,
		)
		if err != nil {
			continue
		}
		voice.IsPublic = true
		voice.IsOwn = false
		voices = append(voices, voice)
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取公开音色列表成功",
		Data: map[string]interface{}{
			"voices": voices,
			"total":  len(voices),
		},
	})
}

// 获取用户可用音色 (公开+自己的私密)
func getAvailableVoices(c *gin.Context) {
	userID := c.Param("user_id")
	
	query := `
		SELECT 
			id, voice_id, voice_name, voice_description, voice_image_url,
			audio_sample_url, is_public, creator_user_id,
			(creator_user_id = $1::uuid) as is_own,
			gender, age_range, language, emotion_tags, quality_score,
			usage_count, like_count, created_at
		FROM voice_clones
		WHERE (is_public = true OR creator_user_id = $1::uuid)
		AND clone_status = 'completed'
		ORDER BY (creator_user_id = $1::uuid) DESC, quality_score DESC, created_at DESC
	`
	
	rows, err := db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询可用音色失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()
	
	var voices []VoiceCustom
	for rows.Next() {
		var voice VoiceCustom
		err := rows.Scan(
			&voice.ID, &voice.VoiceID, &voice.VoiceName, &voice.VoiceDescription,
			&voice.VoiceImageURL, &voice.AudioSampleURL, &voice.IsPublic,
			&voice.CreatorUserID, &voice.IsOwn, &voice.Gender, &voice.AgeRange,
			&voice.Language, &voice.EmotionTags, &voice.QualityScore,
			&voice.UsageCount, &voice.LikeCount, &voice.CreatedAt,
		)
		if err != nil {
			continue
		}
		voices = append(voices, voice)
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取可用音色列表成功",
		Data: map[string]interface{}{
			"voices": voices,
			"total":  len(voices),
		},
	})
}

// 为角色选择音色 (无感获取音色ID)
func selectVoiceForCharacter(c *gin.Context) {
	var req struct {
		UserID      string                 `json:"user_id" binding:"required"`
		CharacterID string                 `json:"character_id" binding:"required"`
		VoiceID     string                 `json:"voice_id" binding:"required"`     // 从音色列表中选择的音色ID
		IsPrimary   bool                   `json:"is_primary"`
		Settings    map[string]interface{} `json:"settings"` // 音色参数设置
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 检查音色是否可用 (公开音色或用户自己的音色)
	var voiceDBID uuid.UUID
	var voiceName string
	checkQuery := `
		SELECT id, voice_name FROM voice_clones
		WHERE voice_id = $1 
		AND (is_public = true OR creator_user_id = $2::uuid)
		AND clone_status = 'completed'
	`
	
	err := db.QueryRow(checkQuery, req.VoiceID, req.UserID).Scan(&voiceDBID, &voiceName)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Code:    404,
			Message: "音色不存在或无权限使用",
		})
		return
	}
	
	// 转换设置为JSON
	settingsJSON, _ := json.Marshal(req.Settings)
	
	// 为角色设置音色 (无感获取音色ID并自动填入)
	insertQuery := `
		INSERT INTO character_voices (character_id, voice_id, is_primary, voice_settings)
		VALUES ($1::uuid, $2, $3, $4)
		ON CONFLICT (character_id, voice_id) DO UPDATE SET
			is_primary = EXCLUDED.is_primary,
			voice_settings = EXCLUDED.voice_settings,
			updated_at = NOW()
		RETURNING id, created_at
	`
	
	var cvID uuid.UUID
	var createdAt time.Time
	err = db.QueryRow(insertQuery, req.CharacterID, voiceDBID, req.IsPrimary, settingsJSON).Scan(&cvID, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "设置角色音色失败: " + err.Error(),
		})
		return
	}
	
	// 记录使用日志
	logQuery := `
		INSERT INTO voice_usage_logs (voice_id, user_id, character_id, usage_type)
		VALUES ($1, $2::uuid, $3::uuid, 'character_creation')
	`
	db.Exec(logQuery, voiceDBID, req.UserID, req.CharacterID)
	
	// 更新音色使用次数
	updateQuery := `UPDATE voice_clones SET usage_count = usage_count + 1, updated_at = NOW() WHERE id = $1`
	db.Exec(updateQuery, voiceDBID)
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "角色音色设置成功",
		Data: map[string]interface{}{
			"character_id":   req.CharacterID,
			"voice_id":       req.VoiceID,
			"voice_name":     voiceName,
			"is_primary":     req.IsPrimary,
			"settings":       req.Settings,
			"relation_id":    cvID,
			"created_at":     createdAt,
		},
	})
}

// 音色试听
func previewVoice(c *gin.Context) {
	voiceID := c.Param("voice_id")
	
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Text   string `json:"text" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 检查音色是否可用
	var voiceDBID uuid.UUID
	var voiceName string
	checkQuery := `
		SELECT id, voice_name FROM voice_clones
		WHERE voice_id = $1 
		AND (is_public = true OR creator_user_id = $2::uuid)
		AND clone_status = 'completed'
	`
	
	err := db.QueryRow(checkQuery, voiceID, req.UserID).Scan(&voiceDBID, &voiceName)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Code:    404,
			Message: "音色不存在或无权限试听",
		})
		return
	}
	
	// 记录试听日志
	logQuery := `
		INSERT INTO voice_usage_logs (voice_id, user_id, usage_type, usage_duration, cost_amount)
		VALUES ($1, $2::uuid, 'voice_test', 5, 0.01)
	`
	db.Exec(logQuery, voiceDBID, req.UserID)
	
	// 更新使用次数
	updateQuery := `UPDATE voice_clones SET usage_count = usage_count + 1, updated_at = NOW() WHERE id = $1`
	db.Exec(updateQuery, voiceDBID)
	
	// TODO: 这里应该调用SiliconFlow TTS API生成试听音频
	previewURL := fmt.Sprintf("/audio/preview/%s_%d.mp3", voiceID, time.Now().Unix())
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "音色试听生成成功",
		Data: map[string]interface{}{
			"voice_id":     voiceID,
			"voice_name":   voiceName,
			"preview_url":  previewURL,
			"text":         req.Text,
			"duration":     5,
		},
	})
}

// 其他API函数占位符
func uploadVoiceImage(c *gin.Context)           { /* TODO: 实现图片上传 */ }
func deleteCustomVoice(c *gin.Context)          { /* TODO: 实现删除音色 */ }
func toggleFavorite(c *gin.Context)             { /* TODO: 实现收藏音色 */ }
func toggleLike(c *gin.Context)                 { /* TODO: 实现点赞音色 */ }
func getCharacterCurrentVoice(c *gin.Context)   { /* TODO: 实现获取角色当前音色 */ }
func updateCharacterVoiceSettings(c *gin.Context) { /* TODO: 实现更新角色音色设置 */ }
