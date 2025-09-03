package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"yunai/internal/adapter"
	"yunai/internal/domain"
	"yunai/internal/repository"

	"github.com/sirupsen/logrus"
)

// VoiceServiceManager 语音服务管理器
type VoiceServiceManager struct {
	logger          *logrus.Logger
	providerRepo    repository.VoiceProviderRepository
	modelRepo       repository.VoiceModelRepository
	templateRepo    repository.VoiceTemplateRepository
	customVoiceRepo repository.CustomVoiceRepository
	adapterManager  *adapter.VoiceAdapterManager
	adapterFactory  adapter.VoiceAdapterFactory
}

// NewVoiceServiceManager 创建语音服务管理器
func NewVoiceServiceManager(
	logger *logrus.Logger,
	providerRepo repository.VoiceProviderRepository,
	modelRepo repository.VoiceModelRepository,
	templateRepo repository.VoiceTemplateRepository,
	customVoiceRepo repository.CustomVoiceRepository,
) *VoiceServiceManager {
	factory := adapter.NewDefaultVoiceAdapterFactory()
	adapterManager := adapter.NewVoiceAdapterManager(factory)

	return &VoiceServiceManager{
		logger:          logger,
		providerRepo:    providerRepo,
		modelRepo:       modelRepo,
		templateRepo:    templateRepo,
		customVoiceRepo: customVoiceRepo,
		adapterManager:  adapterManager,
		adapterFactory:  factory,
	}
}

// InitializeProviders 初始化所有提供商
func (m *VoiceServiceManager) InitializeProviders(ctx context.Context) error {
	m.logger.Info("开始初始化语音服务提供商")

	providers, err := m.providerRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("获取活跃提供商失败: %w", err)
	}

	for _, provider := range providers {
		err := m.initializeProvider(ctx, provider)
		if err != nil {
			m.logger.WithError(err).WithField("provider", provider.Name).Warn("初始化提供商失败")
			continue
		}
		m.logger.WithField("provider", provider.Name).Info("提供商初始化成功")
	}

	m.logger.WithField("count", len(providers)).Info("语音服务提供商初始化完成")
	return nil
}

// initializeProvider 初始化单个提供商
func (m *VoiceServiceManager) initializeProvider(ctx context.Context, provider *domain.VoiceProvider) error {
	// 创建适配器
	voiceAdapter, err := m.adapterFactory.CreateAdapter(provider)
	if err != nil {
		return fmt.Errorf("创建适配器失败: %w", err)
	}

	// 验证配置
	var config map[string]interface{}
	if len(provider.Config) > 0 {
		// 解析配置
		// config = parseConfig(provider.Config)
	}

	err = voiceAdapter.ValidateConfig(config)
	if err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 健康检查
	err = voiceAdapter.HealthCheck(ctx)
	if err != nil {
		m.logger.WithError(err).WithField("provider", provider.Name).Warn("健康检查失败")
		// 不返回错误，允许提供商在不健康状态下注册
	}

	// 注册适配器
	m.adapterManager.RegisterAdapter(provider.ID, voiceAdapter)

	return nil
}

// TextToSpeech 文本转语音 (智能路由)
func (m *VoiceServiceManager) TextToSpeech(ctx context.Context, req *domain.TTSRequest) ([]byte, error) {
	m.logger.WithFields(logrus.Fields{
		"provider_id": req.ProviderID,
		"model_key":   req.ModelKey,
		"text_length": len(req.Text),
	}).Info("开始文本转语音")

	// 1. 选择提供商和模型
	provider, model, err := m.selectTTSProvider(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("选择TTS提供商失败: %w", err)
	}

	// 2. 获取适配器
	voiceAdapter, exists := m.adapterManager.GetAdapter(provider.ID)
	if !exists {
		return nil, fmt.Errorf("提供商 %s 的适配器不存在", provider.Name)
	}

	// 3. 设置请求参数
	req.ProviderID = provider.ID
	req.ModelKey = model.ModelKey

	// 4. 调用TTS服务
	startTime := time.Now()
	audioData, err := voiceAdapter.TextToSpeech(ctx, req)
	duration := time.Since(startTime)

	// 5. 记录指标
	m.recordMetrics(provider.ID, "tts", err == nil, duration)

	if err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"provider": provider.Name,
			"model":    model.ModelKey,
		}).Error("TTS调用失败")
		return nil, fmt.Errorf("TTS调用失败: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"provider":    provider.Name,
		"model":       model.ModelKey,
		"duration_ms": duration.Milliseconds(),
		"audio_size":  len(audioData),
	}).Info("TTS调用成功")

	return audioData, nil
}

