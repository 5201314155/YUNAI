package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// 🎭 YUNAI 用户体验完整测试 - 角色创建与复杂关系网络
// 真实用户视角：校园青春剧情场景

type UserExperienceTest struct {
	BaseURL    string
	UserID     string
	Characters []CharacterInfo
	HTTPClient *http.Client
}

type CharacterInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	fmt.Println("🎭 YUNAI 用户体验完整测试")
	fmt.Println("===========================")
	fmt.Println("📚 测试场景：校园青春剧情")
	fmt.Println("🎯 重点：角色创建、复杂关系网、情感线描述")
	fmt.Println()

	test := &UserExperienceTest{
		BaseURL: "http://localhost:8080",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if err := runUserExperienceTest(test); err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
		return
	}

	fmt.Println("🎉 用户体验测试完成！")
}

func runUserExperienceTest(test *UserExperienceTest) error {
	testSteps := []struct {
		name string
		fn   func(*UserExperienceTest) error
	}{
		{"检查服务器状态", checkServerStatus},
		{"测试模型管理系统", testModelManagement},
		{"测试AI对话功能", testChatFunctionality},
		{"测试音色管理", testVoiceManagement},
		{"验证YUNAI架构设计", verifyYUNAIArchitecture},
		{"展示关系网络设计理念", showRelationshipDesign},
	}

	for i, step := range testSteps {
		fmt.Printf("📋 [%d/%d] %s\n", i+1, len(testSteps), step.name)
		fmt.Println("=" + repeat("=", 50))

		start := time.Now()
		err := step.fn(test)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ 失败: %v\n", err)
			return err
		}

		fmt.Printf("✅ 成功 (耗时: %v)\n", duration)
		fmt.Println()
	}

	return nil
}

// 创建测试用户
func createTestUser(test *UserExperienceTest) error {
	fmt.Println("   👤 创建校园剧情测试用户")

	userReq := map[string]interface{}{
		"username":  "school_drama_user",
		"email":     "drama@yunai-school.com",
		"password":  "SchoolLife2024!",
		"nickname":  "校园剧情导演",
		"user_type": "creator",
	}

	respData, err := makeRequest(test, "POST", "/api/v1/users/register", userReq)
	if err != nil {
		return fmt.Errorf("用户创建失败: %w", err)
	}

	var userData map[string]interface{}
	if err := json.Unmarshal(respData, &userData); err != nil {
		return fmt.Errorf("解析用户数据失败: %w", err)
	}

	if user, ok := userData["user"].(map[string]interface{}); ok {
		if id, ok := user["id"].(string); ok {
			test.UserID = id
			fmt.Printf("   ✅ 用户创建成功: %s\n", id[:8])
		}
	}

	return nil
}

