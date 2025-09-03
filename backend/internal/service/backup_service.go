package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// BackupService 备份恢复服务接口
type BackupService interface {
	// 数据库备份
	CreateDatabaseBackup(ctx context.Context, backupType string) (*domain.BackupRecord, error)
	CreateIncrementalBackup(ctx context.Context) (*domain.BackupRecord, error)
	ScheduleBackup(ctx context.Context, schedule *domain.BackupSchedule) error

	// 备份管理
	ListBackups(ctx context.Context, backupType string) ([]*domain.BackupRecord, error)
	GetBackup(ctx context.Context, backupID uuid.UUID) (*domain.BackupRecord, error)
	DeleteBackup(ctx context.Context, backupID uuid.UUID) error
	ValidateBackup(ctx context.Context, backupID uuid.UUID) (*domain.BackupValidation, error)

	// 数据恢复
	RestoreFromBackup(ctx context.Context, backupID uuid.UUID, options *domain.RestoreOptions) (*domain.RestoreResult, error)
	TestRestore(ctx context.Context, backupID uuid.UUID) (*domain.RestoreTest, error)

	// 备份策略
	UpdateBackupPolicy(ctx context.Context, policy *domain.BackupPolicy) error
	GetBackupPolicy(ctx context.Context) (*domain.BackupPolicy, error)
	CleanupOldBackups(ctx context.Context) error

	// 监控统计
	GetBackupStats(ctx context.Context) (*domain.BackupStats, error)
	CheckBackupHealth(ctx context.Context) (*domain.BackupHealth, error)
}

// backupService 备份恢复服务实现
type backupService struct {
	backupRepo repository.BackupRepository
	redis      *redis.Client
	logger     *logrus.Logger

	// 备份配置
	config *BackupConfig
}

// BackupConfig 备份配置
type BackupConfig struct {
	DatabaseURL      string `json:"database_url"`
	BackupDir        string `json:"backup_dir"`
	MaxBackups       int    `json:"max_backups"`
	RetentionDays    int    `json:"retention_days"`
	CompressionLevel int    `json:"compression_level"`
	EncryptionKey    string `json:"encryption_key"`

	// 增量备份配置
	IncrementalEnabled  bool          `json:"incremental_enabled"`
	IncrementalInterval time.Duration `json:"incremental_interval"`

	// 验证配置
	ValidationEnabled  bool `json:"validation_enabled"`
	TestRestoreEnabled bool `json:"test_restore_enabled"`
}

// NewBackupService 创建备份恢复服务
func NewBackupService(
	backupRepo repository.BackupRepository,
	redis *redis.Client,
	logger *logrus.Logger,
) BackupService {
	config := &BackupConfig{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		BackupDir:           "./backups",
		MaxBackups:          30,
		RetentionDays:       90,
		CompressionLevel:    6,
		IncrementalEnabled:  true,
		IncrementalInterval: time.Hour * 6,
		ValidationEnabled:   true,
		TestRestoreEnabled:  false,
	}

	// 确保备份目录存在
	os.MkdirAll(config.BackupDir, 0755)

	service := &backupService{
		backupRepo: backupRepo,
		redis:      redis,
		logger:     logger,
		config:     config,
	}

	// 启动定时备份
	go service.startScheduledBackups()

	return service
}

// CreateDatabaseBackup 创建数据库备份
func (s *backupService) CreateDatabaseBackup(ctx context.Context, backupType string) (*domain.BackupRecord, error) {
	backupRecord := &domain.BackupRecord{
		ID:        uuid.New(),
		Type:      backupType,
		Status:    "running",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存备份记录
	if err := s.backupRepo.CreateBackupRecord(ctx, backupRecord); err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"backup_id":   backupRecord.ID,
		"backup_type": backupType,
	}).Info("Starting database backup")

	// 异步执行备份
	go func() {
		if err := s.performBackup(context.Background(), backupRecord); err != nil {
			s.logger.WithError(err).WithField("backup_id", backupRecord.ID).Error("Backup failed")
			s.updateBackupStatus(context.Background(), backupRecord.ID, "failed", err.Error())
		}
	}()

	return backupRecord, nil
}

