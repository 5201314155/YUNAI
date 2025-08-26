package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CardGenerator 卡密生成器
type CardGenerator struct {
	prefix string // 卡密前缀，如 "668899"
	length int    // 卡密总长度，如 16
}

// NewCardGenerator 创建卡密生成器
func NewCardGenerator(prefix string, length int) *CardGenerator {
	return &CardGenerator{
		prefix: prefix,
		length: length,
	}
}

// GenerateCardCode 生成单个卡密
func (g *CardGenerator) GenerateCardCode() (string, error) {
	if len(g.prefix) >= g.length {
		return "", fmt.Errorf("prefix length cannot be greater than or equal to total length")
	}

	// 计算需要生成的随机数字位数
	remainingLength := g.length - len(g.prefix)

	// 生成随机数字
	randomPart, err := g.generateRandomDigits(remainingLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate random digits: %w", err)
	}

	// 组合前缀和随机部分
	cardCode := g.prefix + randomPart

	return cardCode, nil
}

// GenerateBatchCardCodes 批量生成卡密
func (g *CardGenerator) GenerateBatchCardCodes(count int) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}

	cardCodes := make([]string, 0, count)
	usedCodes := make(map[string]bool)

	for len(cardCodes) < count {
		code, err := g.GenerateCardCode()
		if err != nil {
			return nil, err
		}

		// 确保不重复
		if !usedCodes[code] {
			cardCodes = append(cardCodes, code)
			usedCodes[code] = true
		}
	}

	return cardCodes, nil
}

// generateRandomDigits 生成指定长度的随机数字字符串
func (g *CardGenerator) generateRandomDigits(length int) (string, error) {
	if length <= 0 {
		return "", nil
	}

	// 计算最大值（10^length - 1）
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	max.Sub(max, big.NewInt(1))

	// 生成随机数
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// 转换为字符串并补零
	result := fmt.Sprintf("%0*s", length, n.String())

	return result, nil
}

// ValidateCardCode 验证卡密格式
func (g *CardGenerator) ValidateCardCode(cardCode string) bool {
	// 检查长度
	if len(cardCode) != g.length {
		return false
	}

	// 检查是否全为数字
	if !isAllDigits(cardCode) {
		return false
	}

	// 检查前缀
	if !strings.HasPrefix(cardCode, g.prefix) {
		return false
	}

	return true
}

// FormatCardCode 格式化卡密显示（添加空格分隔）
func (g *CardGenerator) FormatCardCode(cardCode string) string {
	if len(cardCode) != g.length {
		return cardCode
	}

	// 每4位添加一个空格
	var formatted strings.Builder
	for i, char := range cardCode {
		if i > 0 && i%4 == 0 {
			formatted.WriteString(" ")
		}
		formatted.WriteRune(char)
	}

	return formatted.String()
}

// UnformatCardCode 去除卡密中的格式字符（空格、横线等）
func (g *CardGenerator) UnformatCardCode(cardCode string) string {
	// 移除所有非数字字符
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(cardCode, "")
}

// GenerateCardCodeWithChecksum 生成带校验位的卡密
func (g *CardGenerator) GenerateCardCodeWithChecksum() (string, error) {
	if g.length < len(g.prefix)+2 {
		return "", fmt.Errorf("length too short for checksum")
	}

	// 生成主体部分（预留1位校验位）
	mainLength := g.length - len(g.prefix) - 1
	mainPart, err := g.generateRandomDigits(mainLength)
	if err != nil {
		return "", err
	}

	// 组合前缀和主体
	codeWithoutChecksum := g.prefix + mainPart

	// 计算校验位
	checksum := g.calculateChecksum(codeWithoutChecksum)

	// 组合完整卡密
	fullCode := codeWithoutChecksum + strconv.Itoa(checksum)

	return fullCode, nil
}

// ValidateCardCodeWithChecksum 验证带校验位的卡密
func (g *CardGenerator) ValidateCardCodeWithChecksum(cardCode string) bool {
	if !g.ValidateCardCode(cardCode) {
		return false
	}

	// 分离主体和校验位
	mainPart := cardCode[:len(cardCode)-1]
	checksumStr := cardCode[len(cardCode)-1:]

	// 解析校验位
	checksum, err := strconv.Atoi(checksumStr)
	if err != nil {
		return false
	}

	// 验证校验位
	expectedChecksum := g.calculateChecksum(mainPart)
	return checksum == expectedChecksum
}

