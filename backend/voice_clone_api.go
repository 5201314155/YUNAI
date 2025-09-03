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

// 音色结构体
type VoiceClone struct {
	ID               uuid.UUID `json:"id" db:"id"`
	VoiceID          string    `json:"voice_id" db:"voice_id"`
	VoiceName        string    `json:"voice_name" db:"voice_name"`
	VoiceImageURL    string    `json:"voice_image_url" db:"voice_image_url"`
	VoiceDescription string    `json:"voice_description" db:"voice_description"`
	AudioSampleURL   string    `json:"audio_sample_url" db:"audio_sample_url"`
	OriginalAudioURL string    `json:"original_audio_url" db:"original_audio_url"`
	
	// 权限控制
	IsPublic        bool      `json:"is_public" db:"is_public"`
	CreatorUserID   uuid.UUID `json:"creator_user_id" db:"creator_user_id"`
	IsOwn           bool      `json:"is_own" db:"is_own"` // 查询时使用
	
	// 音色特征
	Gender      string          `json:"gender" db:"gender"`
	AgeRange    string          `json:"age_range" db:"age_range"`
	Language    string          `json:"language" db:"language"`
	Accent      string          `json:"accent" db:"accent"`
	EmotionTags json.RawMessage `json:"emotion_tags" db:"emotion_tags"`
	
	// 技术参数
	ModelProvider string  `json:"model_provider" db:"model_provider"`
	CloneModel    string  `json:"clone_model" db:"clone_model"`
	QualityScore  float64 `json:"quality_score" db:"quality_score"`
	CloneStatus   string  `json:"clone_status" db:"clone_status"`
	
	// 使用统计
	UsageCount    int `json:"usage_count" db:"usage_count"`
	LikeCount     int `json:"like_count" db:"like_count"`
	DownloadCount int `json:"download_count" db:"download_count"`
	
	// 审核状态
	ReviewStatus string     `json:"review_status" db:"review_status"`
	ReviewReason string     `json:"review_reason" db:"review_reason"`
	ReviewedBy   *uuid.UUID `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt   *time.Time `json:"reviewed_at" db:"reviewed_at"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// 音色使用记录
type VoiceUsageLog struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	VoiceID          uuid.UUID  `json:"voice_id" db:"voice_id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	CharacterID      *uuid.UUID `json:"character_id" db:"character_id"`
	UsageType        string     `json:"usage_type" db:"usage_type"`
	UsageDuration    int        `json:"usage_duration" db:"usage_duration"`
	GeneratedAudioURL string    `json:"generated_audio_url" db:"generated_audio_url"`
	CostAmount       float64    `json:"cost_amount" db:"cost_amount"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// 角色音色关联
type CharacterVoice struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	CharacterID   uuid.UUID       `json:"character_id" db:"character_id"`
	VoiceID       uuid.UUID       `json:"voice_id" db:"voice_id"`
	IsPrimary     bool            `json:"is_primary" db:"is_primary"`
	VoiceSettings json.RawMessage `json:"voice_settings" db:"voice_settings"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 数据库连接
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

	// 音色管理API路由
	voiceAPI := r.Group("/api/v1/voices")
	{
		// 获取可用音色列表 (公开音色 + 用户自己的私密音色)
		voiceAPI.GET("/available/:user_id", getAvailableVoices)
		
		// 获取用户自己的音色列表
		voiceAPI.GET("/my/:user_id", getMyVoices)
		
		// 上传音色 (克隆音色)
		voiceAPI.POST("/clone", cloneVoice)
		
		// 更新音色信息
		voiceAPI.PUT("/:voice_id", updateVoice)
		
		// 切换音色公开/私密状态
		voiceAPI.PUT("/:voice_id/privacy", toggleVoicePrivacy)
		
		// 删除音色
		voiceAPI.DELETE("/:voice_id", deleteVoice)
		
		// 试听音色
		voiceAPI.POST("/:voice_id/preview", previewVoice)
		
		// 使用音色 (为角色设置音色)
		voiceAPI.POST("/:voice_id/use", useVoiceForCharacter)
		
		// 收藏/取消收藏音色
		voiceAPI.POST("/:voice_id/favorite", toggleVoiceFavorite)
		
		// 评价音色
		voiceAPI.POST("/:voice_id/review", reviewVoice)
		
		// 获取音色使用统计
		voiceAPI.GET("/:voice_id/stats", getVoiceStats)
	}

	// 角色音色管理API
	characterAPI := r.Group("/api/v1/characters")
	{
		// 获取角色的音色设置
		characterAPI.GET("/:character_id/voices", getCharacterVoices)
		
		// 为角色设置音色
		characterAPI.POST("/:character_id/voices", setCharacterVoice)
		
		// 更新角色音色设置
		characterAPI.PUT("/:character_id/voices/:voice_id", updateCharacterVoice)
		
		// 移除角色音色
		characterAPI.DELETE("/:character_id/voices/:voice_id", removeCharacterVoice)
	}

	// 管理员API
	adminAPI := r.Group("/api/v1/admin/voices")
	{
		// 获取待审核音色列表
		adminAPI.GET("/pending", getPendingVoices)
		
		// 审核音色
		adminAPI.POST("/:voice_id/review", reviewVoiceAdmin)
		
		// 获取所有音色统计
		adminAPI.GET("/stats", getAllVoicesStats)
	}

	fmt.Println("🎵 YUNAI音色管理API服务启动")
	fmt.Println("===========================================")
	fmt.Println("📡 服务地址: http://localhost:8080")
	fmt.Println("📋 API文档:")
	fmt.Println("  GET  /api/v1/voices/available/:user_id - 获取可用音色")
	fmt.Println("  GET  /api/v1/voices/my/:user_id - 获取我的音色")
	fmt.Println("  POST /api/v1/voices/clone - 克隆音色")
	fmt.Println("  PUT  /api/v1/voices/:voice_id - 更新音色")
	fmt.Println("  PUT  /api/v1/voices/:voice_id/privacy - 切换公开状态")
	fmt.Println("  POST /api/v1/voices/:voice_id/preview - 试听音色")
	fmt.Println("  POST /api/v1/voices/:voice_id/use - 使用音色")
	fmt.Println("  POST /api/v1/characters/:character_id/voices - 设置角色音色")
	fmt.Println("\n🚀 服务已启动，等待请求...")

	// 启动服务器
	r.Run(":8080")
}

// 获取可用音色列表 (公开音色 + 用户自己的私密音色)
func getAvailableVoices(c *gin.Context) {
	userID := c.Param("user_id")
	
	query := `
		SELECT 
			id, voice_id, voice_name, voice_image_url, voice_description, 
			audio_sample_url, is_public, creator_user_id,
			(creator_user_id = $1::uuid) as is_own,
			quality_score, usage_count, like_count, gender, age_range, 
			emotion_tags, created_at
		FROM voice_clones
		WHERE (is_public = true AND review_status = 'approved') 
		   OR creator_user_id = $1::uuid
		AND clone_status = 'completed'
		ORDER BY is_public DESC, quality_score DESC, created_at DESC
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
	
	var voices []VoiceClone
	for rows.Next() {
		var voice VoiceClone
		err := rows.Scan(
			&voice.ID, &voice.VoiceID, &voice.VoiceName, &voice.VoiceImageURL,
			&voice.VoiceDescription, &voice.AudioSampleURL, &voice.IsPublic,
			&voice.CreatorUserID, &voice.IsOwn, &voice.QualityScore,
			&voice.UsageCount, &voice.LikeCount, &voice.Gender, &voice.AgeRange,
			&voice.EmotionTags, &voice.CreatedAt,
		)
		if err != nil {
			continue
		}
		voices = append(voices, voice)
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取音色列表成功",
		Data: map[string]interface{}{
			"voices": voices,
			"total":  len(voices),
		},
	})
}

