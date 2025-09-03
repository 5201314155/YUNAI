package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🎯 YUNAI 最小测试服务器启动中...")

	router := gin.Default()

	// 启用CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 健康检查
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "YUNAI最小测试服务",
			"message": "服务运行正常",
		})
	})

	// 单聊API
	router.POST("/api/chat/single", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "请求格式错误: " + err.Error(),
			})
			return
		}

		message, _ := req["message"].(string)

		c.JSON(200, gin.H{
			"message":    fmt.Sprintf("收到你的消息：%s。这是一个测试回复！", message),
			"audio_data": "",
			"emotion":    "friendly",
			"success":    true,
			"chat_type":  "single",
		})
	})

	// 群聊API
	router.POST("/api/chat/group", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "请求格式错误: " + err.Error(),
			})
			return
		}

		message, _ := req["message"].(string)

		c.JSON(200, gin.H{
			"message":    fmt.Sprintf("在群聊中收到消息：%s。大家好！", message),
			"audio_data": "",
			"emotion":    "cheerful",
			"success":    true,
			"chat_type":  "group",
		})
	})

	// 世界观API
	router.POST("/api/world", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success":    true,
			"message":    "欢迎来到YUNAI的虚拟世界！这里有美丽的风景、有趣的角色，还有无限的可能性等待你去探索。",
			"audio_data": "",
			"world_type": "yunai_universe",
		})
	})

	// 关系网API
	router.POST("/api/relationship", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success":      true,
			"message":      "你的关系网络很丰富！与各个AI角色都有不错的互动，亲密度都在稳步提升中。",
			"audio_data":   "",
			"network_type": "relationship_analysis",
		})
	})

	// 朋友圈API
	router.POST("/api/moments", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success":     true,
			"message":     "朋友圈很热闹呢！小雨刚刚分享了一张美丽的日落照片，小明在讨论最新的科技趋势，大家都很活跃！",
			"audio_data":  "",
			"social_type": "moments",
		})
	})

	// TTS API
	router.POST("/api/voice/tts", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "请求格式错误: " + err.Error(),
			})
			return
		}

		text, _ := req["text"].(string)
		if text == "" {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "文本内容不能为空",
			})
			return
		}

		c.JSON(200, gin.H{
			"success":    true,
			"audio_data": "ZmFrZSBhdWRpbyBkYXRh", // base64编码的"fake audio data"
			"text":       text,
			"voice_id":   "default",
			"emotion":    "friendly",
			"duration":   len(text) * 100,
		})
	})

	fmt.Println("🚀 服务器启动成功！")
	fmt.Println("   📡 服务地址: http://localhost:8083")
	fmt.Println("   💬 单聊API: http://localhost:8083/api/chat/single")
	fmt.Println("   👥 群聊API: http://localhost:8083/api/chat/group")
	fmt.Println("   🌍 世界观API: http://localhost:8083/api/world")
	fmt.Println("   🕸️  关系网API: http://localhost:8083/api/relationship")
	fmt.Println("   📱 朋友圈API: http://localhost:8083/api/moments")
	fmt.Println("   🎵 语音API: http://localhost:8083/api/voice/tts")
	fmt.Println("   ❤️  健康检查: http://localhost:8083/api/health")
	fmt.Println()

	// 启动服务器
	router.Run(":8083")
}
