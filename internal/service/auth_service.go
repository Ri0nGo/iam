package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"iam/internal/dto"
	jwtpkg "iam/internal/pkg/jwt"
	"iam/internal/pkg/password"
	"iam/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Logout(ctx context.Context, token string) error
	Me(ctx context.Context, userID uint64) (*dto.CurrentUser, error)
	ParseToken(token string) (*jwtpkg.Claims, error)
}

type authService struct {
	users      repository.UserRepository
	identities repository.AuthIdentityRepository
	redis      *redis.Client
	jwt        *jwtpkg.Manager
	issuer     string
	failLimit  int
	failWindow time.Duration
}

func NewAuthService(users repository.UserRepository, identities repository.AuthIdentityRepository, redis *redis.Client, jwt *jwtpkg.Manager, failLimit int, failWindowSeconds int) AuthService {
	return &authService{users: users, identities: identities, redis: redis, jwt: jwt, failLimit: failLimit, failWindow: time.Duration(failWindowSeconds) * time.Second}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	key := fmt.Sprintf("iam:login:fail:%s", req.Username)
	count, _ := s.redis.Get(ctx, key).Int()
	if s.failLimit > 0 && count >= s.failLimit {
		return nil, fmt.Errorf("login failed too many times, please retry later")
	}

	identity, err := s.identities.GetByIdentity(ctx, "password", req.Username)
	if err != nil {
		s.redis.Incr(ctx, key)
		s.redis.Expire(ctx, key, s.failWindow)
		return nil, fmt.Errorf("invalid username or password")
	}
	if !password.Verify(identity.Credential, req.Password) {
		s.redis.Incr(ctx, key)
		s.redis.Expire(ctx, key, s.failWindow)
		return nil, fmt.Errorf("invalid username or password")
	}

	user, err := s.users.GetByID(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	if user.Status != 1 {
		return nil, fmt.Errorf("user disabled")
	}

	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Code)
	}
	tokenID := fmt.Sprintf("%d", time.Now().UnixNano())
	token, expiresIn, err := s.jwt.Sign(user.ID, user.Username, roles, user.Status, tokenID)
	if err != nil {
		return nil, err
	}
	s.redis.Del(ctx, key)

	return &dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User: dto.CurrentUser{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Status:      user.Status,
			Roles:       roles,
		},
	}, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	claims, err := s.jwt.Parse(token)
	if err != nil {
		return err
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	return s.redis.Set(ctx, fmt.Sprintf("iam:token:blacklist:%s", claims.ID), 1, ttl).Err()
}

func (s *authService) Me(ctx context.Context, userID uint64) (*dto.CurrentUser, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Code)
	}
	return &dto.CurrentUser{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Status: user.Status, Roles: roles}, nil
}

func (s *authService) ParseToken(token string) (*jwtpkg.Claims, error) {
	return s.jwt.Parse(strings.TrimSpace(token))
}

func ParseUserID(subject string) (uint64, error) {
	return strconv.ParseUint(subject, 10, 64)
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
