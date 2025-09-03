package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	FromName string `json:"from_name"`
}

// EmailTemplate 邮件模板
type EmailTemplate struct {
	Subject string
	HTML    string
	Text    string
}

// EmailService 邮件服务接口
type EmailService interface {
	SendVerificationEmail(to, username, code string) error
	SendPasswordResetEmail(to, username, resetURL string) error
	SendWelcomeEmail(to, username string) error
	SendNotificationEmail(to, subject, content string) error
}

// smtpService SMTP邮件服务实现
type smtpService struct {
	config    *SMTPConfig
	templates map[string]*EmailTemplate
	logger    *logrus.Logger
}

// NewSMTPService 创建SMTP邮件服务
func NewSMTPService(config *SMTPConfig, logger *logrus.Logger) EmailService {
	service := &smtpService{
		config:    config,
		templates: make(map[string]*EmailTemplate),
		logger:    logger,
	}

	// 初始化邮件模板
	service.initTemplates()

	return service
}

// initTemplates 初始化邮件模板
func (s *smtpService) initTemplates() {
	// 邮箱验证模板
	s.templates["verification"] = &EmailTemplate{
		Subject: "YUNAI - 邮箱验证码",
		HTML: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>YUNAI 邮箱验证</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #4A90E2;">🤖 YUNAI</h1>
            <p style="color: #666;">AI社交平台</p>
        </div>
        
        <h2>邮箱验证码</h2>
        <p>亲爱的 {{.Username}}，</p>
        <p>您正在进行邮箱验证，验证码为：</p>
        
        <div style="background: #f8f9fa; border: 2px dashed #4A90E2; padding: 20px; text-align: center; margin: 20px 0;">
            <h1 style="color: #4A90E2; font-size: 32px; margin: 0; letter-spacing: 5px;">{{.Code}}</h1>
        </div>
        
        <p><strong>验证码有效期：10分钟</strong></p>
        <p>如果这不是您的操作，请忽略此邮件。</p>
        
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #666; font-size: 12px;">
            此邮件由 YUNAI 系统自动发送，请勿回复。<br>
            © 2025 YUNAI AI社交平台
        </p>
    </div>
</body>
</html>`,
		Text: `YUNAI 邮箱验证

亲爱的 {{.Username}}，

您正在进行邮箱验证，验证码为：{{.Code}}

验证码有效期：10分钟

如果这不是您的操作，请忽略此邮件。

© 2025 YUNAI AI社交平台`,
	}

	// 密码重置模板
	s.templates["password_reset"] = &EmailTemplate{
		Subject: "YUNAI - 密码重置",
		HTML: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>YUNAI 密码重置</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #4A90E2;">🤖 YUNAI</h1>
            <p style="color: #666;">AI社交平台</p>
        </div>
        
        <h2>密码重置</h2>
        <p>亲爱的 {{.Username}}，</p>
        <p>您请求重置YUNAI账户密码。请点击下面的链接重置密码：</p>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}" style="background: #4A90E2; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">
                重置密码
            </a>
        </div>
        
        <p><strong>链接有效期：30分钟</strong></p>
        <p>如果按钮无法点击，请复制以下链接到浏览器：</p>
        <p style="word-break: break-all; background: #f8f9fa; padding: 10px; border-radius: 3px;">{{.ResetURL}}</p>
        
        <p>如果这不是您的操作，请忽略此邮件，您的密码不会被更改。</p>
        
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #666; font-size: 12px;">
            此邮件由 YUNAI 系统自动发送，请勿回复。<br>
            © 2025 YUNAI AI社交平台
        </p>
    </div>
</body>
</html>`,
		Text: `YUNAI 密码重置

亲爱的 {{.Username}}，

您请求重置YUNAI账户密码。请访问以下链接重置密码：

{{.ResetURL}}

链接有效期：30分钟

如果这不是您的操作，请忽略此邮件，您的密码不会被更改。

© 2025 YUNAI AI社交平台`,
	}

	// 欢迎邮件模板
	s.templates["welcome"] = &EmailTemplate{
		Subject: "欢迎加入YUNAI AI社交平台！",
		HTML: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>欢迎加入YUNAI</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #4A90E2;">🤖 YUNAI</h1>
            <p style="color: #666;">AI社交平台</p>
        </div>
        
        <h2>欢迎加入YUNAI！</h2>
        <p>亲爱的 {{.Username}}，</p>
        <p>欢迎加入YUNAI AI社交平台！您即将体验到前所未有的AI社交乐趣。</p>
        
        <div style="background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3 style="color: #4A90E2; margin-top: 0;">🎭 您可以体验：</h3>
            <ul>
                <li>🤖 与AI角色进行真实对话</li>
                <li>👥 创建复杂的关系网络</li>
                <li>📱 智能朋友圈互动</li>
                <li>🎵 语音通话功能</li>
                <li>💰 完整的支付钱包系统</li>
            </ul>
        </div>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="https://yunai.com/app" style="background: #4A90E2; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">
                开始体验
            </a>
        </div>
        
        <p>如有任何问题，请联系我们的客服团队。</p>
        
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #666; font-size: 12px;">
            此邮件由 YUNAI 系统自动发送，请勿回复。<br>
            © 2025 YUNAI AI社交平台
        </p>
    </div>
</body>
</html>`,
		Text: `欢迎加入YUNAI！

亲爱的 {{.Username}}，

欢迎加入YUNAI AI社交平台！您即将体验到前所未有的AI社交乐趣。

您可以体验：
- 与AI角色进行真实对话
- 创建复杂的关系网络
- 智能朋友圈互动
- 语音通话功能
- 完整的支付钱包系统

访问 https://yunai.com/app 开始体验

如有任何问题，请联系我们的客服团队。

© 2025 YUNAI AI社交平台`,
	}
}

// SendVerificationEmail 发送验证邮件
func (s *smtpService) SendVerificationEmail(to, username, code string) error {
	template := s.templates["verification"]
	if template == nil {
		return fmt.Errorf("verification email template not found")
	}

	data := map[string]string{
		"Username": username,
		"Code":     code,
	}

	return s.sendEmail(to, template, data)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *smtpService) SendPasswordResetEmail(to, username, resetURL string) error {
	template := s.templates["password_reset"]
	if template == nil {
		return fmt.Errorf("password reset email template not found")
	}

	data := map[string]string{
		"Username": username,
		"ResetURL": resetURL,
	}

	return s.sendEmail(to, template, data)
}

// SendWelcomeEmail 发送欢迎邮件
func (s *smtpService) SendWelcomeEmail(to, username string) error {
	template := s.templates["welcome"]
	if template == nil {
		return fmt.Errorf("welcome email template not found")
	}

	data := map[string]string{
		"Username": username,
	}

	return s.sendEmail(to, template, data)
}

// SendNotificationEmail 发送通知邮件
func (s *smtpService) SendNotificationEmail(to, subject, content string) error {
	// 构建简单的通知邮件
	template := &EmailTemplate{
		Subject: subject,
		HTML:    fmt.Sprintf("<p>%s</p>", content),
		Text:    content,
	}

	return s.sendEmail(to, template, nil)
}

// sendEmail 发送邮件
func (s *smtpService) sendEmail(to string, emailTemplate *EmailTemplate, data map[string]string) error {
	// 渲染模板
	subject, err := s.renderTemplate(emailTemplate.Subject, data)
	if err != nil {
		return fmt.Errorf("failed to render subject: %w", err)
	}

	htmlBody, err := s.renderTemplate(emailTemplate.HTML, data)
	if err != nil {
		return fmt.Errorf("failed to render HTML body: %w", err)
	}

	textBody, err := s.renderTemplate(emailTemplate.Text, data)
	if err != nil {
		return fmt.Errorf("failed to render text body: %w", err)
	}

	// 构建邮件内容
	message := s.buildMessage(to, subject, htmlBody, textBody)

	// 发送邮件
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	err = smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(message))
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"to":      to,
			"subject": subject,
			"error":   err,
		}).Error("Failed to send email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
	}).Info("Email sent successfully")

	return nil
}

// renderTemplate 渲染模板
func (s *smtpService) renderTemplate(templateStr string, data map[string]string) (string, error) {
	if data == nil {
		return templateStr, nil
	}

	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// buildMessage 构建邮件消息
func (s *smtpService) buildMessage(to, subject, htmlBody, textBody string) string {
	from := s.config.From
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)
	}

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "multipart/alternative; boundary=\"boundary123\""

	// 构建邮件体
	var message strings.Builder

	// 添加头部
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")

	// 添加文本部分
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// 添加HTML部分
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	// 结束边界
	message.WriteString("--boundary123--\r\n")

	return message.String()
}

// MockEmailService 模拟邮件服务（用于测试）
type MockEmailService struct {
	logger *logrus.Logger
	emails []MockEmail
}

// MockEmail 模拟邮件
type MockEmail struct {
	To      string    `json:"to"`
	Subject string    `json:"subject"`
	Content string    `json:"content"`
	SentAt  time.Time `json:"sent_at"`
}

// NewMockEmailService 创建模拟邮件服务
func NewMockEmailService(logger *logrus.Logger) EmailService {
	return &MockEmailService{
		logger: logger,
		emails: make([]MockEmail, 0),
	}
}

// SendVerificationEmail 发送验证邮件（模拟）
func (m *MockEmailService) SendVerificationEmail(to, username, code string) error {
	email := MockEmail{
		To:      to,
		Subject: "YUNAI - 邮箱验证码",
		Content: fmt.Sprintf("验证码：%s", code),
		SentAt:  time.Now(),
	}

	m.emails = append(m.emails, email)
	
	m.logger.WithFields(logrus.Fields{
		"to":       to,
		"username": username,
		"code":     code,
	}).Info("📧 Mock email sent: verification code")

	fmt.Printf("📧 [模拟邮件] 发送验证码到 %s: %s\n", to, code)
	return nil
}

// SendPasswordResetEmail 发送密码重置邮件（模拟）
func (m *MockEmailService) SendPasswordResetEmail(to, username, resetURL string) error {
	email := MockEmail{
		To:      to,
		Subject: "YUNAI - 密码重置",
		Content: fmt.Sprintf("重置链接：%s", resetURL),
		SentAt:  time.Now(),
	}

	m.emails = append(m.emails, email)
	
	m.logger.WithFields(logrus.Fields{
		"to":       to,
		"username": username,
		"resetURL": resetURL,
	}).Info("📧 Mock email sent: password reset")

	fmt.Printf("📧 [模拟邮件] 发送密码重置到 %s: %s\n", to, resetURL)
	return nil
}

// SendWelcomeEmail 发送欢迎邮件（模拟）
func (m *MockEmailService) SendWelcomeEmail(to, username string) error {
	email := MockEmail{
		To:      to,
		Subject: "欢迎加入YUNAI！",
		Content: fmt.Sprintf("欢迎 %s 加入YUNAI AI社交平台！", username),
		SentAt:  time.Now(),
	}

	m.emails = append(m.emails, email)
	
	m.logger.WithFields(logrus.Fields{
		"to":       to,
		"username": username,
	}).Info("📧 Mock email sent: welcome")

	fmt.Printf("📧 [模拟邮件] 发送欢迎邮件到 %s\n", to)
	return nil
}

// SendNotificationEmail 发送通知邮件（模拟）
func (m *MockEmailService) SendNotificationEmail(to, subject, content string) error {
	email := MockEmail{
		To:      to,
		Subject: subject,
		Content: content,
		SentAt:  time.Now(),
	}

	m.emails = append(m.emails, email)
	
	m.logger.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
	}).Info("📧 Mock email sent: notification")

	fmt.Printf("📧 [模拟邮件] 发送通知到 %s: %s\n", to, subject)
	return nil
}

// GetSentEmails 获取已发送的邮件（仅用于测试）
func (m *MockEmailService) GetSentEmails() []MockEmail {
	return m.emails
}
