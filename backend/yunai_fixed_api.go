package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// 数据库连接
var db *sql.DB

// SiliconFlow配置
const (
	SILICONFLOW_API_KEY = "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"
	SILICONFLOW_BASE_URL = "https://api.siliconflow.cn/v1"
)

// API请求/响应结构
type ChatRequest struct {
	ModelID     string    `json:"model_id" binding:"required"`
	Messages    []Message `json:"messages" binding:"required"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

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

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 管理员API - 显示真实模型名称
		admin := api.Group("/admin")
		{
			admin.GET("/models", getAdminModels)           // 管理员看真实模型名称
		}
		
		// 用户API - 显示自定义友好名称
		api.GET("/models", getUserModels)                 // 用户看自定义名称
		
		// 对话API - 内部转换为真实模型名称调用
		api.POST("/chat/completions", chatCompletions)
		
		// 系统信息API
		api.GET("/system/info", getSystemInfo)
	}

	fmt.Println("🚀 YUNAI双API系统")
	fmt.Println("===========================================")
	fmt.Println("📡 服务地址: http://localhost:8080")
	fmt.Println("📋 API端点:")
	fmt.Println("  GET  /api/v1/admin/models - 管理员模型列表 (真实名称)")
	fmt.Println("  GET  /api/v1/models - 用户模型列表 (友好名称)")
	fmt.Println("  POST /api/v1/chat/completions - 对话接口 (自动转换)")
	fmt.Println("  GET  /api/v1/system/info - 系统信息")
	fmt.Println("\n🎛️ 管理员看到真实模型ID")
	fmt.Println("👤 用户看到友好名称")
	fmt.Println("🤖 对话时自动转换为真实模型调用")
	fmt.Println("\n✅ 服务已启动，等待请求...")

	// 启动服务器
	r.Run(":8080")
}

// 管理员获取模型列表 (显示真实模型名称)
func getAdminModels(c *gin.Context) {
	modelType := c.Query("type")
	isActive := c.DefaultQuery("active", "true")
	
	whereClause := "WHERE is_active = $1"
	args := []interface{}{isActive == "true"}
	argIndex := 2
	
	if modelType != "" {
		whereClause += fmt.Sprintf(" AND model_type = $%d", argIndex)
		args = append(args, modelType)
		argIndex++
	}
	
	query := fmt.Sprintf(`
		SELECT 
			id, internal_key, display_name, custom_display_name, provider, model_type, category,
			description, max_tokens, support_streaming, is_featured, weight
		FROM ai_models %s
		ORDER BY weight DESC, display_name ASC
	`, whereClause)
	
	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询管理员模型失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()
	
	var models []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var internalKey, displayName, customDisplayName, provider, modelType, category, description string
		var maxTokens, weight int
		var supportStreaming, isFeatured bool
		
		err := rows.Scan(
			&id, &internalKey, &displayName, &customDisplayName, &provider, &modelType, &category,
			&description, &maxTokens, &supportStreaming, &isFeatured, &weight,
		)
		if err != nil {
			continue
		}
		
		// 管理员端显示完整信息，包括真实模型ID
		model := map[string]interface{}{
			"id":                  id,
			"internal_key":        internalKey,        // 真实模型ID (管理员可见)
			"display_name":        displayName,        // 原始显示名称
			"custom_display_name": customDisplayName,  // 自定义友好名称
			"provider":            provider,
			"type":                modelType,
			"category":            category,
			"description":         description,
			"max_tokens":          maxTokens,
			"support_streaming":   supportStreaming,
			"is_featured":         isFeatured,
			"weight":              weight,
		}
		
		models = append(models, model)
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取管理员模型列表成功",
		Data: map[string]interface{}{
			"models": models,
			"total":  len(models),
		},
	})
}

// 用户获取模型列表 (显示自定义友好名称)
func getUserModels(c *gin.Context) {
	modelType := c.Query("type")
	isActive := c.DefaultQuery("active", "true")
	
	whereClause := "WHERE is_active = $1"
	args := []interface{}{isActive == "true"}
	argIndex := 2
	
	if modelType != "" {
		whereClause += fmt.Sprintf(" AND model_type = $%d", argIndex)
		args = append(args, modelType)
		argIndex++
	}
	
	query := fmt.Sprintf(`
		SELECT 
			id, custom_display_name, provider, model_type, category,
			description, max_tokens, support_streaming, is_featured
		FROM ai_models %s
		ORDER BY weight DESC, custom_display_name ASC
	`, whereClause)
	
	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询用户模型失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()
	
	var models []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var customDisplayName, provider, modelType, category, description string
		var maxTokens int
		var supportStreaming, isFeatured bool
		
		err := rows.Scan(
			&id, &customDisplayName, &provider, &modelType, &category,
			&description, &maxTokens, &supportStreaming, &isFeatured,
		)
		if err != nil {
			continue
		}
		
		// 用户端只显示友好信息，隐藏真实模型ID
		model := map[string]interface{}{
			"id":                id,                    // 数据库ID用于调用
			"name":              customDisplayName,    // 显示自定义友好名称
			"provider":          provider,
			"type":              modelType,
			"category":          category,
			"description":       description,
			"max_tokens":        maxTokens,
			"support_streaming": supportStreaming,
			"is_featured":       isFeatured,
		}
		
		models = append(models, model)
	}
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取用户模型列表成功",
		Data: map[string]interface{}{
			"models": models,
			"total":  len(models),
		},
	})
}

// 对话接口 - 自动转换为真实模型名称调用
func chatCompletions(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 根据模型ID获取真实模型名称
	var internalKey string
	var supportStreaming bool
	err := db.QueryRow(
		"SELECT internal_key, support_streaming FROM ai_models WHERE id = $1::uuid AND is_active = true",
		req.ModelID,
	).Scan(&internalKey, &supportStreaming)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Code:    404,
				Message: "模型不存在或已禁用",
			})
		} else {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Code:    500,
				Message: "查询模型失败: " + err.Error(),
			})
		}
		return
	}
	
	// 检查流式支持
	if req.Stream && !supportStreaming {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "该模型不支持流式输出",
		})
		return
	}
	
	// 构建SiliconFlow请求 (使用真实模型名称)
	sfRequest := map[string]interface{}{
		"model":    internalKey,  // 使用真实模型名称调用SiliconFlow
		"messages": req.Messages,
		"stream":   req.Stream,
	}
	
	if req.Temperature > 0 {
		sfRequest["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		sfRequest["max_tokens"] = req.MaxTokens
	}
	
	jsonData, err := json.Marshal(sfRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "构建请求失败: " + err.Error(),
		})
		return
	}
	
	// 发送请求到SiliconFlow
	httpReq, err := http.NewRequest("POST", SILICONFLOW_BASE_URL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "创建请求失败: " + err.Error(),
		})
		return
	}
	
	httpReq.Header.Set("Authorization", "Bearer "+SILICONFLOW_API_KEY)
	httpReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "请求SiliconFlow失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, APIResponse{
			Code:    resp.StatusCode,
			Message: "SiliconFlow API错误: " + string(body),
		})
		return
	}
	
	// 处理响应
	if req.Stream {
		// 流式响应
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				c.Writer.WriteString(line + "\n\n")
				c.Writer.Flush()
			}
		}
	} else {
		// 非流式响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Code:    500,
				Message: "读取响应失败: " + err.Error(),
			})
			return
		}
		
		// 直接返回SiliconFlow的响应
		c.Header("Content-Type", "application/json")
		c.Writer.Write(body)
	}
}

// 系统信息
func getSystemInfo(c *gin.Context) {
	// 统计模型数量
	var totalModels, activeModels, chatModels, streamingModels int
	
	db.QueryRow("SELECT COUNT(*) FROM ai_models").Scan(&totalModels)
	db.QueryRow("SELECT COUNT(*) FROM ai_models WHERE is_active = true").Scan(&activeModels)
	db.QueryRow("SELECT COUNT(*) FROM ai_models WHERE model_type = 'chat' AND is_active = true").Scan(&chatModels)
	db.QueryRow("SELECT COUNT(*) FROM ai_models WHERE support_streaming = true AND is_active = true").Scan(&streamingModels)
	
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取系统信息成功",
		Data: map[string]interface{}{
			"system": map[string]interface{}{
				"name":    "YUNAI",
				"version": "1.0.0",
				"status":  "running",
			},
			"models": map[string]interface{}{
				"total":     totalModels,
				"active":    activeModels,
				"chat":      chatModels,
				"streaming": streamingModels,
			},
		},
	})
}
