package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// CharacterImageService 角色两图系统服务
type CharacterImageService interface {
	// 验证两图系统
	ValidateTwoImageSystem(ctx context.Context, req *domain.TwoImageValidationRequest) (*domain.TwoImageValidationResult, error)

	// 获取单聊背景
	GetSingleChatBackground(ctx context.Context, characterID uuid.UUID) (*domain.ChatBackgroundResult, error)

	// 获取群聊抠图显示配置
	GetGroupChatCutoutConfig(ctx context.Context, req *domain.GroupChatCutoutRequest) (*domain.GroupChatCutoutResult, error)

	// 处理角色发言时的抠图切换
	HandleCharacterSpeaking(ctx context.Context, req *domain.CharacterSpeakingRequest) (*domain.CharacterSpeakingResult, error)

	// 自动生成角色两图（可选功能）
	GenerateCharacterImages(ctx context.Context, req *domain.ImageGenerationRequest) (*domain.ImageGenerationResult, error)
}

type characterImageService struct {
	characterRepo repository.CharacterRepository
	mediaService  MediaService // 假设有媒体服务
	logger        *logrus.Logger
}

// NewCharacterImageService 创建角色两图系统服务
func NewCharacterImageService(
	characterRepo repository.CharacterRepository,
	mediaService MediaService,
	logger *logrus.Logger,
) CharacterImageService {
	return &characterImageService{
		characterRepo: characterRepo,
		mediaService:  mediaService,
		logger:        logger,
	}
}

// ValidateTwoImageSystem 验证两图系统
func (s *characterImageService) ValidateTwoImageSystem(ctx context.Context, req *domain.TwoImageValidationRequest) (*domain.TwoImageValidationResult, error) {
	s.logger.WithFields(logrus.Fields{
		"character_id":     req.CharacterID,
		"bg_image_url":     req.BgImageURL,
		"cutout_image_url": req.CutoutImageURL,
	}).Debug("Validating two image system")

	result := &domain.TwoImageValidationResult{
		CharacterID: req.CharacterID,
		IsValid:     true,
		Issues:      []string{},
		Suggestions: []string{},
	}

	// 验证背景图
	if req.BgImageURL == "" {
		result.Issues = append(result.Issues, "背景图不能为空")
		result.IsValid = false
	} else {
		// 验证背景图格式和尺寸
		if !s.isValidImageURL(req.BgImageURL) {
			result.Issues = append(result.Issues, "背景图URL格式不正确")
			result.IsValid = false
		}
		
		// 建议背景图规格
		result.Suggestions = append(result.Suggestions, "建议背景图尺寸：1920x1080，支持JPG/PNG格式")
	}

	// 验证抠图
	if req.CutoutImageURL == "" {
		result.Issues = append(result.Issues, "抠图不能为空")
		result.IsValid = false
	} else {
		// 验证抠图格式（必须支持透明背景）
		if !s.isValidCutoutURL(req.CutoutImageURL) {
			result.Issues = append(result.Issues, "抠图必须是PNG格式以支持透明背景")
			result.IsValid = false
		}
		
		// 建议抠图规格
		result.Suggestions = append(result.Suggestions, "建议抠图尺寸：512x768，PNG格式，透明背景")
	}

	// 验证图片可访问性
	if result.IsValid {
		bgAccessible := s.validateImageAccessibility(ctx, req.BgImageURL)
		cutoutAccessible := s.validateImageAccessibility(ctx, req.CutoutImageURL)
		
		if !bgAccessible {
			result.Issues = append(result.Issues, "背景图无法访问")
			result.IsValid = false
		}
		
		if !cutoutAccessible {
			result.Issues = append(result.Issues, "抠图无法访问")
			result.IsValid = false
		}
	}

	return result, nil
}

// GetSingleChatBackground 获取单聊背景
func (s *characterImageService) GetSingleChatBackground(ctx context.Context, characterID uuid.UUID) (*domain.ChatBackgroundResult, error) {
	s.logger.WithField("character_id", characterID).Debug("Getting single chat background")

	// 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	result := &domain.ChatBackgroundResult{
		CharacterID:   characterID,
		CharacterName: character.Name,
		ChatType:      "single",
	}

	// 单聊自动使用角色的背景图
	if character.BgImageURL != nil && *character.BgImageURL != "" {
		result.BackgroundImageURL = *character.BgImageURL
		result.BackgroundType = "character_bg"
		result.Description = fmt.Sprintf("使用%s的背景图作为单聊背景", character.Name)
	} else {
		// 如果没有背景图，使用默认背景
		result.BackgroundImageURL = s.getDefaultChatBackground()
		result.BackgroundType = "default"
		result.Description = "使用默认聊天背景"
	}

	return result, nil
}

