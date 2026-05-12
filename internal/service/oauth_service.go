package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"iam/internal/dto"
	"iam/internal/model"
	jwtpkg "iam/internal/pkg/jwt"
	"iam/internal/pkg/random"
	"iam/internal/repository"

	"github.com/redis/go-redis/v9"
)

type OAuthService interface {
	Authorize(ctx context.Context, query dto.AuthorizeQuery, userID uint64) (*dto.AuthorizeResponse, error)
	Token(ctx context.Context, req dto.TokenRequest) (*dto.TokenResponse, error)
	UserInfo(ctx context.Context, token string) (*dto.CurrentUser, error)
}

type oauthService struct {
	clients   repository.OAuthClientRepository
	auth      AuthService
	users     repository.UserRepository
	redis     *redis.Client
	jwt       *jwtpkg.Manager
	codeTTL   time.Duration
	redisCode string
}

func NewOAuthService(clients repository.OAuthClientRepository, auth AuthService, users repository.UserRepository, redis *redis.Client, jwt *jwtpkg.Manager, codeExpireSeconds int) OAuthService {
	return &oauthService{clients: clients, auth: auth, users: users, redis: redis, jwt: jwt, codeTTL: time.Duration(codeExpireSeconds) * time.Second, redisCode: "iam:oauth:code:"}
}

func (s *oauthService) Authorize(ctx context.Context, query dto.AuthorizeQuery, userID uint64) (*dto.AuthorizeResponse, error) {
	if query.ResponseType != "code" {
		return nil, fmt.Errorf("unsupported response_type")
	}
	client, err := s.validateClient(ctx, query.ClientID, query.RedirectURI)
	if err != nil {
		return nil, err
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Status != 1 {
		return nil, fmt.Errorf("user disabled")
	}

	code, err := random.Hex(16)
	if err != nil {
		return nil, err
	}
	payload := dto.OAuthCodePayload{ClientID: client.ClientID, RedirectURI: client.RedirectURI, UserID: user.ID, Username: user.Username, Scope: query.Scope, State: query.State}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if err := s.redis.Set(ctx, s.redisCode+code, b, s.codeTTL).Err(); err != nil {
		return nil, err
	}

	redirectTo := appendQuery(client.RedirectURI, map[string]string{"code": code, "state": query.State})
	return &dto.AuthorizeResponse{Code: code, State: query.State, RedirectTo: redirectTo, ExpiresIn: int(s.codeTTL.Seconds())}, nil
}

func (s *oauthService) Token(ctx context.Context, req dto.TokenRequest) (*dto.TokenResponse, error) {
	if req.GrantType != "authorization_code" {
		return nil, fmt.Errorf("unsupported grant_type")
	}
	client, err := s.validateClient(ctx, req.ClientID, req.RedirectURI)
	if err != nil {
		return nil, err
	}
	if client.ClientSecret != req.ClientSecret {
		return nil, fmt.Errorf("invalid client_secret")
	}

	key := s.redisCode + req.Code
	raw, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("invalid or expired code")
	}
	var payload dto.OAuthCodePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	if payload.ClientID != req.ClientID || payload.RedirectURI != req.RedirectURI {
		return nil, fmt.Errorf("authorization code mismatch")
	}
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, payload.UserID)
	if err != nil {
		return nil, err
	}
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Code)
	}
	tokenID, err := random.Hex(12)
	if err != nil {
		return nil, err
	}
	accessToken, expiresIn, err := s.jwt.Sign(user.ID, user.Username, roles, user.Status, tokenID)
	if err != nil {
		return nil, err
	}
	return &dto.TokenResponse{AccessToken: accessToken, TokenType: "Bearer", ExpiresIn: expiresIn, Scope: payload.Scope}, nil
}

func (s *oauthService) UserInfo(ctx context.Context, token string) (*dto.CurrentUser, error) {
	claims, err := s.auth.ParseToken(strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	userID, err := ParseUserID(claims.Subject)
	if err != nil {
		return nil, err
	}
	return s.auth.Me(ctx, userID)
}

func (s *oauthService) validateClient(ctx context.Context, clientID string, redirectURI string) (*model.OAuthClient, error) {
	client, err := s.clients.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client_id")
	}
	if client.Status != 1 {
		return nil, fmt.Errorf("client disabled")
	}
	if client.RedirectURI != redirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}
	return client, nil
}

func appendQuery(rawURL string, values map[string]string) string {
	u, _ := url.Parse(rawURL)
	q := u.Query()
	for k, v := range values {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