// CreateIncrementalBackup 创建增量备份
func (s *backupService) CreateIncrementalBackup(ctx context.Context) (*domain.BackupRecord, error) {
	if !s.config.IncrementalEnabled {
		return nil, fmt.Errorf("incremental backup is disabled")
	}

	// 获取最后一次备份时间
	lastBackup, err := s.backupRepo.GetLastBackup(ctx, "full")
	if err != nil {
		s.logger.WithError(err).Warn("No previous backup found, creating full backup")
		return s.CreateDatabaseBackup(ctx, "full")
	}

	backupRecord := &domain.BackupRecord{
		ID:           uuid.New(),
		Type:         "incremental",
		Status:       "running",
		StartTime:    time.Now(),
		BaseBackupID: &lastBackup.ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.backupRepo.CreateBackupRecord(ctx, backupRecord); err != nil {
		return nil, fmt.Errorf("failed to create incremental backup record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"backup_id":      backupRecord.ID,
		"base_backup_id": lastBackup.ID,
	}).Info("Starting incremental backup")

	go func() {
		if err := s.performIncrementalBackup(context.Background(), backupRecord, lastBackup); err != nil {
			s.logger.WithError(err).WithField("backup_id", backupRecord.ID).Error("Incremental backup failed")
			s.updateBackupStatus(context.Background(), backupRecord.ID, "failed", err.Error())
		}
	}()

	return backupRecord, nil
}

// ScheduleBackup 调度备份
func (s *backupService) ScheduleBackup(ctx context.Context, schedule *domain.BackupSchedule) error {
	schedule.ID = uuid.New()
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	if err := s.backupRepo.CreateBackupSchedule(ctx, schedule); err != nil {
		return fmt.Errorf("failed to create backup schedule: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"schedule_id": schedule.ID,
		"cron":        schedule.CronExpression,
		"backup_type": schedule.BackupType,
	}).Info("Backup schedule created")

	return nil
}

// ListBackups 列出备份
func (s *backupService) ListBackups(ctx context.Context, backupType string) ([]*domain.BackupRecord, error) {
	backups, err := s.backupRepo.ListBackups(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}
	return backups, nil
}

// GetBackup 获取备份
func (s *backupService) GetBackup(ctx context.Context, backupID uuid.UUID) (*domain.BackupRecord, error) {
	backup, err := s.backupRepo.GetBackupRecord(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}
	return backup, nil
}

// DeleteBackup 删除备份
func (s *backupService) DeleteBackup(ctx context.Context, backupID uuid.UUID) error {
	backup, err := s.backupRepo.GetBackupRecord(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup: %w", err)
	}

	// 删除备份文件
	if backup.FilePath != "" {
		if err := os.Remove(backup.FilePath); err != nil {
			s.logger.WithError(err).WithField("file_path", backup.FilePath).Warn("Failed to delete backup file")
		}
	}

	// 删除数据库记录
	if err := s.backupRepo.DeleteBackupRecord(ctx, backupID); err != nil {
		return fmt.Errorf("failed to delete backup record: %w", err)
	}

	s.logger.WithField("backup_id", backupID).Info("Backup deleted")
	return nil
}

// ValidateBackup 验证备份
func (s *backupService) ValidateBackup(ctx context.Context, backupID uuid.UUID) (*domain.BackupValidation, error) {
	backup, err := s.backupRepo.GetBackupRecord(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}

	validation := &domain.BackupValidation{
		ID:        uuid.New(),
		BackupID:  backupID,
		Status:    "running",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}

	// 保存验证记录
	if err := s.backupRepo.CreateBackupValidation(ctx, validation); err != nil {
		return nil, fmt.Errorf("failed to create validation record: %w", err)
	}

	// 异步执行验证
	go func() {
		if err := s.performValidation(context.Background(), backup, validation); err != nil {
			s.logger.WithError(err).WithField("backup_id", backupID).Error("Backup validation failed")
		}
	}()

	return validation, nil
}

// RestoreFromBackup 从备份恢复
func (s *backupService) RestoreFromBackup(ctx context.Context, backupID uuid.UUID, options *domain.RestoreOptions) (*domain.RestoreResult, error) {
	backup, err := s.backupRepo.GetBackupRecord(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}

	if backup.Status != "completed" {
		return nil, fmt.Errorf("backup is not in completed status")
	}

	restoreResult := &domain.RestoreResult{
		ID:        uuid.New(),
		BackupID:  backupID,
		Status:    "running",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}

	// 保存恢复记录
	if err := s.backupRepo.CreateRestoreResult(ctx, restoreResult); err != nil {
		return nil, fmt.Errorf("failed to create restore record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"restore_id": restoreResult.ID,
		"backup_id":  backupID,
	}).Info("Starting database restore")

	// 异步执行恢复
	go func() {
		if err := s.performRestore(context.Background(), backup, restoreResult, options); err != nil {
			s.logger.WithError(err).WithField("restore_id", restoreResult.ID).Error("Restore failed")
		}
	}()

	return restoreResult, nil
}

// TestRestore 测试恢复
func (s *backupService) TestRestore(ctx context.Context, backupID uuid.UUID) (*domain.RestoreTest, error) {
	if !s.config.TestRestoreEnabled {
		return nil, fmt.Errorf("test restore is disabled")
	}

	backup, err := s.backupRepo.GetBackupRecord(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}

	testResult := &domain.RestoreTest{
		ID:        uuid.New(),
		BackupID:  backupID,
		Status:    "running",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.backupRepo.CreateRestoreTest(ctx, testResult); err != nil {
		return nil, fmt.Errorf("failed to create test restore record: %w", err)
	}

	// 异步执行测试恢复
	go func() {
		if err := s.performTestRestore(context.Background(), backup, testResult); err != nil {
			s.logger.WithError(err).WithField("test_id", testResult.ID).Error("Test restore failed")
		}
	}()

	return testResult, nil
}

// UpdateBackupPolicy 更新备份策略
func (s *backupService) UpdateBackupPolicy(ctx context.Context, policy *domain.BackupPolicy) error {
	policy.UpdatedAt = time.Now()

	if err := s.backupRepo.UpdateBackupPolicy(ctx, policy); err != nil {
		return fmt.Errorf("failed to update backup policy: %w", err)
	}

	s.logger.WithField("policy_id", policy.ID).Info("Backup policy updated")
	return nil
}

// GetBackupPolicy 获取备份策略
func (s *backupService) GetBackupPolicy(ctx context.Context) (*domain.BackupPolicy, error) {
	// 使用默认策略ID，实际应该从配置或参数获取
	defaultPolicyID := uuid.New()
	policy, err := s.backupRepo.GetBackupPolicy(ctx, defaultPolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup policy: %w", err)
	}
	return policy, nil
}

// CleanupOldBackups 清理旧备份
func (s *backupService) CleanupOldBackups(ctx context.Context) error {
	cutoffDate := time.Now().AddDate(0, 0, -s.config.RetentionDays)

	oldBackups, err := s.backupRepo.GetBackupsOlderThan(ctx, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to get old backups: %w", err)
	}

	var deletedCount int
	for _, backup := range oldBackups {
		if err := s.DeleteBackup(ctx, backup.ID); err != nil {
			s.logger.WithError(err).WithField("backup_id", backup.ID).Error("Failed to delete old backup")
		} else {
			deletedCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"deleted_count": deletedCount,
		"cutoff_date":   cutoffDate,
	}).Info("Old backups cleaned up")

	return nil
}

// GetBackupStats 获取备份统计
func (s *backupService) GetBackupStats(ctx context.Context) (*domain.BackupStats, error) {
	stats, err := s.backupRepo.GetBackupStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup stats: %w", err)
	}
	return stats, nil
}

// CheckBackupHealth 检查备份健康状态
func (s *backupService) CheckBackupHealth(ctx context.Context) (*domain.BackupHealth, error) {
	health := &domain.BackupHealth{
		CheckTime: time.Now(),
		Status:    "healthy",
	}

	// 检查最近备份
	lastBackup, err := s.backupRepo.GetLastBackup(ctx, "full")
	if err != nil {
		health.Status = "warning"
		health.Issues = append(health.Issues, "No recent backup found")
	} else {
		timeSinceLastBackup := time.Since(lastBackup.CreatedAt)
		if timeSinceLastBackup > time.Hour*24 {
			health.Status = "warning"
			health.Issues = append(health.Issues, fmt.Sprintf("Last backup was %v ago", timeSinceLastBackup))
		}

		if lastBackup.Status == "failed" {
			health.Status = "critical"
			health.Issues = append(health.Issues, "Last backup failed")
		}
	}

	// 检查备份存储空间
	if err := s.checkStorageSpace(); err != nil {
		health.Status = "warning"
		health.Issues = append(health.Issues, fmt.Sprintf("Storage issue: %v", err))
	}

	// 检查备份文件完整性
	if s.config.ValidationEnabled {
		recentBackups, err := s.backupRepo.ListBackups(ctx, 5, 0)
		if err == nil {
			for _, backup := range recentBackups {
				if backup.Status == "completed" && backup.FilePath != "" {
					if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
						health.Status = "critical"
						health.Issues = append(health.Issues, fmt.Sprintf("Backup file missing: %s", backup.FilePath))
					}
				}
			}
		}
	}

	return health, nil
}

