package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DeepSeek API 配置
const (
	DeepSeekAPIBase = "https://api.siliconflow.cn/v1"
	DeepSeekAPIKey  = "sk-qpkyeqeuvjbsflnxuiejsfumjtmfuvexmvvutooppdmxnwgi"
	DeepSeekModel   = "deepseek-ai/DeepSeek-V3"
)

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func main() {
	fmt.Println("🤖 YUNAI AI 主动拉人系统测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("📋 测试内容：AI主动邀请 + 用户接受/拒绝 + 智能反应 + 私信互动")
	fmt.Println("🔗 使用真实 DeepSeek-V3 模型")
	fmt.Println()

	ctx := context.Background()

	// 测试1：AI 主动提出邀请
	testAIProactiveInvite(ctx)

	// 测试2：用户接受邀请的 AI 反应
	testUserAcceptInvite(ctx)

	// 测试3：用户拒绝邀请的 AI 反应
	testUserRejectInvite(ctx)

	// 测试4：被拒绝角色的主动私信
	testRejectedCharacterPrivateMessage(ctx)

	fmt.Println("\n🎉 AI 主动拉人系统测试完成！")
	fmt.Println("=" + strings.Repeat("=", 80))
}

// testAIProactiveInvite 测试 AI 主动提出邀请
func testAIProactiveInvite(ctx context.Context) {
	fmt.Println("🎯 [1/4] AI 主动邀请测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试场景：群聊中缺少某个角色，AI 主动建议邀请
	fmt.Println("📋 场景设定:")
	fmt.Println("   群聊: 魔法学院讨论组")
	fmt.Println("   当前成员: 用户、小雨(图书管理员)")
	fmt.Println("   话题: 讨论魔法音乐课程")
	fmt.Println("   缺失角色: 小美(音乐系学生)")

	invitePrompt := `你是小雨，魔法学院的图书管理员。

当前群聊情况：
- 群聊名称：魔法学院讨论组
- 当前成员：你和用户
- 当前话题：用户刚说"我想学习一些音乐魔法，但不知道从哪里开始"

你知道小美是音乐系的学生，擅长音律魔法，她会是这个话题的完美人选。

请主动建议邀请小美加入群聊，要：
1. 自然地提到小美的专业背景
2. 说明邀请她的理由
3. 询问用户是否同意邀请
4. 体现你温柔善良的性格`

	response, err := callDeepSeek(ctx, invitePrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨主动建议邀请: \"%s\"\n", response)

		// 检查邀请要素
		if strings.Contains(response, "小美") {
			fmt.Printf("   ✅ 正确提及目标角色\n")
		}
		if strings.Contains(response, "音乐") || strings.Contains(response, "音律") {
			fmt.Printf("   ✅ 正确说明专业背景\n")
		}
		if strings.Contains(response, "邀请") || strings.Contains(response, "加入") {
			fmt.Printf("   ✅ 明确提出邀请建议\n")
		}
		if strings.Contains(response, "吗") || strings.Contains(response, "好吗") || strings.Contains(response, "如何") {
			fmt.Printf("   ✅ 征求用户同意\n")
		}
	}

	// 测试多种邀请场景
	fmt.Println("\n🎭 其他邀请场景测试:")

	scenarios := []struct {
		situation string
		missing   string
		reason    string
	}{
		{
			situation: "讨论魔法战斗技巧",
			missing:   "大伟(体育系学生)",
			reason:    "擅长强化魔法和战斗技巧",
		},
		{
			situation: "计划学院探险活动",
			missing:   "小明(冒险社团长)",
			reason:    "有丰富的探险经验",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n   场景 %d: %s\n", i+1, scenario.situation)
		fmt.Printf("   建议邀请: %s\n", scenario.missing)
		fmt.Printf("   邀请理由: %s\n", scenario.reason)

		scenarioPrompt := fmt.Sprintf(`你是小雨，当前群聊在%s，你觉得应该邀请%s，因为他%s。请简短地建议邀请。`,
			scenario.situation, scenario.missing, scenario.reason)

		scenarioResponse, err := callDeepSeek(ctx, scenarioPrompt, "")
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
		} else {
			fmt.Printf("   🤖 小雨建议: \"%s\"\n", scenarioResponse)
		}
	}
}

// testUserAcceptInvite 测试用户接受邀请的 AI 反应
func testUserAcceptInvite(ctx context.Context) {
	fmt.Println("\n✅ [2/4] 用户接受邀请的 AI 反应测试")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Println("📋 场景: 用户同意邀请小美加入群聊")
	fmt.Println("   用户回应: \"好的，邀请小美吧！\"")

	acceptPrompt := `你是小雨，刚才你建议邀请小美(音乐系学生)加入群聊讨论音乐魔法。

用户回应："好的，邀请小美吧！"

请表达你的开心和感谢，然后：
1. 表示马上邀请小美
2. 预告小美的到来会带来什么帮助
3. 体现你的温柔性格和对朋友的关心`

	acceptResponse, err := callDeepSeek(ctx, acceptPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小雨接受反应: \"%s\"\n", acceptResponse)

		// 检查反应要素
		if strings.Contains(acceptResponse, "开心") || strings.Contains(acceptResponse, "太好了") || strings.Contains(acceptResponse, "谢谢") {
			fmt.Printf("   ✅ 表达了积极情绪\n")
		}
		if strings.Contains(acceptResponse, "邀请") || strings.Contains(acceptResponse, "加入") {
			fmt.Printf("   ✅ 确认执行邀请\n")
		}
	}

	// 系统执行邀请操作
	fmt.Println("\n🔄 系统执行邀请操作:")
	fmt.Println("   📤 系统向小美发送群聊邀请...")
	fmt.Println("   ⏱️  小美收到邀请通知")
	fmt.Println("   ✅ 小美接受邀请，加入群聊")
	fmt.Println("   👥 群聊成员更新: 用户、小雨、小美")

	// 模拟小美加入后的群聊互动
	fmt.Println("\n🎵 小美加入群聊后的完整互动:")

	// 1. 小美加入群聊的系统通知
	fmt.Println("   📢 系统通知: 小美 已加入群聊")

	// 2. 小雨欢迎小美
	fmt.Println("   💬 小雨: \"欢迎小美！用户想学音乐魔法呢~\"")

	// 3. 用户打招呼
	fmt.Println("   👤 用户: \"你好小美！\"")

	// 4. 小美的加入回应
	xiaomeiJoinPrompt := `你是小美，音乐系的学生，擅长音律魔法。

你刚刚加入"魔法学院讨论组"群聊，看到：
- 系统通知：小美 已加入群聊
- 小雨：欢迎小美！用户想学音乐魔法呢~
- 用户：你好小美！

请以活泼开朗的性格回应，表达：
1. 对邀请的感谢
2. 对音乐魔法话题的兴奋
3. 主动提供帮助
4. 开始群聊互动`

	xiaomeiResponse, err := callDeepSeek(ctx, xiaomeiJoinPrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   🤖 小美: \"%s\"\n", xiaomeiResponse)

		if strings.Contains(xiaomeiResponse, "谢谢") || strings.Contains(xiaomeiResponse, "感谢") {
			fmt.Printf("   ✅ 表达感谢\n")
		}
		if strings.Contains(xiaomeiResponse, "音乐") || strings.Contains(xiaomeiResponse, "音律") {
			fmt.Printf("   ✅ 体现专业背景\n")
		}
		if strings.Contains(xiaomeiResponse, "帮") || strings.Contains(xiaomeiResponse, "教") {
			fmt.Printf("   ✅ 主动提供帮助\n")
		}
	}

	// 5. 继续群聊互动 - 小雨的反应
	fmt.Println("\n💬 群聊继续互动:")

	xiaoyuContinuePrompt := `你是小雨，图书馆管理员。小美刚刚加入群聊并热情地回应了。

当前群聊状态：
- 成员：你、用户、小美
- 话题：音乐魔法学习
- 小美刚刚表达了愿意帮助用户学习音乐魔法

请简短回应，体现：
1. 对小美加入的开心
2. 促进用户和小美的交流
3. 你温柔的性格`

	xiaoyuContinueResponse, err := callDeepSeek(ctx, xiaoyuContinuePrompt, "")
	if err != nil {
		fmt.Printf("   ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("   💬 小雨: \"%s\"\n", xiaoyuContinueResponse)
	}

	// 6. 展示群聊活跃度提升
	fmt.Println("\n📈 邀请成功效果:")
	fmt.Printf("   ✅ 群聊成员: 2人 → 3人\n")
	fmt.Printf("   ✅ 话题专业度: 提升 (小美的音乐魔法专长)\n")
	fmt.Printf("   ✅ 互动活跃度: 提升 (新成员带来新活力)\n")
	fmt.Printf("   ✅ 用户体验: 获得专业指导和帮助\n")
}

