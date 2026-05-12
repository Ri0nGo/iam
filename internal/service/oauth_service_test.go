package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"iam/internal/dto"
	"iam/internal/model"
	jwtpkg "iam/internal/pkg/jwt"
	"iam/internal/pkg/password"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

var errNotFound = errors.New("not found")

type fakeClientRepo struct{ client *model.OAuthClient }

func (f *fakeClientRepo) GetByClientID(_ context.Context, clientID string) (*model.OAuthClient, error) {
	if f.client != nil && f.client.ClientID == clientID {
		return f.client, nil
	}
	return nil, errNotFound
}

func (f *fakeClientRepo) Create(context.Context, *model.OAuthClient) error { return nil }
func (f *fakeClientRepo) GetByID(context.Context, uint64) (*model.OAuthClient, error) {
	return f.client, nil
}
func (f *fakeClientRepo) List(context.Context, string, int) ([]model.OAuthClient, error) {
	if f.client == nil {
		return []model.OAuthClient{}, nil
	}
	return []model.OAuthClient{*f.client}, nil
}
func (f *fakeClientRepo) Update(context.Context, *model.OAuthClient) error { return nil }
func (f *fakeClientRepo) Delete(context.Context, uint64) error             { return nil }

func TestOAuthAuthorizeAndToken(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	hash, err := password.Hash("123456", 4)
	if err != nil {
		t.Fatal(err)
	}
	userRepo := &fakeUserRepo{user: &model.User{Base: model.Base{ID: 1}, Username: "admin", DisplayName: "管理员", Status: 1, Roles: []model.Role{{Code: "admin"}}}}
	identityRepo := &fakeIdentityRepo{identity: &model.AuthIdentity{UserID: 1, IdentityType: "password", Identifier: "admin", Credential: hash}}
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	jwtManager := jwtpkg.NewManager("iam", "secret", int64(time.Hour.Seconds()))
	authService := NewAuthService(userRepo, identityRepo, redisClient, jwtManager, 5, 900)
	oauthService := NewOAuthService(&fakeClientRepo{client: &model.OAuthClient{ClientID: "system-a", ClientSecret: "system-a-secret", Name: "System A", RedirectURI: "http://system-a.local/callback", Status: 1}}, authService, userRepo, redisClient, jwtManager, 300)

	authResp, err := oauthService.Authorize(context.Background(), dto.AuthorizeQuery{ResponseType: "code", ClientID: "system-a", RedirectURI: "http://system-a.local/callback", Scope: "basic", State: "xyz"}, 1)
	if err != nil {
		t.Fatalf("authorize failed: %v", err)
	}
	if authResp.Code == "" {
		t.Fatal("expected authorization code")
	}

	tokenResp, err := oauthService.Token(context.Background(), dto.TokenRequest{GrantType: "authorization_code", ClientID: "system-a", ClientSecret: "system-a-secret", Code: authResp.Code, RedirectURI: "http://system-a.local/callback"})
	if err != nil {
		t.Fatalf("token exchange failed: %v", err)
	}
	if tokenResp.AccessToken == "" {
		t.Fatal("expected access token")
	}

	if _, err := oauthService.Token(context.Background(), dto.TokenRequest{GrantType: "authorization_code", ClientID: "system-a", ClientSecret: "system-a-secret", Code: authResp.Code, RedirectURI: "http://system-a.local/callback"}); err == nil {
		t.Fatal("expected single-use code to fail on second exchange")
	}
}
