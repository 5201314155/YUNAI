package domain

import (
	"time"

	"github.com/google/uuid"
)

// BackupRecord 备份记录
type BackupRecord struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Type         string     `json:"type" gorm:"not null"` // full, incremental, differential
	Status       string     `json:"status" gorm:"default:'running'"` // running, completed, failed
	FilePath     string     `json:"file_path"`
	FileSize     int64      `json:"file_size" gorm:"default:0"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	BaseBackupID *uuid.UUID `json:"base_backup_id" gorm:"type:uuid"` // 用于增量备份
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// BackupSchedule 备份调度
type BackupSchedule struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name           string     `json:"name" gorm:"not null"`
	BackupType     string     `json:"backup_type" gorm:"not null"` // full, incremental
	CronExpression string     `json:"cron_expression" gorm:"not null"`
	IsEnabled      bool       `json:"is_enabled" gorm:"default:true"`
	LastRun        *time.Time `json:"last_run"`
	NextRun        *time.Time `json:"next_run"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// BackupValidation 备份验证
type BackupValidation struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BackupID     uuid.UUID  `json:"backup_id" gorm:"type:uuid;not null"`
	Status       string     `json:"status" gorm:"default:'running'"` // running, completed, failed
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
}

// RestoreResult 恢复结果
type RestoreResult struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BackupID     uuid.UUID  `json:"backup_id" gorm:"type:uuid;not null"`
	Status       string     `json:"status" gorm:"default:'running'"` // running, completed, failed
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
}

// RestoreTest 恢复测试
type RestoreTest struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BackupID     uuid.UUID  `json:"backup_id" gorm:"type:uuid;not null"`
	Status       string     `json:"status" gorm:"default:'running'"` // running, completed, failed
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
}

// RestoreOptions 恢复选项
type RestoreOptions struct {
	TargetDatabase string            `json:"target_database"`
	OverwriteData  bool              `json:"overwrite_data"`
	SkipTables     []string          `json:"skip_tables"`
	OnlyTables     []string          `json:"only_tables"`
	CustomOptions  map[string]string `json:"custom_options"`
}

// BackupPolicy 备份策略
type BackupPolicy struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name              string    `json:"name" gorm:"not null"`
	FullBackupCron    string    `json:"full_backup_cron" gorm:"not null"`
	IncrementalCron   string    `json:"incremental_cron"`
	RetentionDays     int       `json:"retention_days" gorm:"default:30"`
	MaxBackups        int       `json:"max_backups" gorm:"default:10"`
	CompressionLevel  int       `json:"compression_level" gorm:"default:6"`
	EncryptionEnabled bool      `json:"encryption_enabled" gorm:"default:false"`
	ValidationEnabled bool      `json:"validation_enabled" gorm:"default:true"`
	IsActive          bool      `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// BackupStats 备份统计
type BackupStats struct {
	TotalBackups      int     `json:"total_backups"`
	SuccessfulBackups int     `json:"successful_backups"`
	FailedBackups     int     `json:"failed_backups"`
	TotalSize         int64   `json:"total_size"`
	LastBackupTime    *time.Time `json:"last_backup_time"`
	LastBackupStatus  string  `json:"last_backup_status"`
	AvgBackupTime     float64 `json:"avg_backup_time"` // 分钟
	SuccessRate       float64 `json:"success_rate"`
}

// BackupHealth 备份健康状态
type BackupHealth struct {
	Status    string    `json:"status"` // healthy, warning, critical
	CheckTime time.Time `json:"check_time"`
	Issues    []string  `json:"issues"`
}