// SpeechToText 语音转文本 (智能路由)
func (m *VoiceServiceManager) SpeechToText(ctx context.Context, req *domain.ASRRequest) (*domain.ASRResponse, error) {
	m.logger.WithFields(logrus.Fields{
		"provider_id":  req.ProviderID,
		"model_key":    req.ModelKey,
		"audio_length": len(req.AudioData),
	}).Info("开始语音转文本")

	// 1. 选择提供商和模型
	provider, model, err := m.selectASRProvider(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("选择ASR提供商失败: %w", err)
	}

	// 2. 获取适配器
	voiceAdapter, exists := m.adapterManager.GetAdapter(provider.ID)
	if !exists {
		return nil, fmt.Errorf("提供商 %s 的适配器不存在", provider.Name)
	}

	// 3. 设置请求参数
	req.ProviderID = provider.ID
	req.ModelKey = model.ModelKey

	// 4. 调用ASR服务
	startTime := time.Now()
	result, err := voiceAdapter.SpeechToText(ctx, req)
	duration := time.Since(startTime)

	// 5. 记录指标
	m.recordMetrics(provider.ID, "asr", err == nil, duration)

	if err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"provider": provider.Name,
			"model":    model.ModelKey,
		}).Error("ASR调用失败")
		return nil, fmt.Errorf("ASR调用失败: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"provider":    provider.Name,
		"model":       model.ModelKey,
		"duration_ms": duration.Milliseconds(),
		"text_length": len(result.Text),
	}).Info("ASR调用成功")

	return result, nil
}

// CloneVoice 音色克隆
func (m *VoiceServiceManager) CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	m.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"provider_id":  req.ProviderID,
		"voice_name":   req.VoiceName,
		"audio_length": len(req.AudioData),
	}).Info("开始音色克隆")

	// 1. 获取提供商
	provider, err := m.providerRepo.GetByID(ctx, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("获取提供商失败: %w", err)
	}

	// 2. 检查提供商是否支持音色克隆
	voiceAdapter, exists := m.adapterManager.GetAdapter(provider.ID)
	if !exists {
		return nil, fmt.Errorf("提供商 %s 的适配器不存在", provider.Name)
	}

	capabilities := voiceAdapter.GetSupportedCapabilities()
	if !capabilities.VoiceClone {
		return nil, fmt.Errorf("提供商 %s 不支持音色克隆", provider.Name)
	}

	// 3. 调用音色克隆服务
	startTime := time.Now()
	result, err := voiceAdapter.CloneVoice(ctx, req)
	duration := time.Since(startTime)

	// 4. 记录指标
	m.recordMetrics(provider.ID, "voice_clone", err == nil, duration)

	if err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"provider":   provider.Name,
			"voice_name": req.VoiceName,
		}).Error("音色克隆失败")
		return nil, fmt.Errorf("音色克隆失败: %w", err)
	}

	// 5. 保存自定义音色信息
	customVoice := &domain.CustomVoice{
		UserID:        req.UserID,
		ProviderID:    provider.ID,
		VoiceName:     req.VoiceName,
		VoiceID:       result.VoiceID,
		CoverImage:    req.CoverImage,
		ReferenceText: req.ReferenceText,
		Status:        "ready",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = m.customVoiceRepo.Create(ctx, customVoice)
	if err != nil {
		m.logger.WithError(err).Warn("保存自定义音色信息失败")
		// 不返回错误，因为克隆已经成功
	}

	m.logger.WithFields(logrus.Fields{
		"provider":    provider.Name,
		"voice_name":  req.VoiceName,
		"voice_id":    result.VoiceID,
		"duration_ms": duration.Milliseconds(),
	}).Info("音色克隆成功")

	return result, nil
}

// selectTTSProvider 选择TTS提供商和模型
func (m *VoiceServiceManager) selectTTSProvider(ctx context.Context, req *domain.TTSRequest) (*domain.VoiceProvider, *domain.VoiceModel, error) {
	// 如果指定了提供商，直接使用
	if req.ProviderID != "" {
		provider, err := m.providerRepo.GetByID(ctx, req.ProviderID)
		if err != nil {
			return nil, nil, fmt.Errorf("获取指定提供商失败: %w", err)
		}

		model, err := m.selectModelForProvider(ctx, provider.ID, "tts", req.ModelKey)
		if err != nil {
			return nil, nil, fmt.Errorf("选择模型失败: %w", err)
		}

		return provider, model, nil
	}

	// 智能选择最佳提供商
	return m.selectBestTTSProvider(ctx, req)
}

