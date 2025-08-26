package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// UserService 用户服务接口
type UserService interface {
	// 用户注册（自动创建钱包）
	RegisterUser(ctx context.Context, req *domain.RegisterUserRequest) (*domain.User, error)

	// 用户基本操作
	GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// 用户验证
	ValidateUser(ctx context.Context, username, password string) (*domain.User, error)
}

type userService struct {
	userRepo   repository.UserRepository
	walletRepo repository.WalletRepository
	logger     *logrus.Logger
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository, walletRepo repository.WalletRepository, logger *logrus.Logger) UserService {
	return &userService{
		userRepo:   userRepo,
		walletRepo: walletRepo,
		logger:     logger,
	}
}

// RegisterUser 用户注册（自动创建钱包）
func (s *userService) RegisterUser(ctx context.Context, req *domain.RegisterUserRequest) (*domain.User, error) {
	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return nil, domain.ErrUsernameExists
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, domain.ErrEmailExists
	}

	// 创建用户对象
	user := &domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.PasswordHash, // 应该已经在外部加密
		UserType:     domain.UserType(req.UserType),
		Nickname:     req.Nickname,
		AvatarURL:    req.AvatarURL,
		Bio:          req.Bio,
		IsActive:     true,
		IsBanned:     false,
	}

	// 创建用户（钱包会通过数据库触发器自动创建）
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("User registered successfully with auto-created wallet")

	return user, nil
}

// GetUser 获取用户
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetUserByUsername 根据用户名获取用户
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// GetUserByEmail 根据邮箱获取用户
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.userRepo.Update(ctx, user)
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// TODO: 应该同时删除钱包和相关数据
	return s.userRepo.Delete(ctx, id)
}

// ValidateUser 验证用户
func (s *userService) ValidateUser(ctx context.Context, username, password string) (*domain.User, error) {
	// TODO: 实现密码验证逻辑
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 这里应该验证密码哈希
	// if !verifyPassword(password, user.PasswordHash) {
	//     return nil, domain.ErrInvalidCredentials
	// }

	return user, nil
}
