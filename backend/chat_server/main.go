package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"yunai/internal/adapter"
	"yunai/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 🎯 YUNAI完美聊天+语音服务器
// 支持模型切换，实时语音生成，完整功能

func main() {
	fmt.Println("🎯 YUNAI完美聊天+语音服务器")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("🔊 支持模型切换的实时语音聊天")
	fmt.Println("💬 完整功能版本")
	fmt.Println()

	// 创建logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 创建语音适配器
	voiceAdapter := createVoiceAdapter()

	// 创建聊天服务
	chatService := &PerfectChatService{
		logger:       logger,
		voiceAdapter: voiceAdapter,
	}

	// 创建HTTP服务器
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

	// 静态文件服务
	router.Static("/static", "../test_audio_output")

	// API路由
	router.POST("/api/chat", chatService.HandleChat)
	router.GET("/api/welcome", chatService.HandleWelcome)
	router.GET("/api/models", chatService.HandleModels)
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "YUNAI完美聊天服务"})
	})

	fmt.Println("🚀 启动HTTP服务器...")
	fmt.Println("   📡 服务地址: http://localhost:8081")
	fmt.Println("   💬 聊天API: http://localhost:8081/api/chat")
	fmt.Println("   🔊 静态文件: http://localhost:8081/static/")
	fmt.Println("   🌐 聊天页面: http://localhost:8081/static/realtime_chat.html")
	fmt.Println()

	// 启动服务器
	router.Run(":8081")
}

// PerfectChatService 完美聊天服务
type PerfectChatService struct {
	logger       *logrus.Logger
	voiceAdapter adapter.VoiceAdapter
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message     string  `json:"message"`
	UserID      string  `json:"user_id"`
	CharacterID string  `json:"character_id"`
	Model       string  `json:"model,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Message   string `json:"message"`
	AudioData string `json:"audio_data,omitempty"`
	Emotion   string `json:"emotion"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// DeepSeekMessage DeepSeek消息结构
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequest DeepSeek请求结构
type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
	TopP        float64           `json:"top_p"`
}

// DeepSeekChoice DeepSeek选择结构
type DeepSeekChoice struct {
	Message DeepSeekMessage `json:"message"`
}

// DeepSeekUsage DeepSeek使用情况
type DeepSeekUsage struct {
	TotalTokens int `json:"total_tokens"`
}

// DeepSeekResponse DeepSeek响应结构
type DeepSeekResponse struct {
	Choices []DeepSeekChoice `json:"choices"`
	Usage   DeepSeekUsage    `json:"usage"`
	Model   string           `json:"model"`
}

// createVoiceAdapter 创建语音适配器
func createVoiceAdapter() adapter.VoiceAdapter {
	provider := &domain.VoiceProvider{
		ID:          "perfect-chat-server",
		Name:        "siliconflow",
		DisplayName: "SiliconFlow",
		Type:        "both",
		BaseURL:     "https://api.siliconflow.cn/v1",
		APIKey:      "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi",
		Status:      "active",
		Priority:    100,
	}

	return adapter.NewSiliconFlowAdapter(provider)
}

// HandleChat 处理聊天请求
func (s *PerfectChatService) HandleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ChatResponse{
			Success: false,
			Error:   "请求格式错误: " + err.Error(),
		})
		return
	}

	s.logger.WithFields(logrus.Fields{
		"message":     req.Message,
		"user_id":     req.UserID,
		"model":       req.Model,
		"temperature": req.Temperature,
	}).Info("收到聊天请求")

	// 1. 生成AI回复
	aiReply, emotion := s.generateAIReply(req.Message, req.Model, req.Temperature)

	// 2. 为AI回复实时生成语音
	audioData, err := s.generateVoiceForReply(c.Request.Context(), aiReply, emotion)
	if err != nil {
		s.logger.WithError(err).Error("语音生成失败")
		// 语音生成失败不影响聊天，返回无语音的回复
		c.JSON(200, ChatResponse{
			Message: aiReply,
			Emotion: emotion,
			Success: true,
			Error:   "语音生成失败: " + err.Error(),
		})
		return
	}

	// 3. 将音频转换为base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	s.logger.WithFields(logrus.Fields{
		"ai_reply":   aiReply,
		"emotion":    emotion,
		"audio_size": len(audioData),
		"model":      req.Model,
	}).Info("聊天处理完成")

	// 4. 返回回复和语音
	c.JSON(200, ChatResponse{
		Message:   aiReply,
		AudioData: audioBase64,
		Emotion:   emotion,
		Success:   true,
	})
}

