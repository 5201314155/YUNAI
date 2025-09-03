package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// YUNAI 项目全功能测试程序
const (
	BASE_URL = "http://localhost:8092"
	TIMEOUT  = 30 * time.Second
)

type TestResult struct {
	Name     string
	Success  bool
	Message  string
	Duration time.Duration
}

type TestSuite struct {
	client  *http.Client
	results []TestResult
}

func main() {
	fmt.Println("🚀 YUNAI 项目全功能测试")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println()

	suite := &TestSuite{
		client: &http.Client{Timeout: TIMEOUT},
	}

	suite.runAllTests()
	suite.printReport()
}

func (ts *TestSuite) runAllTests() {
	fmt.Println("📋 开始执行全功能测试...")
	fmt.Println()

	// 基础功能测试
	ts.testBasicFunctions()

	// 核心业务功能测试
	ts.testWorldSystem()
	ts.testRelationshipNetwork()
	ts.testMomentsFunctions()
	ts.testSmartMomentsFunctions()
	ts.testGroupChatFunctions()
	ts.testAIInvitationSystem()

	// 语音和媒体功能测试
	ts.testVoiceFunctions()

	// 支付和模型管理测试
	ts.testPaymentSystem()
	ts.testModelManagement()

	// AI和提示词系统测试
	ts.testGlobalPromptSystem()

	// 故事系统测试
	ts.testStoryTriggers()

	// 高级功能测试
	ts.testAnalyticsSystem()
	ts.testNotificationSystem()
	ts.testCharacterReviewSystem()
	ts.testUserIdentitySystem()
	ts.testBackupSystem()
	ts.testSupportSystem()
}