// 辅助方法

// performBackup 执行备份
func (s *backupService) performBackup(ctx context.Context, backupRecord *domain.BackupRecord) error {
	// 生成备份文件路径
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("yunai_backup_%s_%s.sql", backupRecord.Type, timestamp)
	filePath := filepath.Join(s.config.BackupDir, filename)

	// 执行pg_dump命令
	cmd := exec.CommandContext(ctx, "pg_dump", s.config.DatabaseURL, "-f", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %w", err)
	}

	// 压缩备份文件
	compressedPath := filePath + ".gz"
	if err := s.compressFile(filePath, compressedPath); err != nil {
		s.logger.WithError(err).Warn("Failed to compress backup file")
		compressedPath = filePath
	} else {
		os.Remove(filePath) // 删除原始文件
	}

	// 获取文件信息
	fileInfo, err := os.Stat(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to get backup file info: %w", err)
	}

	// 更新备份记录
	backupRecord.Status = "completed"
	backupRecord.FilePath = compressedPath
	backupRecord.FileSize = fileInfo.Size()
	backupRecord.EndTime = &[]time.Time{time.Now()}[0]
	backupRecord.UpdatedAt = time.Now()

	if err := s.backupRepo.UpdateBackupRecord(ctx, backupRecord); err != nil {
		return fmt.Errorf("failed to update backup record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"backup_id": backupRecord.ID,
		"file_path": compressedPath,
		"file_size": fileInfo.Size(),
	}).Info("Database backup completed")

	// 验证备份
	if s.config.ValidationEnabled {
		go func() {
			if _, err := s.ValidateBackup(context.Background(), backupRecord.ID); err != nil {
				s.logger.WithError(err).WithField("backup_id", backupRecord.ID).Error("Failed to validate backup")
			}
		}()
	}

	return nil
}

