package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// New 创建新的日志记录器
func New() *logrus.Logger {
	logger := logrus.New()
	
	// 设置输出
	logger.SetOutput(os.Stdout)
	
	// 设置日志级别
	logger.SetLevel(logrus.InfoLevel)
	
	// 设置格式
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	
	return logger
}

// NewWithLevel 创建指定级别的日志记录器
func NewWithLevel(level logrus.Level) *logrus.Logger {
	logger := New()
	logger.SetLevel(level)
	return logger
}