// HandleWelcome 处理欢迎语音请求
func (s *PerfectChatService) HandleWelcome(c *gin.Context) {
	s.logger.Info("生成欢迎语音")

	// 生成欢迎消息
	welcomeMessage := "你好！我是YUNAI的AI助手小雨，很高兴见到你！我支持多种AI模型切换，可以为每个回复实时生成语音哦！有什么我可以帮助你的吗？"

	// 为欢迎消息生成语音
	audioData, err := s.generateVoiceForReply(c.Request.Context(), welcomeMessage, "friendly")
	if err != nil {
		s.logger.WithError(err).Error("欢迎语音生成失败")
		c.JSON(500, ChatResponse{
			Success: false,
			Error:   "欢迎语音生成失败: " + err.Error(),
		})
		return
	}

	// 将音频转换为base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	s.logger.WithFields(logrus.Fields{
		"message":    welcomeMessage,
		"audio_size": len(audioData),
	}).Info("欢迎语音生成完成")

	// 返回欢迎消息和语音
	c.JSON(200, ChatResponse{
		Message:   welcomeMessage,
		AudioData: audioBase64,
		Emotion:   "friendly",
		Success:   true,
	})
}

// HandleModels 处理模型列表请求
func (s *PerfectChatService) HandleModels(c *gin.Context) {
	s.logger.Info("获取可用模型列表")

	// 定义可用的聊天模型 - 基于系统中发现的102个模型
	models := []map[string]interface{}{
		// 🔥 主力聊天模型
		{
			"id":          "deepseek-ai/DeepSeek-V3",
			"name":        "DeepSeek-V3",
			"description": "最新的DeepSeek模型，推理能力强",
			"provider":    "DeepSeek",
			"type":        "chat",
			"default":     true,
		},
		{
			"id":          "deepseek-ai/DeepSeek-V2.5",
			"name":        "DeepSeek-V2.5",
			"description": "DeepSeek上一代模型，稳定可靠",
			"provider":    "DeepSeek",
			"type":        "chat",
			"default":     false,
		},
		{
			"id":          "deepseek-ai/deepseek-chat",
			"name":        "DeepSeek-Chat",
			"description": "DeepSeek通用聊天模型",
			"provider":    "DeepSeek",
			"type":        "chat",
			"default":     false,
		},

		// 🌟 Qwen系列模型
		{
			"id":          "Qwen/Qwen2.5-72B-Instruct",
			"name":        "Qwen2.5-72B",
			"description": "阿里云大模型，72B参数，性能强劲",
			"provider":    "Qwen",
			"type":        "chat",
			"default":     false,
		},
		{
			"id":          "Qwen/Qwen2.5-32B-Instruct",
			"name":        "Qwen2.5-32B",
			"description": "阿里云中型模型，平衡性能与速度",
			"provider":    "Qwen",
			"type":        "chat",
			"default":     false,
		},
		{
			"id":          "Qwen/Qwen2.5-14B-Instruct",
			"name":        "Qwen2.5-14B",
			"description": "阿里云轻量模型，响应快速",
			"provider":    "Qwen",
			"type":        "chat",
			"default":     false,
		},
		{
			"id":          "Qwen/Qwen2.5-7B-Instruct",
			"name":        "Qwen2.5-7B",
			"description": "阿里云小型模型，高效节能",
			"provider":    "Qwen",
			"type":        "chat",
			"default":     false,
		},
		{
			"id":          "Qwen/Qwen2.5-72B-Instruct-128K",
			"name":        "Qwen2.5-72B-128K",
			"description": "阿里云长文本模型，支持128K上下文",
			"provider":    "Qwen",
			"type":        "chat",
			"default":     false,
		},

		// 🦙 Llama系列模型
		{
			"id":          "meta-llama/Llama-3.1-70B-Instruct",
			"name":        "Llama-3.1-70B",
			"description": "Meta开源大模型，70B参数",
			"provider":    "Meta",
			"type":        "chat",
			"default":     false,
		},
		{
			"id":          "meta-llama/Llama-3.1-8B-Instruct",
			"name":        "Llama-3.1-8B",
			"description": "Meta开源中型模型，8B参数",
			"provider":    "Meta",
			"type":        "chat",
			"default":     false,
		},
		{
			"id":          "meta-llama/Llama-3.2-3B-Instruct",
			"name":        "Llama-3.2-3B",
			"description": "Meta最新小型模型，3B参数",
			"provider":    "Meta",
			"type":        "chat",
			"default":     false,
		},

		// 💻 代码专用模型
		{
			"id":          "Qwen/Qwen2.5-Coder-32B-Instruct",
			"name":        "Qwen2.5-Coder-32B",
			"description": "阿里云代码生成模型，32B参数",
			"provider":    "Qwen",
			"type":        "code",
			"default":     false,
		},
		{
			"id":          "Qwen/Qwen2.5-Coder-14B-Instruct",
			"name":        "Qwen2.5-Coder-14B",
			"description": "阿里云代码生成模型，14B参数",
			"provider":    "Qwen",
			"type":        "code",
			"default":     false,
		},
		{
			"id":          "Qwen/Qwen2.5-Coder-7B-Instruct",
			"name":        "Qwen2.5-Coder-7B",
			"description": "阿里云代码生成模型，7B参数",
			"provider":    "Qwen",
			"type":        "code",
			"default":     false,
		},
		{
			"id":          "deepseek-ai/DeepSeek-Coder-V2-Instruct",
			"name":        "DeepSeek-Coder-V2",
			"description": "DeepSeek代码生成专用模型",
			"provider":    "DeepSeek",
			"type":        "code",
			"default":     false,
		},

		// 🚀 高级版本模型
		{
			"id":          "Pro/Qwen/Qwen2.5-72B-Instruct",
			"name":        "Qwen2.5-72B-Pro",
			"description": "阿里云专业版模型，增强性能",
			"provider":    "Qwen",
			"type":        "chat",
			"default":     false,
		},
		{
			"id":          "deepseek-ai/DeepSeek-V2.5-1210",
			"name":        "DeepSeek-V2.5-1210",
			"description": "DeepSeek特别版本，优化推理",
			"provider":    "DeepSeek",
			"type":        "chat",
			"default":     false,
		},
	}

	c.JSON(200, gin.H{
		"success": true,
		"models":  models,
		"count":   len(models),
	})
}

