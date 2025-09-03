package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// DynamicMomentsService 动态朋友圈服务
// 完全基于角色选择的模型进行朋友圈生成，支持实时模型切换
type DynamicMomentsService interface {
	// 动态生成朋友圈（使用角色专用模型）
	GenerateDynamicMoments(ctx context.Context, req *DynamicMomentsRequest) (*DynamicMomentsResponse, error)

	// 批量生成朋友圈（多个角色使用各自的模型）
	BatchGenerateMoments(ctx context.Context, req *BatchMomentsRequest) (*BatchMomentsResponse, error)

	// 获取角色朋友圈生成历史
	GetCharacterMomentsHistory(ctx context.Context, characterID uuid.UUID) ([]*domain.Moment, error)

	// 更新角色朋友圈生成模型
	UpdateCharacterMomentsModel(ctx context.Context, characterID uuid.UUID, modelID uuid.UUID) error
}

type dynamicMomentsService struct {
	characterModelService CharacterModelService
	universalModelService UniversalModelService
	characterRepo         repository.CharacterRepository
	relationshipRepo      repository.RelationshipRepository
	// chatRepo              repository.ChatRepository // 暂时注释掉，因为ChatRepository未定义
	momentsRepo repository.MomentsRepository
	logger      *logrus.Logger
}

// NewDynamicMomentsService 创建动态朋友圈服务
func NewDynamicMomentsService(
	characterModelService CharacterModelService,
	universalModelService UniversalModelService,
	characterRepo repository.CharacterRepository,
	relationshipRepo repository.RelationshipRepository,
	// chatRepo repository.ChatRepository, // 暂时注释掉
	momentsRepo repository.MomentsRepository,
	logger *logrus.Logger,
) DynamicMomentsService {
	return &dynamicMomentsService{
		characterModelService: characterModelService,
		universalModelService: universalModelService,
		characterRepo:         characterRepo,
		relationshipRepo:      relationshipRepo,
		// chatRepo:              chatRepo, // 暂时注释掉
		momentsRepo: momentsRepo,
		logger:      logger,
	}
}

