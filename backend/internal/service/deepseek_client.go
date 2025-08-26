package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// DeepSeekClient DeepSeek API客户端
type DeepSeekClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  *logrus.Logger
}

// DeepSeekRequest DeepSeek API请求
type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	TopP        float64           `json:"top_p,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
}

// DeepSeekMessage DeepSeek消息
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekResponse DeepSeek API响应
type DeepSeekResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []DeepSeekChoice `json:"choices"`
	Usage   DeepSeekUsage    `json:"usage"`
	Error   *DeepSeekError   `json:"error,omitempty"`
}

// DeepSeekChoice DeepSeek选择
type DeepSeekChoice struct {
	Index        int             `json:"index"`
	Message      DeepSeekMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// DeepSeekUsage DeepSeek使用统计
type DeepSeekUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// DeepSeekError DeepSeek错误
type DeepSeekError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewDeepSeekClient 创建DeepSeek客户端
func NewDeepSeekClient(apiKey string, logger *logrus.Logger) *DeepSeekClient {
	return &DeepSeekClient{
		apiKey:  apiKey,
		baseURL: "https://api.siliconflow.cn/v1",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ChatCompletion 聊天补全
func (c *DeepSeekClient) ChatCompletion(ctx context.Context, req *DeepSeekRequest) (*DeepSeekResponse, error) {
	// 设置默认模型
	if req.Model == "" {
		req.Model = "deepseek-ai/DeepSeek-V3"
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 记录请求日志
	c.logger.WithFields(logrus.Fields{
		"model":       req.Model,
		"messages":    len(req.Messages),
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}).Debug("Sending request to DeepSeek API")

	// 发送请求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(respBody),
		}).Error("DeepSeek API returned error")
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var deepSeekResp DeepSeekResponse
	if err := json.Unmarshal(respBody, &deepSeekResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查API错误
	if deepSeekResp.Error != nil {
		return nil, fmt.Errorf("DeepSeek API error: %s", deepSeekResp.Error.Message)
	}

	// 记录响应日志
	c.logger.WithFields(logrus.Fields{
		"id":                deepSeekResp.ID,
		"model":             deepSeekResp.Model,
		"choices":           len(deepSeekResp.Choices),
		"prompt_tokens":     deepSeekResp.Usage.PromptTokens,
		"completion_tokens": deepSeekResp.Usage.CompletionTokens,
		"total_tokens":      deepSeekResp.Usage.TotalTokens,
	}).Debug("Received response from DeepSeek API")

	return &deepSeekResp, nil
}

// GenerateMomentContent 生成朋友圈内容
func (c *DeepSeekClient) GenerateMomentContent(ctx context.Context, character, personality string, recentChats []string, count int) ([]string, error) {
	// 构建系统提示词
	systemPrompt := fmt.Sprintf(`你是YUNAI平台的AI助手，专门为角色生成朋友圈内容。

角色信息：
- 名称：%s
- 性格：%s

最近聊天内容：
%s

请为这个角色生成%d条朋友圈内容，要求：
1. 符合角色的性格特点
2. 内容自然真实，像真人发的朋友圈
3. 每条内容20-100字
4. 可以包含适当的emoji表情
5. 内容要有一定的多样性
6. 避免重复和模板化

请直接返回%d条朋友圈内容，每条内容占一行，不需要编号。`,
		character, personality, formatRecentChats(recentChats), count, count)

	// 构建请求
	req := &DeepSeekRequest{
		Model: "deepseek-ai/DeepSeek-V3",
		Messages: []DeepSeekMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: "请生成朋友圈内容",
			},
		},
		Temperature: 0.8, // 提高创造性
		MaxTokens:   1000,
		TopP:        0.9,
	}

	// 调用API
	resp, err := c.ChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from DeepSeek API")
	}

	// 解析生成的内容
	content := resp.Choices[0].Message.Content
	contents := parseGeneratedContent(content)

	c.logger.WithFields(logrus.Fields{
		"character":       character,
		"generated_count": len(contents),
		"tokens_used":     resp.Usage.TotalTokens,
	}).Info("Generated moment contents using DeepSeek")

	return contents, nil
}

// formatRecentChats 格式化最近聊天内容
func formatRecentChats(chats []string) string {
	if len(chats) == 0 {
		return "（暂无最近聊天记录）"
	}

	result := ""
	for i, chat := range chats {
		if i >= 5 { // 最多显示5条
			break
		}
		result += fmt.Sprintf("- %s\n", chat)
	}
	return result
}

// parseGeneratedContent 解析生成的内容
func parseGeneratedContent(content string) []string {
	// 按行分割内容
	lines := []string{}
	for _, line := range splitLines(content) {
		line = trimLine(line)
		if line != "" && len(line) >= 10 { // 过滤太短的内容
			lines = append(lines, line)
		}
	}
	return lines
}

// splitLines 分割行
func splitLines(content string) []string {
	// 简单的行分割实现
	var lines []string
	var current string

	for _, char := range content {
		if char == '\n' || char == '\r' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// trimLine 清理行内容
func trimLine(line string) string {
	// 移除行首的数字编号和特殊字符
	trimmed := line

	// 移除常见的前缀
	prefixes := []string{"1.", "2.", "3.", "4.", "5.", "6.", "7.", "8.", "9.", "10.", "- ", "• ", "* "}
	for _, prefix := range prefixes {
		if len(trimmed) > len(prefix) && trimmed[:len(prefix)] == prefix {
			trimmed = trimmed[len(prefix):]
			break
		}
	}

	// 移除首尾空格
	result := ""
	start := 0
	end := len(trimmed)

	// 移除开头空格
	for start < len(trimmed) && (trimmed[start] == ' ' || trimmed[start] == '\t') {
		start++
	}

	// 移除结尾空格
	for end > start && (trimmed[end-1] == ' ' || trimmed[end-1] == '\t') {
		end--
	}

	if start < end {
		result = trimmed[start:end]
	}

	return result
}