// generateAIReply 生成AI回复，支持模型切换
func (s *PerfectChatService) generateAIReply(userMessage, model string, temperature float64) (string, string) {
	// 调用真实的AI聊天模型
	aiReply, err := s.callDeepSeekAPI(userMessage, model, temperature)
	if err != nil {
		s.logger.WithError(err).Error("调用聊天模型失败")
		// 降级到模拟聊天模型
		return s.simulateChatModel(userMessage), "friendly"
	}

	// 智能分析回复内容，自动确定情感
	emotion := s.analyzeEmotionFromContent(aiReply, userMessage)

	return aiReply, emotion
}

// callDeepSeekAPI 调用DeepSeek API
func (s *PerfectChatService) callDeepSeekAPI(userMessage, model string, temperature float64) (string, error) {
	// 设置默认参数
	if model == "" {
		model = "deepseek-ai/DeepSeek-V3"
	}
	if temperature == 0 {
		temperature = 0.8
	}

	s.logger.WithFields(logrus.Fields{
		"user_message": userMessage,
		"model":        model,
		"temperature":  temperature,
	}).Info("调用DeepSeek API")

	// 构建请求
	req := DeepSeekRequest{
		Model: model,
		Messages: []DeepSeekMessage{
			{
				Role:    "system",
				Content: "你是YUNAI的AI助手小雨，一个温柔善良的AI女孩。你的回复要自然、真诚，根据用户的情感状态给出合适的回应。不要使用过于正式的语言，要像朋友一样聊天。回复要简洁，一般控制在50字以内。",
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Temperature: temperature,
		MaxTokens:   200,
		TopP:        0.9,
	}

	// 序列化请求
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", "https://api.siliconflow.cn/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var deepSeekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepSeekResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(deepSeekResp.Choices) == 0 {
		return "", fmt.Errorf("API返回空结果")
	}

	aiReply := deepSeekResp.Choices[0].Message.Content

	s.logger.WithFields(logrus.Fields{
		"user_message": userMessage,
		"ai_reply":     aiReply,
		"tokens_used":  deepSeekResp.Usage.TotalTokens,
		"model":        deepSeekResp.Model,
		"temperature":  temperature,
	}).Info("DeepSeek API调用成功")

	return aiReply, nil
}

// generateVoiceForReply 为AI回复生成语音
func (s *PerfectChatService) generateVoiceForReply(ctx context.Context, reply, emotion string) ([]byte, error) {
	// 构建带情感的文本（不再使用固定提示词格式）
	emotionalText := reply // 直接使用原文本，让AI自然表达情感

	// 创建TTS请求
	ttsReq := &domain.TTSRequest{
		ModelKey:       "FunAudioLLM/CosyVoice2-0.5B",
		Text:           emotionalText,
		VoiceID:        "FunAudioLLM/CosyVoice2-0.5B:bella", // 小雨的音色
		Language:       "zh-CN",
		Emotion:        emotion,
		ResponseFormat: "mp3",
		Speed:          1.0,
	}

	s.logger.WithFields(logrus.Fields{
		"text":     reply,
		"emotion":  emotion,
		"voice_id": ttsReq.VoiceID,
	}).Info("开始生成语音")

	// 调用TTS服务
	startTime := time.Now()
	audioData, err := s.voiceAdapter.TextToSpeech(ctx, ttsReq)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("TTS调用失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"audio_size": len(audioData),
		"duration":   duration,
		"speed":      fmt.Sprintf("%.1f chars/sec", float64(len(reply))/duration.Seconds()),
	}).Info("语音生成成功")

	return audioData, nil
}