// 获取用户自己的音色列表
func getMyVoices(c *gin.Context) {
	userID := c.Param("user_id")
	
	query := `
		SELECT 
			id, voice_id, voice_name, voice_image_url, voice_description,
			audio_sample_url, original_audio_url, is_public, creator_user_id,
			gender, age_range, language, accent, emotion_tags,
			model_provider, clone_model, quality_score, clone_status,
			usage_count, like_count, download_count,
			review_status, review_reason, reviewed_by, reviewed_at,
			created_at, updated_at
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
	
	var voices []VoiceClone
	for rows.Next() {
		var voice VoiceClone
		err := rows.Scan(
			&voice.ID, &voice.VoiceID, &voice.VoiceName, &voice.VoiceImageURL,
			&voice.VoiceDescription, &voice.AudioSampleURL, &voice.OriginalAudioURL,
			&voice.IsPublic, &voice.CreatorUserID, &voice.Gender, &voice.AgeRange,
			&voice.Language, &voice.Accent, &voice.EmotionTags, &voice.ModelProvider,
			&voice.CloneModel, &voice.QualityScore, &voice.CloneStatus,
			&voice.UsageCount, &voice.LikeCount, &voice.DownloadCount,
			&voice.ReviewStatus, &voice.ReviewReason, &voice.ReviewedBy,
			&voice.ReviewedAt, &voice.CreatedAt, &voice.UpdatedAt,
		)
		if err != nil {
			continue
		}
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

// 克隆音色 (上传音色)
func cloneVoice(c *gin.Context) {
	var req struct {
		VoiceName        string   `json:"voice_name" binding:"required"`
		VoiceDescription string   `json:"voice_description"`
		VoiceImageURL    string   `json:"voice_image_url"`
		OriginalAudioURL string   `json:"original_audio_url" binding:"required"`
		IsPublic         bool     `json:"is_public"`
		CreatorUserID    string   `json:"creator_user_id" binding:"required"`
		Gender           string   `json:"gender"`
		AgeRange         string   `json:"age_range"`
		Language         string   `json:"language"`
		EmotionTags      []string `json:"emotion_tags"`
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
			age_range, language, emotion_tags, clone_status, review_status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7::uuid, $8, $9, $10, $11, 'processing', 
			CASE WHEN $6 THEN 'pending' ELSE 'approved' END
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
	// 现在先返回成功响应
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "音色克隆请求已提交，正在处理中...",
		Data: map[string]interface{}{
			"voice_id":   voiceID,
			"id":         newID,
			"status":     "processing",
			"created_at": createdAt,
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
		SET is_public = $1, 
		    review_status = CASE WHEN $1 THEN 'pending' ELSE 'approved' END,
		    updated_at = NOW()
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
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "音色状态更新成功",
		Data: map[string]interface{}{
			"voice_id":  voiceID,
			"is_public": req.IsPublic,
			"status":    "success",
		},
	})
}

// 试听音色
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
	
	// 记录使用日志
	logQuery := `
		INSERT INTO voice_usage_logs (voice_id, user_id, usage_type, usage_duration, cost_amount)
		SELECT id, $2::uuid, 'voice_test', 5, 0.01
		FROM voice_clones WHERE voice_id = $1
	`
	
	_, err := db.Exec(logQuery, voiceID, req.UserID)
	if err != nil {
		log.Printf("记录使用日志失败: %v", err)
	}
	
	// TODO: 这里应该调用SiliconFlow TTS API生成试听音频
	// 现在先返回模拟的音频URL
	
	previewURL := fmt.Sprintf("/audio/preview/%s_%d.mp3", voiceID, time.Now().Unix())
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "音色试听生成成功",
		Data: map[string]interface{}{
			"voice_id":    voiceID,
			"preview_url": previewURL,
			"text":        req.Text,
			"duration":    5,
		},
	})
}

// 使用音色 (为角色设置音色)
func useVoiceForCharacter(c *gin.Context) {
	voiceID := c.Param("voice_id")
	
	var req struct {
		UserID      string                 `json:"user_id" binding:"required"`
		CharacterID string                 `json:"character_id" binding:"required"`
		IsPrimary   bool                   `json:"is_primary"`
		Settings    map[string]interface{} `json:"settings"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 转换设置为JSON
	settingsJSON, _ := json.Marshal(req.Settings)
	
	// 插入或更新角色音色关联
	query := `
		INSERT INTO character_voices (character_id, voice_id, is_primary, voice_settings)
		SELECT $1::uuid, vc.id, $2, $3
		FROM voice_clones vc WHERE vc.voice_id = $4
		ON CONFLICT (character_id, voice_id) DO UPDATE SET
			is_primary = EXCLUDED.is_primary,
			voice_settings = EXCLUDED.voice_settings,
			updated_at = NOW()
	`
	
	_, err := db.Exec(query, req.CharacterID, req.IsPrimary, settingsJSON, voiceID)
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
		SELECT id, $2::uuid, $3::uuid, 'character_creation'
		FROM voice_clones WHERE voice_id = $1
	`
	
	_, err = db.Exec(logQuery, voiceID, req.UserID, req.CharacterID)
	if err != nil {
		log.Printf("记录使用日志失败: %v", err)
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "角色音色设置成功",
		Data: map[string]interface{}{
			"voice_id":     voiceID,
			"character_id": req.CharacterID,
			"is_primary":   req.IsPrimary,
			"settings":     req.Settings,
		},
	})
}

// 获取角色的音色设置
func getCharacterVoices(c *gin.Context) {
	characterID := c.Param("character_id")
	
	query := `
		SELECT 
			cv.id, cv.character_id, cv.voice_id, cv.is_primary, cv.voice_settings,
			vc.voice_name, vc.voice_image_url, vc.audio_sample_url,
			cv.created_at, cv.updated_at
		FROM character_voices cv
		JOIN voice_clones vc ON cv.voice_id = vc.id
		WHERE cv.character_id = $1::uuid
		ORDER BY cv.is_primary DESC, cv.created_at DESC
	`
	
	rows, err := db.Query(query, characterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询角色音色失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()
	
	var characterVoices []map[string]interface{}
	for rows.Next() {
		var cv CharacterVoice
		var voiceName, voiceImageURL, audioSampleURL string
		
		err := rows.Scan(
			&cv.ID, &cv.CharacterID, &cv.VoiceID, &cv.IsPrimary, &cv.VoiceSettings,
			&voiceName, &voiceImageURL, &audioSampleURL,
			&cv.CreatedAt, &cv.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		characterVoices = append(characterVoices, map[string]interface{}{
			"id":                cv.ID,
			"character_id":      cv.CharacterID,
			"voice_id":          cv.VoiceID,
			"is_primary":        cv.IsPrimary,
			"voice_settings":    cv.VoiceSettings,
			"voice_name":        voiceName,
			"voice_image_url":   voiceImageURL,
			"audio_sample_url":  audioSampleURL,
			"created_at":        cv.CreatedAt,
			"updated_at":        cv.UpdatedAt,
		})
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取角色音色成功",
		Data: map[string]interface{}{
			"character_id": characterID,
			"voices":       characterVoices,
			"total":        len(characterVoices),
		},
	})
}

// 其他API函数的占位符...
func updateVoice(c *gin.Context)         { /* TODO: 实现更新音色信息 */ }
func deleteVoice(c *gin.Context)         { /* TODO: 实现删除音色 */ }
func toggleVoiceFavorite(c *gin.Context) { /* TODO: 实现收藏音色 */ }
func reviewVoice(c *gin.Context)         { /* TODO: 实现评价音色 */ }
func getVoiceStats(c *gin.Context)       { /* TODO: 实现获取音色统计 */ }
func setCharacterVoice(c *gin.Context)   { /* TODO: 实现设置角色音色 */ }
func updateCharacterVoice(c *gin.Context) { /* TODO: 实现更新角色音色 */ }
func removeCharacterVoice(c *gin.Context) { /* TODO: 实现移除角色音色 */ }
func getPendingVoices(c *gin.Context)    { /* TODO: 实现获取待审核音色 */ }
func reviewVoiceAdmin(c *gin.Context)    { /* TODO: 实现管理员审核音色 */ }
func getAllVoicesStats(c *gin.Context)   { /* TODO: 实现获取所有音色统计 */ }