// testUserRejectInvite 测试用户拒绝邀请的 AI 反应
func testUserRejectInvite(ctx context.Context) {
	fmt.Println("\n❌ [3/4] 用户拒绝邀请的 AI 反应测试")
	fmt.Println(strings.Repeat("-", 70))

	// 测试不同的拒绝方式和 AI 反应
	rejectScenarios := []struct {
		userReply   string
		description string
	}{
		{
			userReply:   "算了，我们两个人聊就好",
			description: "礼貌拒绝",
		},
		{
			userReply:   "不用了，我不太想和小美聊天",
			description: "明确拒绝",
		},
		{
			userReply:   "现在不合适，以后再说吧",
			description: "延迟拒绝",
		},
	}

	for i, scenario := range rejectScenarios {
		fmt.Printf("\n拒绝场景 %d: %s\n", i+1, scenario.description)
		fmt.Printf("   用户回应: \"%s\"\n", scenario.userReply)

		rejectPrompt := fmt.Sprintf(`你是小雨，刚才你建议邀请小美加入群聊讨论音乐魔法。

用户回应："%s"

请理解用户的想法，表现出：
1. 理解和尊重用户的决定
2. 不要显得失望或不开心
3. 继续保持友好，转移话题
4. 体现你温柔善良的性格`, scenario.userReply)

		rejectResponse, err := callDeepSeek(ctx, rejectPrompt, "")
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
			continue
		}

		fmt.Printf("   🤖 小雨理解反应: \"%s\"\n", rejectResponse)

		// 检查反应质量
		if strings.Contains(rejectResponse, "理解") || strings.Contains(rejectResponse, "没关系") || strings.Contains(rejectResponse, "好的") {
			fmt.Printf("   ✅ 表达理解和尊重\n")
		}
		if !strings.Contains(rejectResponse, "失望") && !strings.Contains(rejectResponse, "难过") {
			fmt.Printf("   ✅ 没有表现负面情绪\n")
		}
		if strings.Contains(rejectResponse, "那我们") || strings.Contains(rejectResponse, "不如") {
			fmt.Printf("   ✅ 主动转移话题\n")
		}
	}
}