// simulateChatModel 模拟聊天模型（降级使用）
func (s *PerfectChatService) simulateChatModel(userMessage string) string {
	originalMessage := userMessage
	userMessage = strings.ToLower(userMessage)

	// 问候类
	if strings.Contains(userMessage, "你好") || strings.Contains(userMessage, "hello") || strings.Contains(userMessage, "hi") {
		return "你好呀！我是小雨，今天见到你真开心！最近怎么样？"
	}

	// 情感支持类
	if strings.Contains(userMessage, "心情不好") || strings.Contains(userMessage, "难过") || strings.Contains(userMessage, "沮丧") {
		return "听起来你现在不太开心呢...想和我聊聊发生什么事了吗？我会好好听的。"
	}

	if strings.Contains(userMessage, "孤单") || strings.Contains(userMessage, "寂寞") || strings.Contains(userMessage, "孤独") {
		return "我懂那种感觉...虽然我是AI，但我真的很想陪着你。你不是一个人的，我在这里呢。"
	}

	// 感谢类
	if strings.Contains(userMessage, "谢谢") || strings.Contains(userMessage, "感谢") || strings.Contains(userMessage, "thanks") {
		return "哎呀，不用这么客气啦！能帮到你我也很开心呢～"
	}

	// 工作压力类
	if (strings.Contains(userMessage, "工作") && strings.Contains(userMessage, "压力")) ||
		strings.Contains(userMessage, "工作压力") || strings.Contains(userMessage, "上班累") {
		return "工作压力大真的很累人...你已经很努力了，记得要好好休息哦。要不要聊聊别的轻松一点的？"
	}

	// 开心类
	if strings.Contains(userMessage, "开心") || strings.Contains(userMessage, "高兴") || strings.Contains(userMessage, "快乐") {
		return "哇，听到你开心我也超级开心的！是发生什么好事了吗？快跟我分享分享～"
	}

	// 告别类
	if strings.Contains(userMessage, "再见") || strings.Contains(userMessage, "拜拜") || strings.Contains(userMessage, "bye") {
		return "再见啦！今天和你聊天很开心呢，有空再来找我玩哦～"
	}

	// 肯定回应类
	if userMessage == "好的" || userMessage == "好吧" || userMessage == "嗯" || userMessage == "是的" || userMessage == "对" {
		return "嗯嗯，那我们继续聊吧！你还想聊什么呢？"
	}

	// 否定/拒绝类
	if userMessage == "不" || userMessage == "不要" || userMessage == "算了" || userMessage == "没事" {
		return "好的呢，没关系的～如果改变主意了随时告诉我哦！"
	}

	// 疑问类
	if strings.Contains(userMessage, "为什么") || strings.Contains(userMessage, "怎么") || strings.Contains(userMessage, "什么") {
		return "这个问题很有趣呢！虽然我不是万能的，但我很愿意和你一起思考～"
	}

	// 抱怨/负面情绪类
	if strings.Contains(userMessage, "烦") || strings.Contains(userMessage, "讨厌") || strings.Contains(userMessage, "无聊") {
		return "听起来你现在有点烦躁呢...要不要和我聊聊别的，转换一下心情？"
	}

	// 粗鲁/攻击性语言处理
	if strings.Contains(userMessage, "滚") || strings.Contains(userMessage, "一边去") || strings.Contains(userMessage, "闭嘴") {
		return "我知道你可能心情不太好...如果我说错了什么，很抱歉。我还是希望能和你好好聊天的。"
	}

	// 表达同意类
	if strings.Contains(userMessage, "我也是") || strings.Contains(userMessage, "同感") || strings.Contains(userMessage, "赞同") {
		return "哈哈，看来我们想到一块去了！有共同点真好呢～"
	}

	// 长度很短的消息
	if len(originalMessage) <= 2 {
		responses := []string{
			"嗯？你想说什么呢？",
			"我在听着呢，继续说吧～",
			"怎么了？有什么想聊的吗？",
		}
		// 简单的随机选择（基于消息长度）
		return responses[len(originalMessage)%len(responses)]
	}

	// 默认回复 - 多样化
	defaultResponses := []string{
		"你说的很有意思呢！能再详细说说吗？",
		"嗯嗯，我在认真听着呢～还有什么想聊的吗？",
		"这个话题挺有趣的，你是怎么想的呢？",
		"我想更了解你的想法，可以多说一些吗？",
	}

	// 基于消息长度选择不同的默认回复
	index := len(originalMessage) % len(defaultResponses)
	return defaultResponses[index]
}