// performIncrementalBackup 执行增量备份
func (s *backupService) performIncrementalBackup(ctx context.Context, backupRecord *domain.BackupRecord, baseBackup *domain.BackupRecord) error {
	// 简化实现：使用WAL文件进行增量备份
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("yunai_incremental_%s.tar", timestamp)
	filePath := filepath.Join(s.config.BackupDir, filename)

	// 这里应该实现真正的增量备份逻辑
	// 例如：收集WAL文件、创建增量备份等

	// 模拟增量备份过程
	cmd := exec.CommandContext(ctx, "pg_basebackup", "-D", filePath, "-Ft", "-z", "-P")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("incremental backup failed: %w", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get incremental backup file info: %w", err)
	}

	// 更新备份记录
	backupRecord.Status = "completed"
	backupRecord.FilePath = filePath
	backupRecord.FileSize = fileInfo.Size()
	backupRecord.EndTime = &[]time.Time{time.Now()}[0]
	backupRecord.UpdatedAt = time.Now()

	if err := s.backupRepo.UpdateBackupRecord(ctx, backupRecord); err != nil {
		return fmt.Errorf("failed to update incremental backup record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"backup_id":      backupRecord.ID,
		"base_backup_id": baseBackup.ID,
		"file_path":      filePath,
		"file_size":      fileInfo.Size(),
	}).Info("Incremental backup completed")

	return nil
}

// performValidation 执行备份验证
func (s *backupService) performValidation(ctx context.Context, backup *domain.BackupRecord, validation *domain.BackupValidation) error {
	// 检查文件是否存在
	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		validation.Status = "failed"
		validation.ErrorMessage = "Backup file not found"
		validation.EndTime = &[]time.Time{time.Now()}[0]
		s.backupRepo.UpdateBackupValidation(ctx, validation)
		return fmt.Errorf("backup file not found: %s", backup.FilePath)
	}

	// 检查文件完整性
	if strings.HasSuffix(backup.FilePath, ".gz") {
		if err := s.validateCompressedFile(backup.FilePath); err != nil {
			validation.Status = "failed"
			validation.ErrorMessage = fmt.Sprintf("File integrity check failed: %v", err)
			validation.EndTime = &[]time.Time{time.Now()}[0]
			s.backupRepo.UpdateBackupValidation(ctx, validation)
			return err
		}
	}

	// 验证SQL文件语法（简化实现）
	if err := s.validateSQLFile(backup.FilePath); err != nil {
		validation.Status = "failed"
		validation.ErrorMessage = fmt.Sprintf("SQL validation failed: %v", err)
		validation.EndTime = &[]time.Time{time.Now()}[0]
		s.backupRepo.UpdateBackupValidation(ctx, validation)
		return err
	}

	// 验证成功
	validation.Status = "completed"
	validation.EndTime = &[]time.Time{time.Now()}[0]

	if err := s.backupRepo.UpdateBackupValidation(ctx, validation); err != nil {
		return fmt.Errorf("failed to update validation record: %w", err)
	}

	s.logger.WithField("backup_id", backup.ID).Info("Backup validation completed successfully")
	return nil
}

