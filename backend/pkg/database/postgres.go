package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// PostgreSQLConfig PostgreSQL 配置
type PostgreSQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// PostgreSQLClient PostgreSQL 客户端
type PostgreSQLClient struct {
	DB     *sqlx.DB
	config *PostgreSQLConfig
	logger *logrus.Logger
}

// NewPostgreSQL 创建新的 PostgreSQL 连接
func NewPostgreSQL() (*PostgreSQLClient, error) {
	config := &PostgreSQLConfig{}
	
	// 从配置文件读取数据库配置
	if err := viper.UnmarshalKey("database", config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal database config: %w", err)
	}
	
	// 设置默认值
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.Name == "" {
		config.Name = "yunai_dev"
	}
	if config.User == "" {
		config.User = "postgres"
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 25
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 5 * time.Minute
	}
	
	// 构建连接字符串
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Name, config.SSLMode)
	
	// 创建数据库连接
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 配置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	
	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	logger := logrus.New()
	logger.Info("PostgreSQL connection established successfully")
	
	return &PostgreSQLClient{
		DB:     db,
		config: config,
		logger: logger,
	}, nil
}

// Close 关闭数据库连接
func (c *PostgreSQLClient) Close() error {
	if c.DB != nil {
		c.logger.Info("Closing PostgreSQL connection")
		return c.DB.Close()
	}
	return nil
}

// Health 检查数据库健康状态
func (c *PostgreSQLClient) Health() error {
	return c.DB.Ping()
}

// GetStats 获取连接池统计信息
func (c *PostgreSQLClient) GetStats() sql.DBStats {
	return c.DB.Stats()
}

// BeginTx 开始事务
func (c *PostgreSQLClient) BeginTx() (*sqlx.Tx, error) {
	return c.DB.Beginx()
}

// WithTransaction 在事务中执行函数
func (c *PostgreSQLClient) WithTransaction(fn func(*sqlx.Tx) error) error {
	tx, err := c.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	
	err = fn(tx)
	return err
}

// Migrate 执行数据库迁移
func (c *PostgreSQLClient) Migrate(migrationPath string) error {
	// 这里可以集成 golang-migrate 或其他迁移工具
	c.logger.Infof("Migration path: %s", migrationPath)
	// TODO: 实现迁移逻辑
	return nil
}

// LogSlowQuery 记录慢查询
func (c *PostgreSQLClient) LogSlowQuery(query string, duration time.Duration) {
	threshold := viper.GetDuration("database.slow_query_threshold")
	if threshold == 0 {
		threshold = time.Second
	}
	
	if duration > threshold {
		c.logger.WithFields(logrus.Fields{
			"query":    query,
			"duration": duration,
		}).Warn("Slow query detected")
	}
}

// ExecuteWithLogging 执行查询并记录日志
func (c *PostgreSQLClient) ExecuteWithLogging(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := c.DB.Exec(query, args...)
	duration := time.Since(start)
	
	c.LogSlowQuery(query, duration)
	
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"query":    query,
			"args":     args,
			"duration": duration,
			"error":    err,
		}).Error("Query execution failed")
	}
	
	return result, err
}

// QueryWithLogging 查询并记录日志
func (c *PostgreSQLClient) QueryWithLogging(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := c.DB.Query(query, args...)
	duration := time.Since(start)
	
	c.LogSlowQuery(query, duration)
	
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"query":    query,
			"args":     args,
			"duration": duration,
			"error":    err,
		}).Error("Query execution failed")
	}
	
	return rows, err
}

// GetConnection 获取原始数据库连接
func (c *PostgreSQLClient) GetConnection() *sqlx.DB {
	return c.DB
}

// GetConfig 获取数据库配置
func (c *PostgreSQLClient) GetConfig() *PostgreSQLConfig {
	return c.config
}