// GetGroupChatCutoutConfig 获取群聊抠图显示配置
func (s *characterImageService) GetGroupChatCutoutConfig(ctx context.Context, req *domain.GroupChatCutoutRequest) (*domain.GroupChatCutoutResult, error) {
	s.logger.WithFields(logrus.Fields{
		"group_chat_id": req.GroupChatID,
		"character_ids": req.CharacterIDs,
	}).Debug("Getting group chat cutout config")

	result := &domain.GroupChatCutoutResult{
		GroupChatID:      req.GroupChatID,
		CharacterCutouts: []*domain.CharacterCutout{},
	}

	// 获取群聊信息
	groupChat, err := s.characterRepo.GetGroupChatByID(ctx, req.GroupChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group chat: %w", err)
	}

	// 群聊使用全局背景
	if groupChat.BackgroundImageURL != nil {
		result.GlobalBackgroundURL = *groupChat.BackgroundImageURL
	}

	// 获取每个角色的抠图配置
	for _, characterID := range req.CharacterIDs {
		character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get character for cutout config")
			continue
		}

		cutout := &domain.CharacterCutout{
			CharacterID:   characterID,
			CharacterName: character.Name,
			IsVisible:     false, // 默认不显示
			AnimationType: "fade_in",
			AnimationDuration: 300, // 300ms
		}

		// 设置抠图URL
		if character.CutoutImageURL != nil && *character.CutoutImageURL != "" {
			cutout.CutoutImageURL = *character.CutoutImageURL
			cutout.HasCutout = true
		} else {
			cutout.HasCutout = false
			cutout.FallbackText = character.Name // 如果没有抠图，显示角色名
		}

		result.CharacterCutouts = append(result.CharacterCutouts, cutout)
	}

	return result, nil
}

// HandleCharacterSpeaking 处理角色发言时的抠图切换
func (s *characterImageService) HandleCharacterSpeaking(ctx context.Context, req *domain.CharacterSpeakingRequest) (*domain.CharacterSpeakingResult, error) {
	s.logger.WithFields(logrus.Fields{
		"group_chat_id":       req.GroupChatID,
		"speaking_character":  req.SpeakingCharacterID,
		"previous_character":  req.PreviousCharacterID,
	}).Debug("Handling character speaking cutout switch")

	result := &domain.CharacterSpeakingResult{
		GroupChatID:         req.GroupChatID,
		SpeakingCharacterID: req.SpeakingCharacterID,
		CutoutChanges:       []*domain.CutoutChange{},
	}

	// 获取发言角色信息
	speakingCharacter, err := s.characterRepo.GetCharacterByID(ctx, req.SpeakingCharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get speaking character: %w", err)
	}

	// 1. 显示当前发言角色的抠图（300ms淡入）
	if speakingCharacter.CutoutImageURL != nil && *speakingCharacter.CutoutImageURL != "" {
		result.CutoutChanges = append(result.CutoutChanges, &domain.CutoutChange{
			CharacterID:   req.SpeakingCharacterID,
			Action:        "show",
			CutoutURL:     *speakingCharacter.CutoutImageURL,
			Animation:     "fade_in",
			Duration:      300,
			ZIndex:        100, // 最高层级
			Opacity:       1.0,
		})
	}

	// 2. 隐藏或淡化之前发言角色的抠图
	if req.PreviousCharacterID != nil && *req.PreviousCharacterID != req.SpeakingCharacterID {
		result.CutoutChanges = append(result.CutoutChanges, &domain.CutoutChange{
			CharacterID: *req.PreviousCharacterID,
			Action:      "fade",
			Animation:   "fade_out",
			Duration:    200,
			ZIndex:      50,
			Opacity:     0.3, // 半透明
		})
	}

	// 3. 处理下一位预告（如果有）
	if req.NextCharacterID != nil {
		nextCharacter, err := s.characterRepo.GetCharacterByID(ctx, *req.NextCharacterID)
		if err == nil && nextCharacter.CutoutImageURL != nil {
			result.CutoutChanges = append(result.CutoutChanges, &domain.CutoutChange{
				CharacterID: *req.NextCharacterID,
				Action:      "preview",
				CutoutURL:   *nextCharacter.CutoutImageURL,
				Animation:   "fade_in",
				Duration:    200,
				ZIndex:      30,
				Opacity:     0.5, // 半透明预览
			})
		}
	}

	// 4. 其他角色完全隐藏
	for _, otherCharacterID := range req.OtherCharacterIDs {
		if otherCharacterID != req.SpeakingCharacterID &&
		   (req.PreviousCharacterID == nil || otherCharacterID != *req.PreviousCharacterID) &&
		   (req.NextCharacterID == nil || otherCharacterID != *req.NextCharacterID) {
			
			result.CutoutChanges = append(result.CutoutChanges, &domain.CutoutChange{
				CharacterID: otherCharacterID,
				Action:      "hide",
				Animation:   "fade_out",
				Duration:    150,
				ZIndex:      0,
				Opacity:     0.0,
			})
		}
	}

	return result, nil
}