// calculateChecksum 计算校验位（使用 Luhn 算法的简化版本）
func (g *CardGenerator) calculateChecksum(code string) int {
	sum := 0
	for i, char := range code {
		digit, _ := strconv.Atoi(string(char))
		if i%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}
		sum += digit
	}
	return (10 - (sum % 10)) % 10
}

// isAllDigits 检查字符串是否全为数字
func isAllDigits(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// GetCardInfo 解析卡密信息
func (g *CardGenerator) GetCardInfo(cardCode string) map[string]interface{} {
	info := make(map[string]interface{})

	info["card_code"] = cardCode
	info["formatted"] = g.FormatCardCode(cardCode)
	info["valid"] = g.ValidateCardCode(cardCode)
	info["prefix"] = g.prefix
	info["length"] = g.length

	if len(cardCode) == g.length && strings.HasPrefix(cardCode, g.prefix) {
		info["prefix_match"] = true
		info["random_part"] = cardCode[len(g.prefix):]
	} else {
		info["prefix_match"] = false
	}

	return info
}

// YunaiCardType YUNAI卡片类型
type YunaiCardType struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"` // 卡号前缀
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`

	// 卡片属性
	DiscountRate float64  `json:"discount_rate,omitempty"` // 折扣率 (0.9 = 9折)
	BonusRate    float64  `json:"bonus_rate,omitempty"`    // 返利率 (0.2 = 20%返利)
	Privileges   []string `json:"privileges,omitempty"`    // 特权列表
}

// YunaiCardGenerator YUNAI专用卡号生成器
type YunaiCardGenerator struct {
	cardType YunaiCardType
}

// NewYunaiCardGenerator 创建YUNAI专用卡号生成器
func NewYunaiCardGenerator(cardType YunaiCardType) *YunaiCardGenerator {
	return &YunaiCardGenerator{
		cardType: cardType,
	}
}

// GenerateYunaiCard 生成YUNAI专用卡号
// 格式：前缀(2位) + 年份后2位 + 月份2位 + 随机数(10位) = 16位纯数字
func (g *YunaiCardGenerator) GenerateYunaiCard() (string, error) {
	now := time.Now()

	// 卡片前缀 + 年份后2位 + 月份2位
	prefix := fmt.Sprintf("%s%02d%02d", g.cardType.Prefix, now.Year()%100, int(now.Month()))

	// 生成10位随机数
	randomPart, err := generateRandomDigits(10)
	if err != nil {
		return "", fmt.Errorf("failed to generate random digits: %w", err)
	}

	cardNumber := prefix + randomPart
	return cardNumber, nil
}

// ValidateYunaiCard 验证YUNAI卡号格式
func (g *YunaiCardGenerator) ValidateYunaiCard(cardNumber string) bool {
	// 检查长度
	if len(cardNumber) != 16 {
		return false
	}

	// 检查前缀
	if !strings.HasPrefix(cardNumber, g.cardType.Prefix) {
		return false
	}

	// 检查格式：前缀 + 4位年月 + 10位数字
	pattern := fmt.Sprintf(`^%s\d{14}$`, g.cardType.Prefix)
	matched, err := regexp.MatchString(pattern, cardNumber)
	if err != nil || !matched {
		return false
	}

	// 验证年月是否合理
	prefixLen := len(g.cardType.Prefix)
	yearMonth := cardNumber[prefixLen : prefixLen+4]
	year, err1 := strconv.Atoi(yearMonth[:2])
	month, err2 := strconv.Atoi(yearMonth[2:4])

	if err1 != nil || err2 != nil {
		return false
	}

	// 年份范围：20-99 (2020-2099)
	if year < 20 || year > 99 {
		return false
	}

	// 月份范围：01-12
	if month < 1 || month > 12 {
		return false
	}

	return true
}

// FormatYunaiCard 格式化显示YUNAI卡号
func (g *YunaiCardGenerator) FormatYunaiCard(cardNumber string) string {
	if len(cardNumber) != 16 {
		return cardNumber
	}

	// 8825 0812 3456 7890
	return fmt.Sprintf("%s %s %s %s",
		cardNumber[0:4],   // 8825
		cardNumber[4:8],   // 0812
		cardNumber[8:12],  // 3456
		cardNumber[12:16]) // 7890
}

// MaskYunaiCard 脱敏显示YUNAI卡号
func (g *YunaiCardGenerator) MaskYunaiCard(cardNumber string) string {
	if len(cardNumber) != 16 {
		return cardNumber
	}

	// 8825********7890
	return cardNumber[0:4] + "********" + cardNumber[12:16]
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateRandomDigits 生成指定长度的随机数字字符串
func generateRandomDigits(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	digits := make([]byte, length)
	for i := 0; i < length; i++ {
		randomNum, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate random digit: %w", err)
		}
		digits[i] = byte('0' + randomNum.Int64())
	}

	return string(digits), nil
}

// GetYunaiCardTypes 获取所有YUNAI卡片类型
func GetYunaiCardTypes() []YunaiCardType {
	return []YunaiCardType{
		{
			Code:        "standard",
			Name:        "标准卡",
			Prefix:      "88",
			Description: "YUNAI标准充值卡",
			Color:       "#1890ff",
			Icon:        "💳",
		},
		{
			Code:         "discount",
			Name:         "折扣卡",
			Prefix:       "89",
			Description:  "享受充值折扣优惠",
			Color:        "#52c41a",
			Icon:         "🎫",
			DiscountRate: 0.9, // 9折
		},
		{
			Code:        "bonus",
			Name:        "返利卡",
			Prefix:      "87",
			Description: "充值获得额外返利",
			Color:       "#faad14",
			Icon:        "💰",
			BonusRate:   0.2, // 20%返利
		},
		{
			Code:         "vip",
			Name:         "VIP卡",
			Prefix:       "86",
			Description:  "VIP专属卡，尊享特权",
			Color:        "#722ed1",
			Icon:         "👑",
			DiscountRate: 0.8, // 8折
			BonusRate:    0.1, // 10%返利
			Privileges:   []string{"priority_support", "exclusive_models", "advanced_features"},
		},
		{
			Code:         "creator",
			Name:         "创作者卡",
			Prefix:       "85",
			Description:  "创作者专用卡，创作无限",
			Color:        "#f759ab",
			Icon:         "🎨",
			DiscountRate: 0.85, // 8.5折
			Privileges:   []string{"unlimited_generation", "commercial_license", "priority_queue"},
		},
		{
			Code:         "enterprise",
			Name:         "企业卡",
			Prefix:       "84",
			Description:  "企业级服务卡",
			Color:        "#13c2c2",
			Icon:         "🏢",
			DiscountRate: 0.7,  // 7折
			BonusRate:    0.15, // 15%返利
			Privileges:   []string{"bulk_operations", "api_access", "dedicated_support"},
		},
	}
}

// GetYunaiCardTypeByCode 根据代码获取卡片类型
func GetYunaiCardTypeByCode(code string) (YunaiCardType, bool) {
	types := GetYunaiCardTypes()
	for _, cardType := range types {
		if cardType.Code == code {
			return cardType, true
		}
	}
	return YunaiCardType{}, false
}

// DefaultYunaiCardGenerator 创建默认的 YUNAI 卡号生成器（标准卡）
func DefaultYunaiCardGenerator() *YunaiCardGenerator {
	standardType, _ := GetYunaiCardTypeByCode("standard")
	return NewYunaiCardGenerator(standardType)
}

// NewYunaiCardGeneratorByType 根据卡片类型代码创建生成器
func NewYunaiCardGeneratorByType(typeCode string) (*YunaiCardGenerator, error) {
	cardType, exists := GetYunaiCardTypeByCode(typeCode)
	if !exists {
		return nil, fmt.Errorf("unknown card type: %s", typeCode)
	}
	return NewYunaiCardGenerator(cardType), nil
}

// GenerateYunaiCardCode 生成 YUNAI 卡号（便捷函数）
func GenerateYunaiCardCode() (string, error) {
	generator := DefaultYunaiCardGenerator()
	return generator.GenerateYunaiCard()
}

// ValidateYunaiCardCode 验证 YUNAI 卡号（便捷函数）
func ValidateYunaiCardCode(cardCode string) bool {
	generator := DefaultYunaiCardGenerator()
	return generator.ValidateYunaiCard(cardCode)
}

// FormatYunaiCardCode 格式化 YUNAI 卡号显示（便捷函数）
func FormatYunaiCardCode(cardCode string) string {
	generator := DefaultYunaiCardGenerator()
	return generator.FormatYunaiCard(cardCode)
}