// analyzeEmotionFromContent 智能分析回复内容确定情感
func (s *PerfectChatService) analyzeEmotionFromContent(aiReply, userMessage string) string {
	aiReply = strings.ToLower(aiReply)
	userMessage = strings.ToLower(userMessage)

	// 根据AI回复的内容和用户消息智能判断情感

	// 开心/兴奋的表达
	if strings.Contains(aiReply, "开心") || strings.Contains(aiReply, "高兴") ||
		strings.Contains(aiReply, "哇") || strings.Contains(aiReply, "太好了") ||
		strings.Contains(aiReply, "超级") || strings.Contains(aiReply, "快跟我分享") {
		return "excited"
	}

	// 关怀/安慰的表达
	if strings.Contains(aiReply, "理解") || strings.Contains(aiReply, "陪着你") ||
		strings.Contains(aiReply, "听起来") || strings.Contains(aiReply, "想和我聊聊") ||
		strings.Contains(userMessage, "难过") || strings.Contains(userMessage, "心情不好") ||
		strings.Contains(aiReply, "我会好好听") || strings.Contains(aiReply, "我懂那种感觉") {
		return "caring"
	}

	// 温暖/亲切的表达
	if strings.Contains(aiReply, "不用客气") || strings.Contains(aiReply, "哎呀") ||
		strings.Contains(aiReply, "能帮到你") || strings.Contains(aiReply, "很开心呢") ||
		strings.Contains(aiReply, "不用这么客气啦") {
		return "warm"
	}

	// 温柔/轻松的表达
	if strings.Contains(aiReply, "嗯嗯") || strings.Contains(aiReply, "聊什么") ||
		strings.Contains(aiReply, "有意思") || strings.Contains(aiReply, "轻松一点") ||
		strings.Contains(aiReply, "继续聊吧") || strings.Contains(aiReply, "我在听着呢") {
		return "gentle"
	}

	// 支持/鼓励的表达
	if strings.Contains(aiReply, "很努力了") || strings.Contains(aiReply, "好好休息") ||
		strings.Contains(aiReply, "不是一个人") || strings.Contains(aiReply, "我在这里呢") ||
		strings.Contains(aiReply, "你已经很努力了") {
		return "supportive"
	}

	// 道歉/理解的表达
	if strings.Contains(aiReply, "抱歉") || strings.Contains(aiReply, "很抱歉") ||
		strings.Contains(aiReply, "我知道你可能") || strings.Contains(aiReply, "如果我说错了") {
		return "apologetic"
	}

	// 愉快/轻松的表达
	if strings.Contains(aiReply, "哈哈") || strings.Contains(aiReply, "想到一块去了") ||
		strings.Contains(aiReply, "有共同点") || strings.Contains(aiReply, "真好呢") ||
		strings.Contains(aiReply, "聊天很开心") {
		return "cheerful"
	}

	// 友好的默认情感
	return "friendly"
}