// GenerateCharacterImages 自动生成角色两图（可选功能）
func (s *characterImageService) GenerateCharacterImages(ctx context.Context, req *domain.ImageGenerationRequest) (*domain.ImageGenerationResult, error) {
	s.logger.WithFields(logrus.Fields{
		"character_id": req.CharacterID,
		"generate_bg":  req.GenerateBackground,
		"generate_cutout": req.GenerateCutout,
	}).Info("Generating character images")

	result := &domain.ImageGenerationResult{
		CharacterID: req.CharacterID,
		Success:     false,
		GeneratedImages: []*domain.GeneratedImage{},
	}

	// 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 生成背景图
	if req.GenerateBackground {
		bgPrompt := s.buildBackgroundPrompt(character, req.BackgroundStyle)
		// 这里应该调用实际的图像生成服务
		// bgResult, err := s.mediaService.GenerateImage(ctx, bgPrompt, "background")
		
		result.GeneratedImages = append(result.GeneratedImages, &domain.GeneratedImage{
			Type:        "background",
			URL:         "https://generated-bg-url.jpg", // 模拟生成的URL
			Prompt:      bgPrompt,
			Style:       req.BackgroundStyle,
		})
	}

	// 生成抠图
	if req.GenerateCutout {
		cutoutPrompt := s.buildCutoutPrompt(character, req.CutoutStyle)
		// cutoutResult, err := s.mediaService.GenerateImage(ctx, cutoutPrompt, "cutout")
		
		result.GeneratedImages = append(result.GeneratedImages, &domain.GeneratedImage{
			Type:        "cutout",
			URL:         "https://generated-cutout-url.png", // 模拟生成的URL
			Prompt:      cutoutPrompt,
			Style:       req.CutoutStyle,
		})
	}

	result.Success = len(result.GeneratedImages) > 0
	return result, nil
}

// 辅助方法

func (s *characterImageService) isValidImageURL(url string) bool {
	return strings.HasPrefix(url, "http") && 
		   (strings.HasSuffix(url, ".jpg") || 
		    strings.HasSuffix(url, ".jpeg") || 
		    strings.HasSuffix(url, ".png"))
}

func (s *characterImageService) isValidCutoutURL(url string) bool {
	return strings.HasPrefix(url, "http") && strings.HasSuffix(url, ".png")
}

func (s *characterImageService) validateImageAccessibility(ctx context.Context, url string) bool {
	// 这里应该实现实际的URL可访问性检查
	// 暂时返回true
	return true
}

func (s *characterImageService) getDefaultChatBackground() string {
	return "https://yunai-assets.com/defaults/chat_background.jpg"
}

func (s *characterImageService) buildBackgroundPrompt(character *domain.Character, style string) string {
	prompt := fmt.Sprintf("为角色%s生成背景图，", character.Name)
	if character.Personality != nil {
		prompt += fmt.Sprintf("性格：%s，", *character.Personality)
	}
	prompt += fmt.Sprintf("风格：%s", style)
	return prompt
}

func (s *characterImageService) buildCutoutPrompt(character *domain.Character, style string) string {
	prompt := fmt.Sprintf("为角色%s生成透明背景立绘，", character.Name)
	if character.Personality != nil {
		prompt += fmt.Sprintf("性格：%s，", *character.Personality)
	}
	prompt += fmt.Sprintf("风格：%s，透明背景PNG格式", style)
	return prompt
}
