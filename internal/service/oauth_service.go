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
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.TokenResponse, error)
	CheckToken(ctx context.Context, token string, openID string) error
	UserInfo(ctx context.Context, token string, openID string) (*dto.OAuthUserInfo, error)
}

type oauthService struct {
	clients      repository.OAuthClientRepository
	auth         AuthService
	users        repository.UserRepository
	redis        *redis.Client
	jwt          *jwtpkg.Manager
	codeTTL      time.Duration
	redisCode    string
	redisRefresh string
}

func NewOAuthService(clients repository.OAuthClientRepository, auth AuthService, users repository.UserRepository, redis *redis.Client, jwt *jwtpkg.Manager, codeExpireSeconds int) OAuthService {
	return &oauthService{clients: clients, auth: auth, users: users, redis: redis, jwt: jwt, codeTTL: time.Duration(codeExpireSeconds) * time.Second, redisCode: "iam:oauth:code:", redisRefresh: "iam:oauth:refresh:"}
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
	client, err := s.validateClientID(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	if client.ClientSecret != req.Secret {
		return nil, fmt.Errorf("invalid secret")
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
	if payload.ClientID != req.ClientID {
		return nil, fmt.Errorf("authorization code mismatch")
	}
	if req.RedirectURI != "" && payload.RedirectURI != req.RedirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, payload.UserID)
	if err != nil {
		return nil, err
	}
	if user.OpenID == nil || *user.OpenID == "" {
		return nil, fmt.Errorf("user openid missing")
	}
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Code)
	}
	tokenID, err := random.Hex(12)
	if err != nil {
		return nil, err
	}
	accessToken, expiresIn, err := s.jwt.SignWithOptions(user.ID, user.Username, roles, user.Status, tokenID, jwtpkg.SignOptions{TokenUse: jwtpkg.TokenUseOAuth2, ClientID: client.ClientID, Scope: payload.Scope})
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.createRefreshToken(ctx, dto.OAuthRefreshPayload{ClientID: client.ClientID, UserID: user.ID, Username: user.Username, OpenID: *user.OpenID, Scope: payload.Scope})
	if err != nil {
		return nil, err
	}
	return &dto.TokenResponse{AccessToken: accessToken, ExpiresIn: expiresIn, RefreshToken: refreshToken, OpenID: *user.OpenID, Scope: payload.Scope}, nil
}

func (s *oauthService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.TokenResponse, error) {
	if req.GrantType != "refresh_token" {
		return nil, fmt.Errorf("unsupported grant_type")
	}
	client, err := s.validateClientID(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	raw, err := s.redis.Get(ctx, s.redisRefresh+req.RefreshToken).Result()
	if err != nil {
		return nil, fmt.Errorf("invalid refresh_token")
	}
	var payload dto.OAuthRefreshPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	if payload.ClientID != client.ClientID {
		return nil, fmt.Errorf("refresh_token mismatch")
	}
	user, err := s.users.GetByID(ctx, payload.UserID)
	if err != nil {
		return nil, err
	}
	if user.OpenID == nil || *user.OpenID != payload.OpenID {
		return nil, fmt.Errorf("openid mismatch")
	}
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Code)
	}
	tokenID, err := random.Hex(12)
	if err != nil {
		return nil, err
	}
	accessToken, expiresIn, err := s.jwt.SignWithOptions(user.ID, user.Username, roles, user.Status, tokenID, jwtpkg.SignOptions{TokenUse: jwtpkg.TokenUseOAuth2, ClientID: client.ClientID, Scope: payload.Scope})
	if err != nil {
		return nil, err
	}
	return &dto.TokenResponse{AccessToken: accessToken, ExpiresIn: expiresIn, RefreshToken: req.RefreshToken, OpenID: payload.OpenID, Scope: payload.Scope}, nil
}

func (s *oauthService) CheckToken(ctx context.Context, token string, openID string) error {
	_, err := s.parseOAuthTokenForOpenID(ctx, token, openID)
	return err
}

func (s *oauthService) UserInfo(ctx context.Context, token string, openID string) (*dto.OAuthUserInfo, error) {
	user, err := s.parseOAuthTokenForOpenID(ctx, token, openID)
	if err != nil {
		return nil, err
	}
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Code)
	}
	return &dto.OAuthUserInfo{OpenID: *user.OpenID, Username: user.Username, DisplayName: user.DisplayName, Status: user.Status, Roles: roles}, nil
}

func (s *oauthService) parseOAuthTokenForOpenID(ctx context.Context, token string, openID string) (*model.User, error) {
	claims, err := s.auth.ParseToken(strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	if claims.TokenUse != jwtpkg.TokenUseOAuth2 {
		return nil, fmt.Errorf("invalid token use")
	}
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil, fmt.Errorf("missing openid")
	}
	userID, err := ParseUserID(claims.Subject)
	if err != nil {
		return nil, err
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.OpenID == nil || *user.OpenID != openID {
		return nil, fmt.Errorf("openid mismatch")
	}
	return user, nil
}

func (s *oauthService) validateClient(ctx context.Context, clientID string, redirectURI string) (*model.OAuthClient, error) {
	client, err := s.validateClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if client.RedirectURI != redirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}
	return client, nil
}

func (s *oauthService) validateClientID(ctx context.Context, clientID string) (*model.OAuthClient, error) {
	client, err := s.clients.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client_id")
	}
	if client.Status != 1 {
		return nil, fmt.Errorf("client disabled")
	}
	return client, nil
}

func (s *oauthService) createRefreshToken(ctx context.Context, payload dto.OAuthRefreshPayload) (string, error) {
	refreshToken, err := random.Hex(24)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if err := s.redis.Set(ctx, s.redisRefresh+refreshToken, b, 180*24*time.Hour).Err(); err != nil {
		return "", err
	}
	return refreshToken, nil
}

func appendQuery(rawURL string, values map[string]string) string {
	u, _ := url.Parse(rawURL)
	if strings.HasPrefix(u.Fragment, "/") {
		u.Fragment = appendFragmentQuery(u.Fragment, values)
		return u.String()
	}
	q := u.Query()
	for k, v := range values {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func appendFragmentQuery(fragment string, values map[string]string) string {
	path, rawQuery, _ := strings.Cut(fragment, "?")
	q, _ := url.ParseQuery(rawQuery)
	for k, v := range values {
		if v != "" {
			q.Set(k, v)
		}
	}
	if encoded := q.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}