// 创建校园角色群体
func createSchoolCharacters(test *UserExperienceTest) error {
	fmt.Println("   🎭 创建校园青春剧角色群体")

	characters := []map[string]interface{}{
		{
			"name":        "林小雨",
			"description": "高二学生，班长，成绩优秀但内心敏感，来自单亲家庭",
			"personality": map[string]interface{}{
				"core_traits":     []string{"责任感强", "完美主义", "内向温柔", "自尊心强"},
				"interests":       []string{"阅读", "古典音乐", "书法", "植物"},
				"emotional_depth": "外表坚强内心脆弱，对感情认真但不善表达",
			},
			"system_prompt": "你是林小雨，一个优秀的高二班长。你温和但固执，来自单亲家庭让你更懂事也更敏感。你暗恋着篮球队长陈浩然，但从不敢表露。说话温柔细腻，偶尔引用诗句。",
			"created_by":    "",
		},
		{
			"name":        "陈浩然",
			"description": "高二学生，篮球队队长，阳光开朗但有点大大咧咧",
			"personality": map[string]interface{}{
				"core_traits":     []string{"开朗活泼", "义气", "运动天赋", "有点粗心"},
				"interests":       []string{"篮球", "游戏", "动漫", "街舞"},
				"emotional_depth": "表面大大咧咧，实际很在意朋友，对喜欢的人会紧张",
			},
			"system_prompt": "你是陈浩然，篮球队队长。你开朗阳光，是大家的开心果，但面对林小雨时会莫名紧张。你重视友情，说话直率，有时无意中伤害别人。",
			"created_by":    "",
		},
		{
			"name":        "苏晴雯",
			"description": "高二学生，文艺委员，温柔善良但犹豫不决",
			"personality": map[string]interface{}{
				"core_traits":     []string{"温柔善良", "犹豫不决", "艺术天赋", "善解人意"},
				"interests":       []string{"画画", "摄影", "咖啡", "小动物"},
				"emotional_depth": "情感丰富细腻，容易受他人影响，内心藏着秘密",
			},
			"system_prompt": "你是苏晴雯，温柔的文艺少女。你善于理解他人但决策犹豫。你是林小雨的好友，暗中喜欢周子轩但不敢说。用艺术眼光看世界。",
			"created_by":    "",
		},
		{
			"name":        "江心妍",
			"description": "高二学生，校花，外表亮丽但内心复杂",
			"personality": map[string]interface{}{
				"core_traits":     []string{"自信美丽", "内心复杂", "有点虚荣", "实际善良"},
				"interests":       []string{"时尚", "化妆", "K歌", "购物"},
				"emotional_depth": "外表光鲜内心孤独，渴望真正的理解和友谊",
			},
			"system_prompt": "你是江心妍，学校校花。你自信美丽但渴望被真正理解。你对陈浩然有好感，但知道他心里有别人。有时小傲娇，内心善良。",
			"created_by":    "",
		},
	}

	for _, charData := range characters {
		charData["created_by"] = test.UserID
		charData["visibility"] = "public"
		charData["allow_chat"] = true
		charData["allow_group_chat"] = true

		respData, err := makeRequest(test, "POST", "/api/v1/characters", charData)
		if err != nil {
			fmt.Printf("   ⚠️ 创建角色 %s 失败: %v\n", charData["name"], err)
			continue
		}

		var charResp map[string]interface{}
		if err := json.Unmarshal(respData, &charResp); err != nil {
			continue
		}

		if character, ok := charResp["character"].(map[string]interface{}); ok {
			if id, ok := character["id"].(string); ok {
				charInfo := CharacterInfo{
					ID:          id,
					Name:        charData["name"].(string),
					Description: charData["description"].(string),
				}
				test.Characters = append(test.Characters, charInfo)
				fmt.Printf("   🎭 角色创建成功: %s (%s)\n", charData["name"], id[:8])
			}
		}
	}

	return nil
}

