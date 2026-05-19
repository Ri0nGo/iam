package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
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

func TestAppendQuerySupportsHashCallback(t *testing.T) {
	redirectTo := appendQuery("http://127.0.0.1:8000/#/oauth/callback", map[string]string{"code": "abc", "state": "xyz"})
	parsed, err := url.Parse(redirectTo)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.RawQuery != "" {
		t.Fatalf("expected query before fragment to be empty, got %q", parsed.RawQuery)
	}
	if !strings.HasPrefix(parsed.Fragment, "/oauth/callback?") {
		t.Fatalf("expected params to be appended inside hash route, got fragment %q", parsed.Fragment)
	}
	fragmentQuery := strings.TrimPrefix(parsed.Fragment, "/oauth/callback?")
	values, err := url.ParseQuery(fragmentQuery)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("code") != "abc" || values.Get("state") != "xyz" {
		t.Fatalf("unexpected fragment query values: %s", fragmentQuery)
	}
}

func TestAppendQuerySupportsHashCallbackWithBasePath(t *testing.T) {
	redirectTo := appendQuery("http://127.0.0.1:8000/xx1/xx2/#/oauth/callback", map[string]string{"code": "abc", "state": "xyz"})
	parsed, err := url.Parse(redirectTo)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/xx1/xx2/" {
		t.Fatalf("expected base path to be preserved, got %q", parsed.Path)
	}
	if parsed.RawQuery != "" {
		t.Fatalf("expected query before fragment to be empty, got %q", parsed.RawQuery)
	}
	if !strings.HasPrefix(parsed.Fragment, "/oauth/callback?") {
		t.Fatalf("expected params to be appended inside hash route, got fragment %q", parsed.Fragment)
	}
	fragmentQuery := strings.TrimPrefix(parsed.Fragment, "/oauth/callback?")
	values, err := url.ParseQuery(fragmentQuery)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("code") != "abc" || values.Get("state") != "xyz" {
		t.Fatalf("unexpected fragment query values: %s", fragmentQuery)
	}
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
	openID := "ou_admin"
	userRepo := &fakeUserRepo{user: &model.User{Base: model.Base{ID: 1}, Username: "admin", OpenID: &openID, DisplayName: "管理员", Status: 1, Roles: []model.Role{{Code: "admin"}}}}
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

	tokenResp, err := oauthService.Token(context.Background(), dto.TokenRequest{GrantType: "authorization_code", ClientID: "system-a", Secret: "system-a-secret", Code: authResp.Code})
	if err != nil {
		t.Fatalf("token exchange failed: %v", err)
	}
	if tokenResp.AccessToken == "" || tokenResp.RefreshToken == "" {
		t.Fatal("expected access token and refresh token")
	}
	if tokenResp.OpenID != openID {
		t.Fatalf("expected openid %q, got %q", openID, tokenResp.OpenID)
	}

	if err := oauthService.CheckToken(context.Background(), tokenResp.AccessToken, tokenResp.OpenID); err != nil {
		t.Fatalf("expected token check success, got %v", err)
	}
	refreshed, err := oauthService.RefreshToken(context.Background(), dto.RefreshTokenRequest{ClientID: "system-a", GrantType: "refresh_token", RefreshToken: tokenResp.RefreshToken})
	if err != nil {
		t.Fatalf("refresh token failed: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.OpenID != openID {
		t.Fatalf("unexpected refreshed response: %+v", refreshed)
	}

	if _, err := oauthService.Token(context.Background(), dto.TokenRequest{GrantType: "authorization_code", ClientID: "system-a", Secret: "system-a-secret", Code: authResp.Code}); err == nil {
		t.Fatal("expected single-use code to fail on second exchange")
	}
}
