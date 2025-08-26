package auth

import (
	"strings"
	"testing"
)

func TestCardGenerator_GenerateCardCode(t *testing.T) {
	generator := NewCardGenerator("668899", 16)
	
	// 测试生成单个卡密
	cardCode, err := generator.GenerateCardCode()
	if err != nil {
		t.Fatalf("Failed to generate card code: %v", err)
	}
	
	// 验证长度
	if len(cardCode) != 16 {
		t.Errorf("Expected card code length 16, got %d", len(cardCode))
	}
	
	// 验证前缀
	if !strings.HasPrefix(cardCode, "668899") {
		t.Errorf("Expected card code to start with '668899', got %s", cardCode)
	}
	
	// 验证是否全为数字
	if !isAllDigits(cardCode) {
		t.Errorf("Card code should contain only digits, got %s", cardCode)
	}
	
	t.Logf("Generated card code: %s", cardCode)
}

func TestCardGenerator_GenerateBatchCardCodes(t *testing.T) {
	generator := NewCardGenerator("668899", 16)
	
	// 测试批量生成
	count := 10
	cardCodes, err := generator.GenerateBatchCardCodes(count)
	if err != nil {
		t.Fatalf("Failed to generate batch card codes: %v", err)
	}
	
	// 验证数量
	if len(cardCodes) != count {
		t.Errorf("Expected %d card codes, got %d", count, len(cardCodes))
	}
	
	// 验证唯一性
	uniqueCodes := make(map[string]bool)
	for _, code := range cardCodes {
		if uniqueCodes[code] {
			t.Errorf("Duplicate card code found: %s", code)
		}
		uniqueCodes[code] = true
		
		// 验证每个卡密的格式
		if !generator.ValidateCardCode(code) {
			t.Errorf("Invalid card code generated: %s", code)
		}
	}
	
	t.Logf("Generated %d unique card codes", len(cardCodes))
}

func TestCardGenerator_ValidateCardCode(t *testing.T) {
	generator := NewCardGenerator("668899", 16)
	
	testCases := []struct {
		cardCode string
		expected bool
		desc     string
	}{
		{"6688990012345678", true, "valid card code"},
		{"6688990087654321", true, "another valid card code"},
		{"668899001234567", false, "too short"},
		{"66889900123456789", false, "too long"},
		{"6688990012345abc", false, "contains letters"},
		{"1234560012345678", false, "wrong prefix"},
		{"", false, "empty string"},
		{"668899-0012-3456-78", false, "contains hyphens"},
	}
	
	for _, tc := range testCases {
		result := generator.ValidateCardCode(tc.cardCode)
		if result != tc.expected {
			t.Errorf("ValidateCardCode(%s) = %v, expected %v (%s)", 
				tc.cardCode, result, tc.expected, tc.desc)
		}
	}
}

func TestCardGenerator_FormatCardCode(t *testing.T) {
	generator := NewCardGenerator("668899", 16)
	
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"6688990012345678", "6688 9900 1234 5678", "normal formatting"},
		{"1234567890123456", "1234 5678 9012 3456", "another normal case"},
		{"123", "123", "too short, no formatting"},
		{"", "", "empty string"},
	}
	
	for _, tc := range testCases {
		result := generator.FormatCardCode(tc.input)
		if result != tc.expected {
			t.Errorf("FormatCardCode(%s) = %s, expected %s (%s)", 
				tc.input, result, tc.expected, tc.desc)
		}
	}
}

func TestCardGenerator_UnformatCardCode(t *testing.T) {
	generator := NewCardGenerator("668899", 16)
	
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"6688 9900 1234 5678", "6688990012345678", "remove spaces"},
		{"6688-9900-1234-5678", "6688990012345678", "remove hyphens"},
		{"6688_9900_1234_5678", "6688990012345678", "remove underscores"},
		{"6688990012345678", "6688990012345678", "no formatting to remove"},
		{"6688 9900-1234_5678", "6688990012345678", "mixed formatting"},
		{"", "", "empty string"},
	}
	
	for _, tc := range testCases {
		result := generator.UnformatCardCode(tc.input)
		if result != tc.expected {
			t.Errorf("UnformatCardCode(%s) = %s, expected %s (%s)", 
				tc.input, result, tc.expected, tc.desc)
		}
	}
}

func TestCardGenerator_GenerateCardCodeWithChecksum(t *testing.T) {
	generator := NewCardGenerator("668899", 16)
	
	// 生成带校验位的卡密
	cardCode, err := generator.GenerateCardCodeWithChecksum()
	if err != nil {
		t.Fatalf("Failed to generate card code with checksum: %v", err)
	}
	
	// 验证基本格式
	if !generator.ValidateCardCode(cardCode) {
		t.Errorf("Generated card code with checksum is invalid: %s", cardCode)
	}
	
	// 验证校验位
	if !generator.ValidateCardCodeWithChecksum(cardCode) {
		t.Errorf("Checksum validation failed for: %s", cardCode)
	}
	
	t.Logf("Generated card code with checksum: %s", cardCode)
}

func TestDefaultYunaiCardGenerator(t *testing.T) {
	generator := DefaultYunaiCardGenerator()
	
	// 验证默认配置
	if generator.prefix != "668899" {
		t.Errorf("Expected default prefix '668899', got '%s'", generator.prefix)
	}
	
	if generator.length != 16 {
		t.Errorf("Expected default length 16, got %d", generator.length)
	}
	
	// 测试便捷函数
	cardCode, err := GenerateYunaiCardCode()
	if err != nil {
		t.Fatalf("GenerateYunaiCardCode failed: %v", err)
	}
	
	if !ValidateYunaiCardCode(cardCode) {
		t.Errorf("Generated YUNAI card code is invalid: %s", cardCode)
	}
	
	formatted := FormatYunaiCardCode(cardCode)
	if !strings.Contains(formatted, " ") {
		t.Errorf("Formatted card code should contain spaces: %s", formatted)
	}
	
	t.Logf("YUNAI card code: %s", cardCode)
	t.Logf("Formatted: %s", formatted)
}

func TestCardGenerator_GetCardInfo(t *testing.T) {
	generator := NewCardGenerator("668899", 16)
	cardCode := "6688990012345678"
	
	info := generator.GetCardInfo(cardCode)
	
	// 验证返回的信息
	if info["card_code"] != cardCode {
		t.Errorf("Expected card_code %s, got %v", cardCode, info["card_code"])
	}
	
	if info["valid"] != true {
		t.Errorf("Expected valid to be true, got %v", info["valid"])
	}
	
	if info["prefix"] != "668899" {
		t.Errorf("Expected prefix '668899', got %v", info["prefix"])
	}
	
	if info["length"] != 16 {
		t.Errorf("Expected length 16, got %v", info["length"])
	}
	
	if info["prefix_match"] != true {
		t.Errorf("Expected prefix_match to be true, got %v", info["prefix_match"])
	}
	
	if info["random_part"] != "0012345678" {
		t.Errorf("Expected random_part '0012345678', got %v", info["random_part"])
	}
	
	t.Logf("Card info: %+v", info)
}

func BenchmarkCardGenerator_GenerateCardCode(b *testing.B) {
	generator := NewCardGenerator("668899", 16)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateCardCode()
		if err != nil {
			b.Fatalf("Failed to generate card code: %v", err)
		}
	}
}

func BenchmarkCardGenerator_ValidateCardCode(b *testing.B) {
	generator := NewCardGenerator("668899", 16)
	cardCode := "6688990012345678"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ValidateCardCode(cardCode)
	}
}
