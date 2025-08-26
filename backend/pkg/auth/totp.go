package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TOTPConfig TOTP 配置
type TOTPConfig struct {
	Issuer      string // 发行者名称
	AccountName string // 账户名称
	Secret      string // 密钥
	Period      int    // 时间周期（秒）
	Digits      int    // 验证码位数
	Algorithm   string // 算法
}

// TOTPManager TOTP 管理器
type TOTPManager struct {
	issuer string
	period int
	digits int
}

// NewTOTPManager 创建 TOTP 管理器
func NewTOTPManager(issuer string) *TOTPManager {
	return &TOTPManager{
		issuer: issuer,
		period: 30, // 默认 30 秒
		digits: 6,  // 默认 6 位数字
	}
}

// GenerateSecret 生成 TOTP 密钥
func (m *TOTPManager) GenerateSecret() (string, error) {
	// 生成 20 字节的随机密钥
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}
	
	// 使用 base32 编码
	return base32.StdEncoding.EncodeToString(secret), nil
}

// GenerateQRCodeURL 生成二维码 URL
func (m *TOTPManager) GenerateQRCodeURL(accountName, secret string) string {
	config := &TOTPConfig{
		Issuer:      m.issuer,
		AccountName: accountName,
		Secret:      secret,
		Period:      m.period,
		Digits:      m.digits,
		Algorithm:   "SHA1",
	}
	
	return m.buildOTPAuthURL(config)
}

// buildOTPAuthURL 构建 OTP Auth URL
func (m *TOTPManager) buildOTPAuthURL(config *TOTPConfig) string {
	v := url.Values{}
	v.Set("secret", config.Secret)
	v.Set("issuer", config.Issuer)
	v.Set("algorithm", config.Algorithm)
	v.Set("digits", strconv.Itoa(config.Digits))
	v.Set("period", strconv.Itoa(config.Period))
	
	return fmt.Sprintf("otpauth://totp/%s:%s?%s",
		url.QueryEscape(config.Issuer),
		url.QueryEscape(config.AccountName),
		v.Encode())
}

// GenerateCode 生成 TOTP 验证码
func (m *TOTPManager) GenerateCode(secret string, timestamp time.Time) (string, error) {
	// 解码 base32 密钥
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}
	
	// 计算时间步长
	timeStep := timestamp.Unix() / int64(m.period)
	
	// 生成验证码
	code := m.generateHOTP(key, timeStep)
	
	// 格式化为指定位数
	format := fmt.Sprintf("%%0%dd", m.digits)
	return fmt.Sprintf(format, code), nil
}

// ValidateCode 验证 TOTP 验证码
func (m *TOTPManager) ValidateCode(secret, code string, timestamp time.Time) bool {
	// 允许前后一个时间窗口的误差（总共 3 个窗口）
	for i := -1; i <= 1; i++ {
		testTime := timestamp.Add(time.Duration(i) * time.Duration(m.period) * time.Second)
		expectedCode, err := m.GenerateCode(secret, testTime)
		if err != nil {
			continue
		}
		
		if code == expectedCode {
			return true
		}
	}
	
	return false
}

// generateHOTP 生成 HOTP 验证码
func (m *TOTPManager) generateHOTP(key []byte, counter int64) int {
	// 将计数器转换为 8 字节大端序
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))
	
	// 使用 HMAC-SHA1 计算哈希
	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)
	
	// 动态截取
	offset := hash[19] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff
	
	// 取模得到指定位数的验证码
	return int(truncated % uint32(math.Pow10(m.digits)))
}

// GetCurrentCode 获取当前时间的验证码
func (m *TOTPManager) GetCurrentCode(secret string) (string, error) {
	return m.GenerateCode(secret, time.Now())
}

// ValidateCurrentCode 验证当前时间的验证码
func (m *TOTPManager) ValidateCurrentCode(secret, code string) bool {
	return m.ValidateCode(secret, code, time.Now())
}

// GetTimeRemaining 获取当前验证码剩余有效时间
func (m *TOTPManager) GetTimeRemaining() time.Duration {
	now := time.Now()
	nextPeriod := ((now.Unix() / int64(m.period)) + 1) * int64(m.period)
	return time.Duration(nextPeriod-now.Unix()) * time.Second
}

// GetPeriod 获取时间周期
func (m *TOTPManager) GetPeriod() int {
	return m.period
}

// GetDigits 获取验证码位数
func (m *TOTPManager) GetDigits() int {
	return m.digits
}

// SetPeriod 设置时间周期
func (m *TOTPManager) SetPeriod(period int) {
	if period > 0 {
		m.period = period
	}
}

// SetDigits 设置验证码位数
func (m *TOTPManager) SetDigits(digits int) {
	if digits >= 4 && digits <= 8 {
		m.digits = digits
	}
}

// TOTPSetup TOTP 设置信息
type TOTPSetup struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
	ManualKey string `json:"manual_key"`
	Issuer    string `json:"issuer"`
	Account   string `json:"account"`
}

// SetupTOTP 设置 TOTP
func (m *TOTPManager) SetupTOTP(accountName string) (*TOTPSetup, error) {
	secret, err := m.GenerateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}
	
	qrCodeURL := m.GenerateQRCodeURL(accountName, secret)
	
	// 格式化手动输入密钥（每 4 个字符一组）
	manualKey := formatManualKey(secret)
	
	return &TOTPSetup{
		Secret:    secret,
		QRCodeURL: qrCodeURL,
		ManualKey: manualKey,
		Issuer:    m.issuer,
		Account:   accountName,
	}, nil
}

// formatManualKey 格式化手动输入密钥
func formatManualKey(secret string) string {
	var formatted strings.Builder
	for i, char := range secret {
		if i > 0 && i%4 == 0 {
			formatted.WriteString(" ")
		}
		formatted.WriteRune(char)
	}
	return formatted.String()
}

// ValidateSecret 验证密钥格式
func ValidateSecret(secret string) bool {
	// 移除空格并转换为大写
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	
	// 检查是否为有效的 base32 编码
	if _, err := base32.StdEncoding.DecodeString(secret); err != nil {
		return false
	}
	
	// 检查长度（通常为 16 或 32 字符）
	if len(secret) != 16 && len(secret) != 32 {
		return false
	}
	
	return true
}

// NormalizeSecret 标准化密钥格式
func NormalizeSecret(secret string) string {
	// 移除空格并转换为大写
	return strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
}