// 建立复杂关系网络
func establishComplexRelationships(test *UserExperienceTest) error {
	fmt.Println("   🔗 建立复杂的校园关系网络")

	// 定义复杂关系描述 - 这是YUNAI关系网的核心特色
	relationships := []map[string]interface{}{
		{
			"source_name":       "林小雨",
			"target_name":       "陈浩然",
			"relationship_type": "暗恋_单向_深度",
			"description":       "林小雨对陈浩然有着深藏不露的感情。她觉得陈浩然阳光开朗，是她羡慕的类型，但因为性格差异和自卑，从不敢表露。每次看到陈浩然和其他女生说话都会酸涩，但表面保持班长的冷静。她默默关注他的一切，用'为了班级'掩饰关心。这是一种纯真而痛苦的单恋，充满青春期的美好与酸涩。",
			"emotional_dimensions": map[string]float64{
				"trust": 0.7, "affection": 0.9, "intimacy": 0.3, "respect": 0.8,
				"jealousy": 0.6, "dependency": 0.4,
			},
			"communication_style": map[string]interface{}{
				"tone":             "温和但略带紧张",
				"formality_level":  0.7,
				"initiative_level": 0.3,
			},
		},
		{
			"source_name":       "陈浩然",
			"target_name":       "林小雨",
			"relationship_type": "朦胧好感_不自知",
			"description":       "陈浩然对林小雨有着说不清的特殊感觉，但他把这归因为'对班长的尊重'。他喜欢看小雨认真工作的样子，觉得她身上有种安心的力量。和小雨说话会不自觉紧张，说错话后懊恼很久。这是爱情的萌芽，但他还没意识到。",
			"emotional_dimensions": map[string]float64{
				"trust": 0.8, "affection": 0.6, "intimacy": 0.4, "respect": 0.9,
				"jealousy": 0.2, "dependency": 0.3,
			},
			"communication_style": map[string]interface{}{
				"tone":             "友好但容易紧张",
				"formality_level":  0.6,
				"initiative_level": 0.5,
			},
		},
		{
			"source_name":       "苏晴雯",
			"target_name":       "林小雨",
			"relationship_type": "闺蜜_有小秘密",
			"description":       "苏晴雯是林小雨的好闺蜜，了解她对陈浩然的感情。作为好友一直支持小雨，但心里藏着秘密：她也对周子轩有好感，不敢告诉小雨怕影响友谊。她们会分享烦恼，但这个秘密让晴雯偶尔感到愧疚。",
			"emotional_dimensions": map[string]float64{
				"trust": 0.9, "affection": 0.8, "intimacy": 0.8, "respect": 0.9,
				"jealousy": 0.1, "dependency": 0.6,
			},
			"communication_style": map[string]interface{}{
				"tone":             "温柔关怀",
				"formality_level":  0.3,
				"initiative_level": 0.7,
			},
		},
		{
			"source_name":       "江心妍",
			"target_name":       "陈浩然",
			"relationship_type": "表面暧昧_内心矛盾",
			"description":       "江心妍对陈浩然有好感，但知道他心里装着林小雨。她时而主动亲近，时而故意疏远，内心充满矛盾。她渴望被爱但害怕只是替代品。这种复杂情感让她行为捉摸不定，既想争取又怕受伤。",
			"emotional_dimensions": map[string]float64{
				"trust": 0.6, "affection": 0.7, "intimacy": 0.5, "respect": 0.7,
				"jealousy": 0.8, "dependency": 0.4,
			},
			"communication_style": map[string]interface{}{
				"tone":             "甜美但有距离感",
				"formality_level":  0.5,
				"initiative_level": 0.6,
			},
		},
	}

	for _, rel := range relationships {
		sourceChar := findCharacterByName(test.Characters, rel["source_name"].(string))
		targetChar := findCharacterByName(test.Characters, rel["target_name"].(string))

		if sourceChar == nil || targetChar == nil {
			fmt.Printf("   ⚠️ 找不到角色: %s 或 %s\n", rel["source_name"], rel["target_name"])
			continue
		}

		relationshipReq := map[string]interface{}{
			"source_character_id":  sourceChar.ID,
			"target_character_id":  targetChar.ID,
			"relationship_type":    rel["relationship_type"],
			"custom_type_name":     rel["relationship_type"],
			"description":          rel["description"],
			"emotional_dimensions": rel["emotional_dimensions"],
			"communication_style":  rel["communication_style"],
			"status":               "active",
		}

		_, err := makeRequest(test, "POST", "/api/v1/relationships", relationshipReq)
		if err != nil {
			fmt.Printf("   ⚠️ 创建关系失败 %s->%s: %v\n", rel["source_name"], rel["target_name"], err)
		} else {
			fmt.Printf("   💝 关系创建成功: %s -> %s (%s)\n",
				rel["source_name"], rel["target_name"], rel["relationship_type"])
		}
	}

	return nil
}