func (ts *TestSuite) testBasicFunctions() {
	fmt.Println("🔧 测试基础功能...")

	ts.runTest("健康检查", func() (bool, string) {
		resp, err := ts.client.Get(BASE_URL + "/health")
		if err != nil {
			return false, fmt.Sprintf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return false, fmt.Sprintf("状态码错误: %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "ok") {
			return false, "响应内容错误"
		}

		return true, "服务正常运行"
	})

	ts.runTest("API文档", func() (bool, string) {
		resp, err := ts.client.Get(BASE_URL + "/api/docs")
		if err != nil {
			return false, fmt.Sprintf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return false, fmt.Sprintf("状态码错误: %d", resp.StatusCode)
		}

		return true, "API文档可访问"
	})
}

func (ts *TestSuite) testWorldSystem() {
	fmt.Println("🌍 测试世界观系统...")

	ts.runTest("创建世界观", func() (bool, string) {
		worldData := map[string]interface{}{
			"name":             "魔法学院",
			"description":      "一个充满魔法的学院世界",
			"background_story": "在这个世界里，魔法是日常生活的一部分...",
		}
		return ts.postRequest("/worlds", worldData)
	})

	ts.runTest("获取世界观列表", func() (bool, string) {
		return ts.getRequest("/worlds")
	})
}

func (ts *TestSuite) testRelationshipNetwork() {
	fmt.Println("🕸️ 测试关系网系统...")

	ts.runTest("获取角色关系网络", func() (bool, string) {
		return ts.getRequest("/relationships/network/550e8400-e29b-41d4-a716-446655440000")
	})
}

func (ts *TestSuite) testMomentsFunctions() {
	fmt.Println("📱 测试朋友圈功能...")

	ts.runTest("智能生成朋友圈", func() (bool, string) {
		momentData := map[string]interface{}{
			"character_id": "550e8400-e29b-41d4-a716-446655440000",
			"count":        2,
		}
		return ts.postRequest("/smart-moments/generate", momentData)
	})

	ts.runTest("获取朋友圈列表", func() (bool, string) {
		return ts.getRequest("/moments")
	})
}

func (ts *TestSuite) testGroupChatFunctions() {
	fmt.Println("👥 测试群聊功能...")

	ts.runTest("创建群聊", func() (bool, string) {
		groupData := map[string]interface{}{
			"name":        "魔法学院讨论组",
			"description": "讨论魔法学习心得",
			"creator_id":  "user-123",
			"max_members": 50,
		}
		return ts.postRequest("/group-chats", groupData)
	})

	ts.runTest("获取群聊列表", func() (bool, string) {
		return ts.getRequest("/group-chats")
	})

	ts.runTest("添加群成员", func() (bool, string) {
		memberData := map[string]interface{}{
			"character_id": "char-456",
			"user_id":      "user-123",
		}
		return ts.postRequest("/group-chats/group-123/members", memberData)
	})

	ts.runTest("发送群消息", func() (bool, string) {
		messageData := map[string]interface{}{
			"sender_character_id": "char-456",
			"content":             "大家好！我是小雨",
			"mentioned_users":     []string{"user-123"},
		}
		return ts.postRequest("/group-chats/group-123/messages", messageData)
	})
}

func (ts *TestSuite) testVoiceFunctions() {
	fmt.Println("🎵 测试语音功能...")

	ts.runTest("语音通话", func() (bool, string) {
		voiceData := map[string]interface{}{
			"character_id": "char-456",
			"user_id":      "user-123",
		}
		return ts.postRequest("/voice/call", voiceData)
	})
}

func (ts *TestSuite) testPaymentSystem() {
	fmt.Println("💰 测试支付系统...")

	ts.runTest("获取充值套餐", func() (bool, string) {
		return ts.getRequest("/api/v1/payment/packages")
	})

	ts.runTest("绑定支付卡片", func() (bool, string) {
		// 使用随机卡号避免冲突
		cardNumber := fmt.Sprintf("1234567890%06d", time.Now().UnixNano()%1000000)
		cardData := map[string]interface{}{
			"card_number": cardNumber,
			"card_holder": "Test User",
		}
		return ts.postRequest("/api/v1/payment/cards", cardData)
	})
}

func (ts *TestSuite) testModelManagement() {
	fmt.Println("🧠 测试模型管理...")

	ts.runTest("获取可用模型", func() (bool, string) {
		return ts.getRequest("/models/")
	})
}

func (ts *TestSuite) testGlobalPromptSystem() {
	fmt.Println("🌐 测试全局提示词系统...")

	ts.runTest("生成全局提示词", func() (bool, string) {
		promptData := map[string]interface{}{
			"character_id":  "char-456",
			"user_id":       "user-123",
			"function_type": "chat",
		}
		return ts.postRequest("/global-prompt/generate", promptData)
	})
}

func (ts *TestSuite) testStoryTriggers() {
	fmt.Println("📖 测试故事触发器...")

	ts.runTest("创建故事章节", func() (bool, string) {
		chapterData := map[string]interface{}{
			"title":    "图书馆的秘密",
			"content":  "在古老的图书馆深处，隐藏着一个神秘的魔法阵...",
			"world_id": "world-123",
		}
		return ts.postRequest("/chapters", chapterData)
	})

	ts.runTest("获取章节列表", func() (bool, string) {
		return ts.getRequest("/chapters")
	})

	ts.runTest("创建触发器", func() (bool, string) {
		triggerData := map[string]interface{}{
			"name": "进入图书馆触发器",
		}
		return ts.postRequest("/triggers", triggerData)
	})

	ts.runTest("获取触发器列表", func() (bool, string) {
		return ts.getRequest("/triggers")
	})
}

func (ts *TestSuite) runTest(name string, testFunc func() (bool, string)) {
	start := time.Now()
	success, message := testFunc()
	duration := time.Since(start)

	result := TestResult{
		Name:     name,
		Success:  success,
		Message:  message,
		Duration: duration,
	}

	ts.results = append(ts.results, result)

	status := "✅"
	if !success {
		status = "❌"
	}

	fmt.Printf("  %s %s (%v) - %s\n", status, name, duration, message)
}

func (ts *TestSuite) postRequest(endpoint string, data interface{}) (bool, string) {
	jsonData, _ := json.Marshal(data)
	resp, err := ts.client.Post(BASE_URL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, "请求成功"
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Sprintf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
}

func (ts *TestSuite) getRequest(endpoint string) (bool, string) {
	resp, err := ts.client.Get(BASE_URL + endpoint)
	if err != nil {
		return false, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, "请求成功"
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Sprintf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
}

// 打印测试报告
func (ts *TestSuite) printReport() {
	fmt.Println()
	fmt.Println("📊 测试报告")
	fmt.Println("=" + strings.Repeat("=", 80))

	totalTests := len(ts.results)
	successCount := 0
	failureCount := 0
	totalDuration := time.Duration(0)

	for _, result := range ts.results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
		totalDuration += result.Duration
	}

	fmt.Printf("📈 总体统计:\n")
	fmt.Printf("   总测试数: %d\n", totalTests)
	fmt.Printf("   成功: %d (%.1f%%)\n", successCount, float64(successCount)/float64(totalTests)*100)
	fmt.Printf("   失败: %d (%.1f%%)\n", failureCount, float64(failureCount)/float64(totalTests)*100)
	fmt.Printf("   总耗时: %v\n", totalDuration)
	fmt.Printf("   平均耗时: %v\n", totalDuration/time.Duration(totalTests))
	fmt.Println()

	if failureCount > 0 {
		fmt.Println("❌ 失败的测试:")
		for _, result := range ts.results {
			if !result.Success {
				fmt.Printf("   • %s: %s\n", result.Name, result.Message)
			}
		}
		fmt.Println()
	}

	fmt.Println("🎯 测试建议:")
	if successCount == totalTests {
		fmt.Println("   🎉 所有测试通过！系统功能完整，可以进行下一步开发。")
	} else if float64(successCount)/float64(totalTests) >= 0.8 {
		fmt.Println("   👍 大部分功能正常，建议修复失败的测试项。")
	} else {
		fmt.Println("   ⚠️  需要重点关注失败的功能模块，建议优先修复核心功能。")
	}

	// 保存详细日志
	ts.saveDetailedLog()
}

// 保存详细测试日志
func (ts *TestSuite) saveDetailedLog() {
	logContent := fmt.Sprintf("YUNAI 项目全功能测试报告\n")
	logContent += fmt.Sprintf("测试时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	logContent += fmt.Sprintf("=" + strings.Repeat("=", 50) + "\n\n")

	for i, result := range ts.results {
		status := "PASS"
		if !result.Success {
			status = "FAIL"
		}

		logContent += fmt.Sprintf("%d. [%s] %s (%v)\n", i+1, status, result.Name, result.Duration)
		if !result.Success {
			logContent += fmt.Sprintf("   错误: %s\n", result.Message)
		}
		logContent += "\n"
	}

	// 写入文件
	os.WriteFile("test_results.log", []byte(logContent), 0644)
	fmt.Println("📝 详细测试日志已保存到 test_results.log")
}

// 智能朋友圈功能测试
func (ts *TestSuite) testSmartMomentsFunctions() {
	fmt.Println("🤖 测试智能朋友圈功能...")

	ts.runTest("基于聊天生成朋友圈", func() (bool, string) {
		momentData := map[string]interface{}{
			"character_id":  "550e8400-e29b-41d4-a716-446655440001", // 使用正确的角色ID
			"based_on_chat": true,
			"auto_interact": false,
			"count":         1,
		}
		return ts.postRequest("/smart-moments/generate", momentData)
	})

	ts.runTest("智能生成朋友圈", func() (bool, string) {
		momentData := map[string]interface{}{
			"character_id": "550e8400-e29b-41d4-a716-446655440001", // 使用正确的角色ID
			"count":        2,
		}
		return ts.postRequest("/smart-moments/generate", momentData)
	})

	ts.runTest("获取生成上下文", func() (bool, string) {
		return ts.getRequest("/smart-moments/generation-context/550e8400-e29b-41d4-a716-446655440001") // 使用正确的角色ID
	})

	ts.runTest("触发自动互动", func() (bool, string) {
		interactionData := map[string]interface{}{
			"interaction_type": "like",
		}
		return ts.postRequest("/smart-moments/550e8400-e29b-41d4-a716-446655440001/auto-interact", interactionData) // 使用正确的角色ID
	})

	ts.runTest("处理@提及", func() (bool, string) {
		mentionData := map[string]interface{}{
			"mentioned_characters": []string{"550e8400-e29b-41d4-a716-446655440001"},
		}
		return ts.postRequest("/smart-moments/550e8400-e29b-41d4-a716-446655440001/process-mentions", mentionData) // 使用正确的角色ID
	})
}

// AI邀请系统测试
func (ts *TestSuite) testAIInvitationSystem() {
	fmt.Println("🎯 测试AI邀请系统...")

	ts.runTest("分析邀请场景", func() (bool, string) {
		analyzeData := map[string]interface{}{
			"user_id":      "550e8400-e29b-41d4-a716-446655440000",
			"context_type": "social",
		}
		return ts.postRequest("/ai-invitation/analyze", analyzeData)
	})

	ts.runTest("寻找匹配角色", func() (bool, string) {
		matchData := map[string]interface{}{
			"user_id":     "550e8400-e29b-41d4-a716-446655440000",
			"preferences": []string{"friendly", "intelligent"},
		}
		return ts.postRequest("/ai-invitation/find-matches", matchData)
	})

	ts.runTest("获取邀请建议", func() (bool, string) {
		suggestionData := map[string]interface{}{
			"user_id":      "550e8400-e29b-41d4-a716-446655440000",
			"character_id": "550e8400-e29b-41d4-a716-446655440001",
		}
		return ts.postRequest("/ai-invitation/suggestions", suggestionData)
	})

	ts.runTest("执行邀请", func() (bool, string) {
		executeData := map[string]interface{}{
			"user_id":      "550e8400-e29b-41d4-a716-446655440000",
			"character_id": "550e8400-e29b-41d4-a716-446655440001",
			"message":      "你好，很高兴认识你！",
		}
		return ts.postRequest("/ai-invitation/execute", executeData)
	})
}

// 分析系统测试
func (ts *TestSuite) testAnalyticsSystem() {
	fmt.Println("📊 测试分析系统...")

	ts.runTest("获取用户分析", func() (bool, string) {
		return ts.getRequest("/analytics/users/550e8400-e29b-41d4-a716-446655440000")
	})

	ts.runTest("获取系统统计", func() (bool, string) {
		return ts.getRequest("/analytics/system/stats")
	})

	ts.runTest("获取活跃度报告", func() (bool, string) {
		return ts.getRequest("/analytics/activity/report")
	})
}

// 通知系统测试
func (ts *TestSuite) testNotificationSystem() {
	fmt.Println("🔔 测试通知系统...")

	ts.runTest("获取通知列表", func() (bool, string) {
		return ts.getRequest("/notifications")
	})

	ts.runTest("创建通知", func() (bool, string) {
		notificationData := map[string]interface{}{
			"user_id": "550e8400-e29b-41d4-a716-446655440000",
			"title":   "测试通知",
			"content": "这是一个测试通知",
			"type":    "system",
		}
		return ts.postRequest("/notifications", notificationData)
	})

	ts.runTest("标记通知已读", func() (bool, string) {
		return ts.postRequest("/notifications/550e8400-e29b-41d4-a716-446655440000/read", nil)
	})
}

// 角色审核系统测试
func (ts *TestSuite) testCharacterReviewSystem() {
	fmt.Println("👤 测试角色审核系统...")

	ts.runTest("获取待审核角色", func() (bool, string) {
		return ts.getRequest("/character-review/pending")
	})

	ts.runTest("提交角色审核", func() (bool, string) {
		reviewData := map[string]interface{}{
			"character_id": "550e8400-e29b-41d4-a716-446655440000",
			"status":       "approved",
			"comments":     "角色设定合理",
		}
		return ts.postRequest("/character-review/submit", reviewData)
	})
}

// 用户身份系统测试
func (ts *TestSuite) testUserIdentitySystem() {
	fmt.Println("🆔 测试用户身份系统...")

	ts.runTest("获取用户身份", func() (bool, string) {
		return ts.getRequest("/user-identity/550e8400-e29b-41d4-a716-446655440000")
	})

	ts.runTest("更新用户身份", func() (bool, string) {
		identityData := map[string]interface{}{
			"user_id":    "550e8400-e29b-41d4-a716-446655440000",
			"identity":   "premium_user",
			"attributes": map[string]interface{}{"level": 5},
		}
		return ts.postRequest("/user-identity/update", identityData)
	})
}

// 备份系统测试
func (ts *TestSuite) testBackupSystem() {
	fmt.Println("💾 测试备份系统...")

	ts.runTest("创建备份", func() (bool, string) {
		backupData := map[string]interface{}{
			"backup_type": "full",
			"description": "全量备份测试",
		}
		return ts.postRequest("/backup/create", backupData)
	})

	ts.runTest("获取备份列表", func() (bool, string) {
		return ts.getRequest("/backup/list")
	})

	ts.runTest("获取备份状态", func() (bool, string) {
		return ts.getRequest("/backup/status")
	})
}

// 客户支持系统测试
func (ts *TestSuite) testSupportSystem() {
	fmt.Println("🎧 测试客户支持系统...")

	ts.runTest("创建工单", func() (bool, string) {
		ticketData := map[string]interface{}{
			"user_id":     "550e8400-e29b-41d4-a716-446655440000",
			"title":       "测试工单",
			"description": "这是一个测试工单",
			"priority":    "medium",
		}
		return ts.postRequest("/support/tickets", ticketData)
	})

	ts.runTest("获取工单列表", func() (bool, string) {
		return ts.getRequest("/support/tickets")
	})

	ts.runTest("获取知识库文章", func() (bool, string) {
		return ts.getRequest("/support/knowledge/articles")
	})
}