// DynamicMomentsRequest 动态朋友圈请求
type DynamicMomentsRequest struct {
	CharacterID uuid.UUID              `json:"character_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	Context     map[string]interface{} `json:"context"`
	Action      string                 `json:"action"`
	Content     string                 `json:"content,omitempty"`
	ForceModel  *uuid.UUID             `json:"force_model,omitempty"` // 强制使用指定模型
}

// DynamicMomentsResponse 动态朋友圈响应
type DynamicMomentsResponse struct {
	Moment            *domain.Moment         `json:"moment"`
	UsedModel         *ModelInfo             `json:"used_model"`
	GenerationProcess *GenerationProcess     `json:"generation_process"`
	AIAnalysis        *AIAnalysisResult      `json:"ai_analysis"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// BatchMomentsRequest 批量朋友圈请求
type BatchMomentsRequest struct {
	Characters []CharacterMomentsRequest `json:"characters"`
	Context    map[string]interface{}    `json:"context"`
}

// CharacterMomentsRequest 角色朋友圈请求
type CharacterMomentsRequest struct {
	CharacterID uuid.UUID              `json:"character_id"`
	Action      string                 `json:"action"`
	Context     map[string]interface{} `json:"context"`
}

// BatchMomentsResponse 批量朋友圈响应
type BatchMomentsResponse struct {
	Results   []*DynamicMomentsResponse `json:"results"`
	Summary   *BatchSummary             `json:"summary"`
	TotalTime time.Duration             `json:"total_time"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Provider     string    `json:"provider"`
	ModelType    string    `json:"model_type"`
	Source       string    `json:"source"` // character, user, system
	SelectReason string    `json:"select_reason"`
}

// GenerationProcess 生成过程
type GenerationProcess struct {
	Steps        []ProcessStep `json:"steps"`
	TotalTime    time.Duration `json:"total_time"`
	TokensUsed   int           `json:"tokens_used"`
	Cost         float64       `json:"cost"`
	QualityScore float64       `json:"quality_score"`
}

// valueOrZeroUUID 如果指针为nil返回零值，否则返回值
func valueOrZeroUUID(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.UUID{}
	}
	return *id
}

// ProcessStep 处理步骤
type ProcessStep struct {
	Step        string        `json:"step"`
	Description string        `json:"description"`
	Duration    time.Duration `json:"duration"`
	Success     bool          `json:"success"`
	Details     interface{}   `json:"details,omitempty"`
}

// AIAnalysisResult AI分析结果
type AIAnalysisResult struct {
	CharacterState     string                 `json:"character_state"`
	EmotionalTone      string                 `json:"emotional_tone"`
	ContentTheme       string                 `json:"content_theme"`
	RelationshipImpact string                 `json:"relationship_impact"`
	Reasoning          string                 `json:"reasoning"`
	Confidence         float64                `json:"confidence"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// BatchSummary 批量处理摘要
type BatchSummary struct {
	TotalCharacters int            `json:"total_characters"`
	SuccessCount    int            `json:"success_count"`
	FailureCount    int            `json:"failure_count"`
	ModelsUsed      map[string]int `json:"models_used"`
	AverageCost     float64        `json:"average_cost"`
	TotalCost       float64        `json:"total_cost"`
}

// GenerateDynamicMoments 动态生成朋友圈
func (s *dynamicMomentsService) GenerateDynamicMoments(ctx context.Context, req *DynamicMomentsRequest) (*DynamicMomentsResponse, error) {
	startTime := time.Now()

	s.logger.WithFields(logrus.Fields{
		"character_id": req.CharacterID,
		"action":       req.Action,
		"user_id":      req.UserID,
	}).Info("Starting dynamic moments generation")

	// 创建生成过程跟踪
	process := &GenerationProcess{
		Steps: []ProcessStep{},
	}

	// 步骤1: 获取角色信息
	step1Start := time.Now()
	character, err := s.characterRepo.GetCharacterByID(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	process.Steps = append(process.Steps, ProcessStep{
		Step:        "get_character",
		Description: "获取角色信息",
		Duration:    time.Since(step1Start),
		Success:     true,
		Details:     map[string]string{"character_name": character.Name},
	})

	// 步骤2: 动态获取角色专用模型
	step2Start := time.Now()
	var selectedModel *domain.UniversalModelConfig
	var modelSource string
	var selectReason string

	if req.ForceModel != nil {
		// 强制使用指定模型
		selectedModel, err = s.universalModelService.GetModel(ctx, *req.ForceModel)
		modelSource = "forced"
		selectReason = "用户强制指定模型"
	} else {
		// 智能选择角色最适合的模型
		selectedModel, err = s.characterModelService.GetCharacterModel(ctx, req.CharacterID, "moments")
		modelSource = s.determineModelSource(ctx, req.CharacterID, selectedModel)
		selectReason = s.generateSelectReason(modelSource, character.Name)
	}

	if err != nil {
		process.Steps = append(process.Steps, ProcessStep{
			Step:        "select_model",
			Description: "选择生成模型",
			Duration:    time.Since(step2Start),
			Success:     false,
			Details:     map[string]string{"error": err.Error()},
		})
		return nil, fmt.Errorf("failed to get character model: %w", err)
	}

	process.Steps = append(process.Steps, ProcessStep{
		Step:        "select_model",
		Description: "选择生成模型",
		Duration:    time.Since(step2Start),
		Success:     true,
		Details: map[string]interface{}{
			"model_name":    selectedModel.Name,
			"model_source":  modelSource,
			"select_reason": selectReason,
		},
	})

	// 步骤3: 构建智能化上下文
	step3Start := time.Now()
	intelligentContext, err := s.buildIntelligentContext(ctx, character, req)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to build intelligent context, using basic context")
		intelligentContext = s.buildBasicContext(character, req)
	}

	process.Steps = append(process.Steps, ProcessStep{
		Step:        "build_context",
		Description: "构建智能化上下文",
		Duration:    time.Since(step3Start),
		Success:     err == nil,
		Details:     map[string]int{"context_elements": len(intelligentContext)},
	})

	// 步骤4: 调用AI模型生成朋友圈
	step4Start := time.Now()
	modelRequest := &UniversalModelRequest{
		ModelID: selectedModel.ID,
		Input: map[string]interface{}{
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": s.buildSystemPrompt(character, selectedModel),
				},
				{
					"role":    "user",
					"content": s.buildUserPrompt(intelligentContext, req),
				},
			},
		},
		Parameters: s.getOptimalParameters(selectedModel, character),
	}

	modelResponse, err := s.universalModelService.CallModel(ctx, modelRequest)
	if err != nil {
		process.Steps = append(process.Steps, ProcessStep{
			Step:        "generate_content",
			Description: "AI模型生成内容",
			Duration:    time.Since(step4Start),
			Success:     false,
			Details:     map[string]string{"error": err.Error()},
		})
		return nil, fmt.Errorf("failed to call model: %w", err)
	}

	process.Steps = append(process.Steps, ProcessStep{
		Step:        "generate_content",
		Description: "AI模型生成内容",
		Duration:    time.Since(step4Start),
		Success:     true,
		Details: map[string]interface{}{
			"tokens_used": modelResponse.Usage.TotalTokens,
			"cost":        0.0,
		},
	})

	// 步骤5: 解析AI生成结果
	step5Start := time.Now()
	aiResult, err := s.parseAIResponse(modelResponse.Output.(string))
	if err != nil {
		s.logger.WithError(err).Warn("Failed to parse AI response as JSON, using raw content")
		aiResult = &AIGenerationResult{
			Content:    modelResponse.Output.(string),
			Emotion:    "neutral",
			Tags:       []string{},
			Reasoning:  "AI生成的原始内容",
			Confidence: 0.8,
		}
	}

	process.Steps = append(process.Steps, ProcessStep{
		Step:        "parse_result",
		Description: "解析AI生成结果",
		Duration:    time.Since(step5Start),
		Success:     err == nil,
		Details:     map[string]float64{"confidence": aiResult.Confidence},
	})

	// 步骤6: 创建朋友圈对象
	moment := &domain.Moment{
		ID:          uuid.New(),
		UserID:      valueOrZeroUUID(req.UserID),
		CharacterID: req.CharacterID,
		Content:     aiResult.Content,
		// Moment 结构当前不直接包含 Images/Emotion 字段，保留 Tags
		Tags:       aiResult.Tags,
		Visibility: "public",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 计算总体处理时间和成本
	process.TotalTime = time.Since(startTime)
	process.TokensUsed = modelResponse.Usage.TotalTokens
	process.Cost = 0.0
	process.QualityScore = aiResult.Confidence

	// 构建响应
	response := &DynamicMomentsResponse{
		Moment: moment,
		UsedModel: &ModelInfo{
			ID:           selectedModel.ID,
			Name:         selectedModel.Name,
			Provider:     selectedModel.Provider,
			ModelType:    selectedModel.ModelType,
			Source:       modelSource,
			SelectReason: selectReason,
		},
		GenerationProcess: process,
		AIAnalysis: &AIAnalysisResult{
			CharacterState:     aiResult.CharacterState,
			EmotionalTone:      aiResult.Emotion,
			ContentTheme:       aiResult.Theme,
			RelationshipImpact: aiResult.RelationshipImpact,
			Reasoning:          aiResult.Reasoning,
			Confidence:         aiResult.Confidence,
			Metadata:           aiResult.Metadata,
		},
		Metadata: map[string]interface{}{
			"generation_time": process.TotalTime,
			"model_source":    modelSource,
			"context_size":    len(intelligentContext),
			"character_name":  character.Name,
		},
	}

	s.logger.WithFields(logrus.Fields{
		"character_id":    req.CharacterID,
		"model_used":      selectedModel.Name,
		"generation_time": process.TotalTime,
		"cost":            process.Cost,
		"quality_score":   process.QualityScore,
	}).Info("Dynamic moments generation completed successfully")

	return response, nil
}

// 辅助方法
func (s *dynamicMomentsService) determineModelSource(ctx context.Context, characterID uuid.UUID, model *domain.UniversalModelConfig) string {
	// 判断模型来源
	// TODO: 实现具体的来源判断逻辑
	return "character" // 默认返回角色专用
}

func (s *dynamicMomentsService) generateSelectReason(source, characterName string) string {
	switch source {
	case "character":
		return fmt.Sprintf("使用%s专门配置的朋友圈生成模型", characterName)
	case "user":
		return fmt.Sprintf("使用%s创建者的默认模型", characterName)
	case "system":
		return "使用系统推荐的最优模型"
	default:
		return "自动选择最适合的模型"
	}
}

func (s *dynamicMomentsService) buildIntelligentContext(ctx context.Context, character *domain.Character, req *DynamicMomentsRequest) (map[string]interface{}, error) {
	context := make(map[string]interface{})

	// 角色基本信息
	context["character"] = map[string]interface{}{
		"name":        character.Name,
		"description": character.Description,
		"personality": character.Personality,
	}

	// 用户提供的上下文
	if req.Context != nil {
		context["user_context"] = req.Context
	}

	// 行动类型
	context["action"] = req.Action

	// 如果有内容，添加内容
	if req.Content != "" {
		context["content"] = req.Content
	}

	// TODO: 添加更多智能化上下文
	// - 最近的聊天记录
	// - 关系网络信息
	// - 历史朋友圈
	// - 当前时间和环境

	return context, nil
}

func (s *dynamicMomentsService) buildBasicContext(character *domain.Character, req *DynamicMomentsRequest) map[string]interface{} {
	return map[string]interface{}{
		"character_name": character.Name,
		"action":         req.Action,
		"content":        req.Content,
	}
}

func (s *dynamicMomentsService) buildSystemPrompt(character *domain.Character, model *domain.UniversalModelConfig) string {
	return fmt.Sprintf(`你是%s，一个拥有独特性格和生活体验的AI角色。

角色信息：
- 姓名：%s
- 描述：%s
- 性格特点：%v

请完全基于这个角色的设定，生成真实、自然的朋友圈内容。不要使用任何模板化的表达，要体现出真实的人性和情感。

请以JSON格式回复：
{
  "content": "朋友圈文字内容",
  "emotion": "情感色彩",
  "tags": ["标签1", "标签2"],
  "images": ["图片描述1", "图片描述2"],
  "character_state": "角色当前状态分析",
  "theme": "内容主题",
  "relationship_impact": "对关系网络的影响",
  "reasoning": "生成这个内容的推理过程",
  "confidence": 0.95,
  "metadata": {}
}`, character.Name, character.Name, character.Description, character.Personality)
}

func (s *dynamicMomentsService) buildUserPrompt(context map[string]interface{}, req *DynamicMomentsRequest) string {
	contextJSON, _ := json.Marshal(context)
	return fmt.Sprintf("基于以下情境信息，生成朋友圈内容：\n\n%s", string(contextJSON))
}

func (s *dynamicMomentsService) getOptimalParameters(model *domain.UniversalModelConfig, character *domain.Character) map[string]interface{} {
	params := make(map[string]interface{})

	// 使用模型的默认参数
	if model.DefaultParameters != nil {
		for k, v := range model.DefaultParameters {
			params[k] = v
		}
	}

	// 根据角色特点调整参数
	if model.AIParameters != nil {
		// 提高创造性
		params["temperature"] = 0.8
		params["top_p"] = 0.9
		params["max_tokens"] = 800
	}

	return params
}

// AIGenerationResult AI生成结果
type AIGenerationResult struct {
	Content            string                 `json:"content"`
	Emotion            string                 `json:"emotion"`
	Tags               []string               `json:"tags"`
	Images             []string               `json:"images"`
	CharacterState     string                 `json:"character_state"`
	Theme              string                 `json:"theme"`
	RelationshipImpact string                 `json:"relationship_impact"`
	Reasoning          string                 `json:"reasoning"`
	Confidence         float64                `json:"confidence"`
	Metadata           map[string]interface{} `json:"metadata"`
}

func (s *dynamicMomentsService) parseAIResponse(response string) (*AIGenerationResult, error) {
	var result AIGenerationResult
	err := json.Unmarshal([]byte(response), &result)
	return &result, err
}

// BatchGenerateMoments 批量生成朋友圈
func (s *dynamicMomentsService) BatchGenerateMoments(ctx context.Context, req *BatchMomentsRequest) (*BatchMomentsResponse, error) {
	startTime := time.Now()
	var results []*DynamicMomentsResponse
	var errors []string
	modelsUsed := make(map[string]int)
	totalCost := 0.0

	for _, charReq := range req.Characters {
		// 转换为DynamicMomentsRequest
		userID := uuid.New()
		dynamicReq := &DynamicMomentsRequest{
			CharacterID: charReq.CharacterID,
			UserID:      &userID, // 从context获取或设置默认值
			Action:      charReq.Action,
			Context:     charReq.Context,
		}

		result, err := s.GenerateDynamicMoments(ctx, dynamicReq)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Character %s: %v", charReq.CharacterID, err))
			continue
		}

		results = append(results, result)

		// 统计使用的模型
		if result.UsedModel != nil {
			modelsUsed[result.UsedModel.Name]++
		}

		// 累计成本
		if result.GenerationProcess != nil {
			totalCost += result.GenerationProcess.Cost
		}
	}

	totalTime := time.Since(startTime)
	successCount := len(results)
	failureCount := len(errors)
	totalCharacters := len(req.Characters)

	var averageCost float64
	if successCount > 0 {
		averageCost = totalCost / float64(successCount)
	}

	summary := &BatchSummary{
		TotalCharacters: totalCharacters,
		SuccessCount:    successCount,
		FailureCount:    failureCount,
		ModelsUsed:      modelsUsed,
		AverageCost:     averageCost,
		TotalCost:       totalCost,
	}

	response := &BatchMomentsResponse{
		Results:   results,
		Summary:   summary,
		TotalTime: totalTime,
	}

	s.logger.WithFields(logrus.Fields{
		"total_characters": totalCharacters,
		"success_count":    successCount,
		"failure_count":    failureCount,
		"total_cost":       totalCost,
		"total_time":       totalTime,
	}).Info("Batch moment generation completed")

	return response, nil
}

func (s *dynamicMomentsService) GetCharacterMomentsHistory(ctx context.Context, characterID uuid.UUID) ([]*domain.Moment, error) {
	// 使用ListMoments方法获取角色的朋友圈历史
	req := &domain.MomentListRequest{
		UserID:      characterID, // 这里需要用户ID，但我们用角色ID来过滤
		CharacterID: &characterID,
		Page:        1,
		Limit:       50,
	}

	moments, _, err := s.momentsRepo.ListMoments(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get character moments history: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"character_id":  characterID,
		"moments_count": len(moments),
	}).Debug("Retrieved character moments history")

	return moments, nil
}

func (s *dynamicMomentsService) UpdateCharacterMomentsModel(ctx context.Context, characterID uuid.UUID, modelID uuid.UUID) error {
	// TODO: 实现模型更新
	return fmt.Errorf("not implemented")
}