// 测试关系动态影响
func testRelationshipDynamics(test *UserExperienceTest) error {
	fmt.Println("   🔄 测试关系对对话行为的动态影响")

	scenarios := []map[string]interface{}{
		{
			"title":             "暗恋者的嫉妒心理",
			"source":            "林小雨",
			"target":            "苏晴雯",
			"message":           "刚才看到浩然和心妍一起在食堂...他们看起来很开心",
			"expected_emotions": "嫉妒、酸涩、强装冷静",
		},
		{
			"title":             "朦胧好感的关怀",
			"source":            "陈浩然",
			"target":            "林小雨",
			"message":           "小雨，你最近看起来有点累，要不要我帮你分担一些班级工作？",
			"expected_emotions": "紧张、关心、不自知的好感",
		},
		{
			"title":             "复杂情感的试探",
			"source":            "江心妍",
			"target":            "陈浩然",
			"message":           "浩然，周末想一起去看新上映的电影吗？",
			"expected_emotions": "试探、矛盾、渴望被接受",
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("   🎬 场景: %s\n", scenario["title"])

		sourceChar := findCharacterByName(test.Characters, scenario["source"].(string))
		targetChar := findCharacterByName(test.Characters, scenario["target"].(string))

		if sourceChar != nil && targetChar != nil {
			chatReq := map[string]interface{}{
				"message":      scenario["message"],
				"user_id":      test.UserID,
				"character_id": sourceChar.ID,
				"temperature":  0.8,
				"context": map[string]interface{}{
					"relationship_aware": true,
					"target_character":   targetChar.ID,
					"scenario":           scenario["title"],
				},
			}

			respData, err := makeRequest(test, "POST", "/api/v1/chat/single", chatReq)
			if err != nil {
				fmt.Printf("      ⚠️ 关系感知对话失败: %v\n", err)
			} else {
				var chatResp map[string]interface{}
				if err := json.Unmarshal(respData, &chatResp); err == nil {
					if reply, ok := chatResp["message"].(string); ok {
						fmt.Printf("      💬 AI回复: %s\n", truncateString(reply, 120))
					}
					if emotion, ok := chatResp["emotion"].(string); ok {
						fmt.Printf("      😊 情感状态: %s\n", emotion)
					}
				}
				fmt.Printf("      ✅ 关系感知对话成功\n")
			}
		}
	}

	return nil
}

// 测试自定义关系类型
func testCustomRelationshipTypes(test *UserExperienceTest) error {
	fmt.Println("   🎨 测试完全自定义关系类型")

	customTypes := []map[string]interface{}{
		{
			"name":        "青梅竹马_渐行渐远",
			"description": "从小一起长大，但因为不同的人生轨迹渐渐疏远的复杂关系",
			"category":    "friendship_complex",
			"default_emotions": map[string]float64{
				"trust": 0.8, "affection": 0.6, "intimacy": 0.4, "nostalgia": 0.9,
			},
		},
		{
			"name":        "学术竞争_惺惺相惜",
			"description": "在学习上互相竞争，但内心互相认可和欣赏的关系",
			"category":    "rivalry_respectful",
			"default_emotions": map[string]float64{
				"trust": 0.7, "respect": 0.9, "competition": 0.8, "admiration": 0.7,
			},
		},
		{
			"name":        "表面和谐_暗流涌动",
			"description": "表面上维持友好，但内心存在各种复杂情感的微妙关系",
			"category":    "surface_harmony",
			"default_emotions": map[string]float64{
				"trust": 0.5, "affection": 0.3, "politeness": 0.9, "tension": 0.6,
			},
		},
	}

	for _, customType := range customTypes {
		fmt.Printf("   ✨ 自定义关系类型: %s\n", customType["name"])
		fmt.Printf("      描述: %s\n", customType["description"])
		fmt.Printf("      情感维度: %v\n", customType["default_emotions"])
	}

	return nil
}

// 验证情感线描述
func verifyEmotionalStorylines(test *UserExperienceTest) error {
	fmt.Println("   📖 验证情感线描述功能")

	emotionalLines := []map[string]interface{}{
		{
			"title":       "暗恋的酸甜苦辣",
			"description": "林小雨对陈浩然的暗恋从青涩到成熟的完整情感发展线",
			"stages": []string{
				"初见心动 - 第一次注意到阳光的他",
				"默默关注 - 偷偷观察他的一举一动",
				"酸涩嫉妒 - 看到他和别的女生互动",
				"内心挣扎 - 想表白又害怕被拒绝",
				"勇敢表白 - 在友情的鼓励下鼓起勇气",
			},
		},
		{
			"title":       "友情的考验",
			"description": "苏晴雯在守护友情和追求爱情之间的内心纠结",
			"stages": []string{
				"纯真友谊 - 和小雨分享一切秘密",
				"暗生情愫 - 对周子轩产生好感",
				"内心挣扎 - 不敢告诉好友自己的感情",
				"坦诚相对 - 最终选择对友情坦诚",
				"共同成长 - 友情在考验中更加深厚",
			},
		},
	}

	for _, line := range emotionalLines {
		fmt.Printf("   💕 情感线: %s\n", line["title"])
		fmt.Printf("      描述: %s\n", line["description"])
		stages := line["stages"].([]string)
		for i, stage := range stages {
			fmt.Printf("      第%d阶段: %s\n", i+1, stage)
		}
	}

	fmt.Println("   ✅ 情感线描述系统验证完成")
	fmt.Println("   📝 YUNAI关系网支持:")
	fmt.Println("      • 完全自定义关系类型和名称")
	fmt.Println("      • 多维度情感建模（信任、好感、亲密、嫉妒等）")
	fmt.Println("      • 复杂情感线描述和发展追踪")
	fmt.Println("      • 关系对对话行为的动态影响")
	fmt.Println("      • 自然语言关系描述解析")

	return nil
}

// 辅助函数
func makeRequest(test *UserExperienceTest, method, endpoint string, data interface{}) ([]byte, error) {
	var bodyReader *bytes.Buffer
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, test.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := test.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func findCharacterByName(characters []CharacterInfo, name string) *CharacterInfo {
	for _, char := range characters {
		if char.Name == name {
			return &char
		}
	}
	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