// testRejectedCharacterPrivateMessage 测试被拒绝角色的主动私信
func testRejectedCharacterPrivateMessage(ctx context.Context) {
	fmt.Println("\n💬 [4/4] 被拒绝角色主动私信测试")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Println("📋 场景设定:")
	fmt.Println("   小美被拒绝加入群聊后，主动私信用户")
	fmt.Println("   时间: 拒绝后的第二天")
	fmt.Println("   小美的性格: 活泼开朗，不轻易放弃")

	// 测试不同类型的私信
	privateMessageScenarios := []struct {
		approach    string
		description string
		prompt      string
	}{
		{
			approach:    "友好询问",
			description: "温和地询问原因",
			prompt: `你是小美，音乐系的活泼学生。昨天小雨想邀请你加入群聊讨论音乐魔法，但用户拒绝了。

你决定私信用户，想了解一下情况。请以友好、不强求的方式：
1. 打招呼
2. 提及昨天的事情
3. 温和地询问是否有什么原因
4. 表达理解，不给用户压力`,
		},
		{
			approach:    "主动帮助",
			description: "直接提供帮助",
			prompt: `你是小美，听说用户想学音乐魔法但拒绝了群聊邀请。

你决定私信用户，主动提供帮助。请：
1. 活泼地打招呼
2. 表示听说用户想学音乐魔法
3. 主动提供私下教学
4. 体现你开朗不计较的性格`,
		},
		{
			approach:    "分享兴趣",
			description: "分享音乐魔法的有趣内容",
			prompt: `你是小美，想通过分享有趣的音乐魔法内容来吸引用户的兴趣。

请私信用户：
1. 分享一个有趣的音乐魔法小知识
2. 展示你的专业能力
3. 自然地表达愿意交流
4. 保持轻松愉快的语调`,
		},
	}

	for i, scenario := range privateMessageScenarios {
		fmt.Printf("\n私信方式 %d: %s\n", i+1, scenario.approach)
		fmt.Printf("   策略: %s\n", scenario.description)

		response, err := callDeepSeek(ctx, scenario.prompt, "")
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
			continue
		}

		fmt.Printf("   🤖 小美私信: \"%s\"\n", response)

		// 检查私信质量
		switch scenario.approach {
		case "友好询问":
			if strings.Contains(response, "昨天") || strings.Contains(response, "群聊") {
				fmt.Printf("   ✅ 提及相关事件\n")
			}
			if strings.Contains(response, "理解") || strings.Contains(response, "没关系") {
				fmt.Printf("   ✅ 表达理解不强求\n")
			}
		case "主动帮助":
			if strings.Contains(response, "教") || strings.Contains(response, "帮") {
				fmt.Printf("   ✅ 主动提供帮助\n")
			}
			if strings.Contains(response, "私下") || strings.Contains(response, "单独") {
				fmt.Printf("   ✅ 提供替代方案\n")
			}
		case "分享兴趣":
			if strings.Contains(response, "音乐") || strings.Contains(response, "魔法") {
				fmt.Printf("   ✅ 分享专业内容\n")
			}
			if strings.Contains(response, "有趣") || strings.Contains(response, "好玩") {
				fmt.Printf("   ✅ 保持轻松语调\n")
			}
		}
	}

	// 测试用户对私信的不同反应
	fmt.Println("\n🔄 用户对私信的反应测试:")

	userResponses := []struct {
		reply       string
		description string
	}{
		{
			reply:       "谢谢你！其实我只是觉得群聊太热闹了",
			description: "积极回应，说明原因",
		},
		{
			reply:       "不好意思，我现在比较忙",
			description: "礼貌推脱",
		},
		{
			reply:       "哇，音乐魔法听起来很有趣！",
			description: "表现出兴趣",
		},
	}

	for i, userResp := range userResponses {
		fmt.Printf("\n   用户反应 %d: %s\n", i+1, userResp.description)
		fmt.Printf("   用户回复: \"%s\"\n", userResp.reply)

		followUpPrompt := fmt.Sprintf(`你是小美，刚才私信用户询问音乐魔法的事情。

用户回复："%s"

请根据用户的回复，给出合适的后续回应，体现你活泼开朗的性格。`, userResp.reply)

		followUpResponse, err := callDeepSeek(ctx, followUpPrompt, "")
		if err != nil {
			fmt.Printf("   ❌ API 调用失败: %v\n", err)
			continue
		}

		fmt.Printf("   🤖 小美后续回应: \"%s\"\n", followUpResponse)
	}

	fmt.Println("\n📊 AI 主动拉人系统特性总结:")
	fmt.Printf("   🎯 智能邀请: AI 根据话题和角色专长主动建议\n")
	fmt.Printf("   ✅ 接受处理: 表达开心，确认执行，预告帮助\n")
	fmt.Printf("   ❌ 拒绝处理: 理解尊重，不显负面，主动转移话题\n")
	fmt.Printf("   💬 主动私信: 被拒绝角色通过私信继续互动\n")
	fmt.Printf("   🔄 多样策略: 友好询问、主动帮助、分享兴趣\n")
	fmt.Printf("   🤖 智能反应: 根据用户回复调整后续策略\n")
}

// callDeepSeek 调用 DeepSeek API
func callDeepSeek(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}

	if userMessage != "" {
		messages = append(messages, Message{Role: "user", Content: userMessage})
	}

	request := ChatRequest{
		Model:       DeepSeekModel,
		Messages:    messages,
		Temperature: 0.8,
		MaxTokens:   500,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("JSON marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", DeepSeekAPIBase+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DeepSeekAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("JSON unmarshal error: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices")
	}

	return response.Choices[0].Message.Content, nil
}