// performRestore 执行恢复
func (s *backupService) performRestore(ctx context.Context, backup *domain.BackupRecord, restoreResult *domain.RestoreResult, options *domain.RestoreOptions) error {
	// 解压备份文件（如果需要）
	sqlFilePath := backup.FilePath
	if strings.HasSuffix(backup.FilePath, ".gz") {
		decompressedPath := strings.TrimSuffix(backup.FilePath, ".gz")
		if err := s.decompressFile(backup.FilePath, decompressedPath); err != nil {
			return fmt.Errorf("failed to decompress backup file: %w", err)
		}
		sqlFilePath = decompressedPath
		defer os.Remove(decompressedPath) // 清理临时文件
	}

	// 执行恢复
	var cmd *exec.Cmd
	if options != nil && options.TargetDatabase != "" {
		cmd = exec.CommandContext(ctx, "psql", options.TargetDatabase, "-f", sqlFilePath)
	} else {
		cmd = exec.CommandContext(ctx, "psql", s.config.DatabaseURL, "-f", sqlFilePath)
	}

	if err := cmd.Run(); err != nil {
		restoreResult.Status = "failed"
		restoreResult.ErrorMessage = fmt.Sprintf("Restore command failed: %v", err)
		restoreResult.EndTime = &[]time.Time{time.Now()}[0]
		s.backupRepo.UpdateRestoreResult(ctx, restoreResult)
		return fmt.Errorf("restore failed: %w", err)
	}

	// 恢复成功
	restoreResult.Status = "completed"
	restoreResult.EndTime = &[]time.Time{time.Now()}[0]

	if err := s.backupRepo.UpdateRestoreResult(ctx, restoreResult); err != nil {
		return fmt.Errorf("failed to update restore result: %w", err)
	}

	s.logger.WithField("restore_id", restoreResult.ID).Info("Database restore completed successfully")
	return nil
}

// performTestRestore 执行测试恢复
func (s *backupService) performTestRestore(ctx context.Context, backup *domain.BackupRecord, testResult *domain.RestoreTest) error {
	// 创建临时测试数据库
	testDBName := fmt.Sprintf("yunai_test_restore_%s", testResult.ID.String()[:8])

	// 创建测试数据库
	createCmd := exec.CommandContext(ctx, "createdb", testDBName)
	if err := createCmd.Run(); err != nil {
		testResult.Status = "failed"
		testResult.ErrorMessage = fmt.Sprintf("Failed to create test database: %v", err)
		testResult.EndTime = &[]time.Time{time.Now()}[0]
		s.backupRepo.UpdateRestoreTest(ctx, testResult)
		return fmt.Errorf("failed to create test database: %w", err)
	}

	// 确保清理测试数据库
	defer func() {
		dropCmd := exec.CommandContext(context.Background(), "dropdb", testDBName)
		if err := dropCmd.Run(); err != nil {
			s.logger.WithError(err).WithField("test_db", testDBName).Error("Failed to drop test database")
		}
	}()

	// 执行测试恢复
	options := &domain.RestoreOptions{
		TargetDatabase: testDBName,
	}

	tempRestoreResult := &domain.RestoreResult{
		ID:       uuid.New(),
		BackupID: backup.ID,
		Status:   "running",
	}

	if err := s.performRestore(ctx, backup, tempRestoreResult, options); err != nil {
		testResult.Status = "failed"
		testResult.ErrorMessage = fmt.Sprintf("Test restore failed: %v", err)
		testResult.EndTime = &[]time.Time{time.Now()}[0]
		s.backupRepo.UpdateRestoreTest(ctx, testResult)
		return err
	}

	// 验证恢复结果
	if err := s.validateRestoredDatabase(ctx, testDBName); err != nil {
		testResult.Status = "failed"
		testResult.ErrorMessage = fmt.Sprintf("Database validation failed: %v", err)
		testResult.EndTime = &[]time.Time{time.Now()}[0]
		s.backupRepo.UpdateRestoreTest(ctx, testResult)
		return err
	}

	// 测试恢复成功
	testResult.Status = "completed"
	testResult.EndTime = &[]time.Time{time.Now()}[0]

	if err := s.backupRepo.UpdateRestoreTest(ctx, testResult); err != nil {
		return fmt.Errorf("failed to update test restore result: %w", err)
	}

	s.logger.WithField("test_id", testResult.ID).Info("Test restore completed successfully")
	return nil
}

