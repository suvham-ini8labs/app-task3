package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"user-service/internal/auth"
	"user-service/internal/models"
	"user-service/pkg/logger"
	"user-service/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUserRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log, _ := logger.New("info", "console")
	svc := NewUserService(mockRepo, "secret", 1*time.Hour, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &models.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once()

		user, err := svc.Register(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Username, user.Username)
		assert.Equal(t, req.Email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("EmailAlreadyRegistered", func(t *testing.T) {
		req := &models.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		existingUser := &models.User{Email: req.Email}
		mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil).Once()

		user, err := svc.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "email already registered", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("ValidationFailures", func(t *testing.T) {
		tests := []struct {
			name     string
			req      *models.RegisterRequest
			expected string
		}{
			{"UsernameRequired", &models.RegisterRequest{Email: "a@b.com", Password: "pwd"}, "username is required"},
			{"EmailRequired", &models.RegisterRequest{Username: "u", Password: "pwd"}, "email is required"},
			{"PasswordRequired", &models.RegisterRequest{Username: "u", Email: "a@b.com"}, "password is required"},
			{"PasswordTooShort", &models.RegisterRequest{Username: "u", Email: "a@b.com", Password: "123"}, "password must be at least 6 characters"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				user, err := svc.Register(ctx, tc.req)
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.Equal(t, tc.expected, err.Error())
			})
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		req := &models.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("db error")).Once()

		user, err := svc.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "db error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log, _ := logger.New("info", "console")
	svc := NewUserService(mockRepo, "secret", 1*time.Hour, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := utils.HashPassword(password)
		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: password,
		}

		user := &models.User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(user, nil).Once()

		resp, err := svc.Login(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, user.Username, resp.User.Username)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ValidationFailures", func(t *testing.T) {
		tests := []struct {
			name     string
			req      *models.LoginRequest
			expected string
		}{
			{"EmailRequired", &models.LoginRequest{Password: "pwd"}, "email is required"},
			{"PasswordRequired", &models.LoginRequest{Email: "a@b.com"}, "password is required"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := svc.Login(ctx, tc.req)
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Equal(t, tc.expected, err.Error())
			})
		}
	})

	t.Run("InvalidCredentials_UserNotFound", func(t *testing.T) {
		req := &models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, nil).Once()

		resp, err := svc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "invalid credentials", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidCredentials_WrongPassword", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := utils.HashPassword("otherpassword")
		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: password,
		}

		user := &models.User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(user, nil).Once()

		resp, err := svc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "invalid credentials", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("db error")).Once()

		resp, err := svc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "db error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log, _ := logger.New("info", "console")
	svc := NewUserService(mockRepo, "secret", 1*time.Hour, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		id := 1
		user := &models.User{
			ID:       id,
			Username: "testuser",
			Email:    "test@example.com",
		}

		mockRepo.On("GetByID", ctx, id).Return(user, nil).Once()

		result, err := svc.GetUser(ctx, id)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidID", func(t *testing.T) {
		result, err := svc.GetUser(ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid user id", err.Error())
	})

	t.Run("NotFound", func(t *testing.T) {
		id := 99
		mockRepo.On("GetByID", ctx, id).Return(nil, nil).Once()

		result, err := svc.GetUser(ctx, id)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "user not found", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		id := 1
		mockRepo.On("GetByID", ctx, id).Return(nil, errors.New("db error")).Once()

		result, err := svc.GetUser(ctx, id)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "db error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log, _ := logger.New("info", "console")
	svc := NewUserService(mockRepo, "secret", 1*time.Hour, log)

	ctx := context.Background()

	t.Run("Success_Username", func(t *testing.T) {
		id := 1
		existingUser := &models.User{
			ID:           id,
			Username:     "oldname",
			Email:        "old@example.com",
			PasswordHash: "oldhash",
		}

		req := &models.UpdateUserRequest{
			Username: "newname",
		}

		mockRepo.On("GetByID", ctx, id).Return(existingUser, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
			return u.Username == "newname" && u.Email == "old@example.com"
		})).Return(nil).Once()

		result, err := svc.UpdateUser(ctx, id, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "newname", result.Username)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success_Full", func(t *testing.T) {
		id := 1
		existingUser := &models.User{ID: id}
		req := &models.UpdateUserRequest{
			Username: "newname",
			Email:    "new@example.com",
			Password: "newpassword123",
		}

		mockRepo.On("GetByID", ctx, id).Return(existingUser, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once()

		result, err := svc.UpdateUser(ctx, id, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "newname", result.Username)
		assert.Equal(t, "new@example.com", result.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidID", func(t *testing.T) {
		result, err := svc.UpdateUser(ctx, 0, &models.UpdateUserRequest{})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid user id", err.Error())
	})

	t.Run("UserNotFound", func(t *testing.T) {
		id := 99
		mockRepo.On("GetByID", ctx, id).Return(nil, nil).Once()

		result, err := svc.UpdateUser(ctx, id, &models.UpdateUserRequest{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "user not found", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("PasswordTooShort", func(t *testing.T) {
		id := 1
		existingUser := &models.User{ID: id}
		req := &models.UpdateUserRequest{
			Password: "123",
		}

		mockRepo.On("GetByID", ctx, id).Return(existingUser, nil).Once()

		result, err := svc.UpdateUser(ctx, id, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "password must be at least 6 characters", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError_GetByID", func(t *testing.T) {
		id := 1
		mockRepo.On("GetByID", ctx, id).Return(nil, errors.New("db error")).Once()

		result, err := svc.UpdateUser(ctx, id, &models.UpdateUserRequest{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "db error", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError_Update", func(t *testing.T) {
		id := 1
		mockRepo.On("GetByID", ctx, id).Return(&models.User{ID: id}, nil).Once()
		mockRepo.On("Update", ctx, mock.Anything).Return(errors.New("update error")).Once()

		result, err := svc.UpdateUser(ctx, id, &models.UpdateUserRequest{Username: "new"})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "update error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestValidateToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log, _ := logger.New("info", "console")
	secret := "secret"
	svc := NewUserService(mockRepo, secret, 1*time.Hour, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := 1
		username := "testuser"
		email := "test@example.com"

		jwtManager := auth.NewJWTManager(secret, 1*time.Hour)
		token, _ := jwtManager.Generate(userID, username, email)

		user := &models.User{
			ID:       userID,
			Username: username,
			Email:    email,
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil).Once()

		result, err := svc.ValidateToken(ctx, token)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		token := "invalid-token"

		result, err := svc.ValidateToken(ctx, token)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		secret := "secret"
		jwtManager := auth.NewJWTManager(secret, 1*time.Hour)
		token, _ := jwtManager.Generate(1, "u", "e")

		mockRepo.On("GetByID", ctx, 1).Return(nil, nil).Once()

		result, err := svc.ValidateToken(ctx, token)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "user not found", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		secret := "secret"
		jwtManager := auth.NewJWTManager(secret, 1*time.Hour)
		token, _ := jwtManager.Generate(1, "u", "e")

		mockRepo.On("GetByID", ctx, 1).Return(nil, errors.New("db error")).Once()

		result, err := svc.ValidateToken(ctx, token)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "db error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestHealth(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log, _ := logger.New("info", "console")
	svc := NewUserService(mockRepo, "secret", 1*time.Hour, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Health", ctx).Return(nil).Once()
		err := svc.Health(ctx)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		mockRepo.On("Health", ctx).Return(errors.New("db error")).Once()
		err := svc.Health(ctx)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