// selectASRProvider 选择ASR提供商和模型
func (m *VoiceServiceManager) selectASRProvider(ctx context.Context, req *domain.ASRRequest) (*domain.VoiceProvider, *domain.VoiceModel, error) {
	// 如果指定了提供商，直接使用
	if req.ProviderID != "" {
		provider, err := m.providerRepo.GetByID(ctx, req.ProviderID)
		if err != nil {
			return nil, nil, fmt.Errorf("获取指定提供商失败: %w", err)
		}

		model, err := m.selectModelForProvider(ctx, provider.ID, "asr", req.ModelKey)
		if err != nil {
			return nil, nil, fmt.Errorf("选择模型失败: %w", err)
		}

		return provider, model, nil
	}

	// 智能选择最佳提供商
	return m.selectBestASRProvider(ctx, req)
}

// selectBestTTSProvider 选择最佳TTS提供商
func (m *VoiceServiceManager) selectBestTTSProvider(ctx context.Context, req *domain.TTSRequest) (*domain.VoiceProvider, *domain.VoiceModel, error) {
	// 获取所有支持TTS的提供商
	providers, err := m.providerRepo.ListByType(ctx, "tts")
	if err != nil {
		return nil, nil, fmt.Errorf("获取TTS提供商失败: %w", err)
	}

	if len(providers) == 0 {
		return nil, nil, fmt.Errorf("没有可用的TTS提供商")
	}

	// 按优先级排序
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority > providers[j].Priority
	})

	// 选择第一个可用的提供商
	for _, provider := range providers {
		model, err := m.selectModelForProvider(ctx, provider.ID, "tts", "")
		if err != nil {
			continue
		}

		return provider, model, nil
	}

	return nil, nil, fmt.Errorf("没有可用的TTS模型")
}

// selectBestASRProvider 选择最佳ASR提供商
func (m *VoiceServiceManager) selectBestASRProvider(ctx context.Context, req *domain.ASRRequest) (*domain.VoiceProvider, *domain.VoiceModel, error) {
	// 获取所有支持ASR的提供商
	providers, err := m.providerRepo.ListByType(ctx, "asr")
	if err != nil {
		return nil, nil, fmt.Errorf("获取ASR提供商失败: %w", err)
	}

	if len(providers) == 0 {
		return nil, nil, fmt.Errorf("没有可用的ASR提供商")
	}

	// 按优先级排序
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority > providers[j].Priority
	})

	// 选择第一个可用的提供商
	for _, provider := range providers {
		model, err := m.selectModelForProvider(ctx, provider.ID, "asr", "")
		if err != nil {
			continue
		}

		return provider, model, nil
	}

	return nil, nil, fmt.Errorf("没有可用的ASR模型")
}

// selectModelForProvider 为提供商选择模型
func (m *VoiceServiceManager) selectModelForProvider(ctx context.Context, providerID, modelType, modelKey string) (*domain.VoiceModel, error) {
	// 如果指定了模型，直接使用
	if modelKey != "" {
		model, err := m.modelRepo.GetByProviderAndKey(ctx, providerID, modelKey)
		if err != nil {
			return nil, fmt.Errorf("获取指定模型失败: %w", err)
		}
		return model, nil
	}

	// 选择权重最高的模型
	models, err := m.modelRepo.ListByProviderAndType(ctx, providerID, modelType)
	if err != nil {
		return nil, fmt.Errorf("获取模型列表失败: %w", err)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("提供商 %s 没有可用的 %s 模型", providerID, modelType)
	}

	// 按权重排序
	sort.Slice(models, func(i, j int) bool {
		return models[i].Weight > models[j].Weight
	})

	return models[0], nil
}

// recordMetrics 记录指标
func (m *VoiceServiceManager) recordMetrics(providerID, operation string, success bool, duration time.Duration) {
	// 这里可以集成监控系统，如Prometheus
	m.logger.WithFields(logrus.Fields{
		"provider_id": providerID,
		"operation":   operation,
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	}).Debug("记录语音服务指标")
}