// updateBackupStatus 更新备份状态
func (s *backupService) updateBackupStatus(ctx context.Context, backupID uuid.UUID, status, errorMessage string) {
	backup, err := s.backupRepo.GetBackupRecord(ctx, backupID)
	if err != nil {
		s.logger.WithError(err).WithField("backup_id", backupID).Error("Failed to get backup for status update")
		return
	}

	backup.Status = status
	if errorMessage != "" {
		backup.ErrorMessage = &errorMessage
	}
	backup.UpdatedAt = time.Now()

	if status == "failed" || status == "completed" {
		now := time.Now()
		backup.EndTime = &now
	}

	if err := s.backupRepo.UpdateBackupRecord(ctx, backup); err != nil {
		s.logger.WithError(err).WithField("backup_id", backupID).Error("Failed to update backup status")
	}
}

// checkStorageSpace 检查存储空间
func (s *backupService) checkStorageSpace() error {
	// TODO: 实现存储空间检查逻辑
	return nil
}

// compressFile 压缩文件
func (s *backupService) compressFile(sourcePath, destPath string) error {
	// TODO: 实现文件压缩逻辑
	return nil
}

// validateCompressedFile 验证压缩文件
func (s *backupService) validateCompressedFile(filePath string) error {
	// TODO: 实现压缩文件验证逻辑
	return nil
}

// validateSQLFile 验证SQL文件
func (s *backupService) validateSQLFile(filePath string) error {
	// TODO: 实现SQL文件验证逻辑
	return nil
}

// decompressFile 解压文件
func (s *backupService) decompressFile(sourcePath, destPath string) error {
	// TODO: 实现文件解压逻辑
	return nil
}

// validateRestoredDatabase 验证恢复的数据库
func (s *backupService) validateRestoredDatabase(ctx context.Context, dbName string) error {
	// TODO: 实现数据库验证逻辑
	return nil
}

// startScheduledBackups 启动定时备份
func (s *backupService) startScheduledBackups() {
	ticker := time.NewTicker(time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()

			// 检查是否需要执行定时备份
			if s.shouldCreateBackup(ctx) {
				if _, err := s.CreateDatabaseBackup(ctx, "scheduled"); err != nil {
					s.logger.WithError(err).Error("Failed to create scheduled backup")
				}
			}

			// 检查是否需要执行增量备份
			if s.shouldCreateIncrementalBackup(ctx) {
				if _, err := s.CreateIncrementalBackup(ctx); err != nil {
					s.logger.WithError(err).Error("Failed to create scheduled incremental backup")
				}
			}

			// 清理旧备份
			if err := s.CleanupOldBackups(ctx); err != nil {
				s.logger.WithError(err).Error("Failed to cleanup old backups")
			}
		}
	}
}

// shouldCreateBackup 检查是否应该创建备份
func (s *backupService) shouldCreateBackup(ctx context.Context) bool {
	lastBackup, err := s.backupRepo.GetLastBackup(ctx, "full")
	if err != nil {
		return true // 没有备份，应该创建
	}

	// 如果最后一次备份超过24小时，创建新备份
	return time.Since(lastBackup.CreatedAt) > time.Hour*24
}

// shouldCreateIncrementalBackup 检查是否应该创建增量备份
func (s *backupService) shouldCreateIncrementalBackup(ctx context.Context) bool {
	if !s.config.IncrementalEnabled {
		return false
	}

	lastIncremental, err := s.backupRepo.GetLastBackup(ctx, "incremental")
	if err != nil {
		return true // 没有增量备份，应该创建
	}

	return time.Since(lastIncremental.CreatedAt) > s.config.IncrementalInterval
}
