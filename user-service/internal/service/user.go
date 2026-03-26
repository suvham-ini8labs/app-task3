package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"user-service/internal/auth"
	"user-service/internal/models"
	"user-service/internal/repository"
	"user-service/pkg/logger"
	"user-service/pkg/utils"
)

type userService struct {
	repo       repository.UserRepository
	jwtManager *auth.JWTManager
	logger     *logger.Logger
}

func NewUserService(repo repository.UserRepository, jwtSecret string, jwtExpiration time.Duration, log *logger.Logger) UserService {
	return &userService{
		repo:       repo,
		jwtManager: auth.NewJWTManager(jwtSecret, jwtExpiration),
		logger:     log,
	}
}

func (s *userService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Validate input
	if req.Username == "" {
		s.logger.Debug("Registration failed: username required")
		return nil, errors.New("username is required")
	}
	if req.Email == "" {
		s.logger.Debug("Registration failed: email required")
		return nil, errors.New("email is required")
	}
	if req.Password == "" {
		s.logger.Debug("Registration failed: password required")
		return nil, errors.New("password is required")
	}
	if len(req.Password) < 6 {
		s.logger.Debug("Registration failed: password too short")
		return nil, errors.New("password must be at least 6 characters")
	}

	// Check if user exists
	existingUser, _ := s.repo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		s.logger.Info("Registration failed: email already registered", "email", req.Email)
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, err
	}

	s.logger.Info("User registered successfully", "id", user.ID, "email", user.Email)
	return user, nil
}

func (s *userService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Validate input
	if req.Email == "" {
		s.logger.Debug("Login failed: email required")
		return nil, errors.New("email is required")
	}
	if req.Password == "" {
		s.logger.Debug("Login failed: password required")
		return nil, errors.New("password is required")
	}

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to get user by email", "error", err, "email", req.Email)
		return nil, err
	}
	if user == nil {
		s.logger.Info("Login failed: user not found", "email", req.Email)
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		s.logger.Info("Login failed: invalid password", "email", req.Email)
		return nil, errors.New("invalid credentials")
	}

	// Generate token
	token, err := s.jwtManager.Generate(user.ID, user.Username, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password hash
	user.PasswordHash = ""

	s.logger.Info("User logged in successfully", "id", user.ID, "email", user.Email)
	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *userService) GetUser(ctx context.Context, id int) (*models.User, error) {
	if id <= 0 {
		s.logger.Debug("Get user failed: invalid user id", "id", id)
		return nil, errors.New("invalid user id")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err, "id", id)
		return nil, err
	}
	if user == nil {
		s.logger.Info("User not found", "id", id)
		return nil, errors.New("user not found")
	}

	// Don't return password hash
	user.PasswordHash = ""
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id int, req *models.UpdateUserRequest) (*models.User, error) {
	if id <= 0 {
		s.logger.Debug("Update user failed: invalid user id", "id", id)
		return nil, errors.New("invalid user id")
	}

	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user for update", "error", err, "id", id)
		return nil, err
	}
	if user == nil {
		s.logger.Info("User not found for update", "id", id)
		return nil, errors.New("user not found")
	}

	// Update fields
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		if len(req.Password) < 6 {
			s.logger.Debug("Update failed: password too short", "id", id)
			return nil, errors.New("password must be at least 6 characters")
		}
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			s.logger.Error("Failed to hash password", "error", err)
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = hashedPassword
	}

	// Update in database
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user", "error", err, "id", id)
		return nil, err
	}

	// Don't return password hash
	user.PasswordHash = ""
	s.logger.Info("User updated successfully", "id", id)
	return user, nil
}

func (s *userService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	claims, err := s.jwtManager.Verify(token)
	if err != nil {
		s.logger.Debug("Token validation failed", "error", err)
		return nil, err
	}

	// Get user from database
	user, err := s.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		s.logger.Error("Failed to get user for token validation", "error", err, "user_id", claims.UserID)
		return nil, err
	}
	if user == nil {
		s.logger.Info("User not found for token validation", "user_id", claims.UserID)
		return nil, errors.New("user not found")
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *userService) Health(ctx context.Context) error {
	return s.repo.Health(ctx)
}
