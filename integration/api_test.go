package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"iam/internal/dto"
	"iam/internal/handler"
	"iam/internal/model"
	jwtpkg "iam/internal/pkg/jwt"
	"iam/internal/pkg/password"
	"iam/internal/repository"
	"iam/internal/router"
	"iam/internal/service"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type responseEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type authLoginData struct {
	AccessToken string          `json:"access_token"`
	ExpiresIn   int64           `json:"expires_in"`
	User        dto.CurrentUser `json:"user"`
}

type integrationSuite struct {
	t          *testing.T
	engine     *gin.Engine
	redis      *redis.Client
	miniRedis  *miniredis.Miniredis
	db         *gorm.DB
	adminToken string
}

func TestAuthFlow(t *testing.T) {
	s := newIntegrationSuite(t)
	defer s.close()

	login := s.loginAsAdmin()
	if login.AccessToken == "" {
		t.Fatal("expected admin access token")
	}

	resp := s.doJSON(http.MethodGet, "/api/iam/auth/me", nil, login.AccessToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for /auth/me, got %d", resp.Code)
	}

	logoutResp := s.doJSON(http.MethodPost, "/api/iam/auth/logout", nil, login.AccessToken)
	if logoutResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for logout, got %d", logoutResp.Code)
	}

	blockedResp := s.doJSON(http.MethodGet, "/api/iam/auth/me", nil, login.AccessToken)
	if blockedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", blockedResp.Code)
	}
}

func TestOAuth2Flow(t *testing.T) {
	s := newIntegrationSuite(t)
	defer s.close()

	login := s.loginAsAdmin()

	authorizeResp := s.doJSON(http.MethodGet, "/api/iam/oauth/authorize?response_type=code&client_id=system-a&redirect_uri=http://system-a.local/callback&state=xyz&scope=basic", nil, login.AccessToken)
	if authorizeResp.Code != http.StatusFound {
		t.Fatalf("expected 302 for oauth authorize, got %d", authorizeResp.Code)
	}

	redirectURL, err := url.Parse(authorizeResp.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	code := redirectURL.Query().Get("code")
	if code == "" {
		t.Fatal("expected authorization code")
	}

	tokenURL := "/api/iam/oauth/token?client_id=system-a&secret=system-a-secret&grant_type=authorization_code&code=" + url.QueryEscape(code)
	tokenResp := s.doJSON(http.MethodGet, tokenURL, nil, "")
	if tokenResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for oauth token, got %d body=%s", tokenResp.Code, tokenResp.Body.String())
	}

	var tokenEnvelope responseEnvelope
	decodeJSON(t, tokenResp.Body.Bytes(), &tokenEnvelope)
	var tokenData dto.TokenResponse
	decodeJSON(t, tokenEnvelope.Data, &tokenData)
	if tokenData.AccessToken == "" || tokenData.RefreshToken == "" {
		t.Fatal("expected access token and refresh token from oauth token endpoint")
	}
	if tokenData.OpenID != "ou_admin" {
		t.Fatalf("expected openid ou_admin from oauth token endpoint, got %q", tokenData.OpenID)
	}

	loginTokenUserinfoResp := s.doJSON(http.MethodGet, "/api/iam/oauth/userinfo?access_token="+url.QueryEscape(login.AccessToken), nil, "")
	if loginTokenUserinfoResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected login token to be rejected by oauth userinfo, got %d", loginTokenUserinfoResp.Code)
	}

	userinfoResp := s.doJSON(http.MethodGet, "/api/iam/oauth/userinfo?access_token="+url.QueryEscape(tokenData.AccessToken)+"&openid="+url.QueryEscape(tokenData.OpenID), nil, "")
	if userinfoResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for oauth userinfo, got %d", userinfoResp.Code)
	}
	checkResp := s.doJSON(http.MethodGet, "/api/iam/oauth/auth?access_token="+url.QueryEscape(tokenData.AccessToken)+"&openid="+url.QueryEscape(tokenData.OpenID), nil, "")
	if checkResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for oauth auth, got %d body=%s", checkResp.Code, checkResp.Body.String())
	}
	refreshResp := s.doJSON(http.MethodGet, "/api/iam/oauth/refresh_token?client_id=system-a&grant_type=refresh_token&refresh_token="+url.QueryEscape(tokenData.RefreshToken), nil, "")
	if refreshResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for oauth refresh_token, got %d body=%s", refreshResp.Code, refreshResp.Body.String())
	}

	secondTokenResp := s.doJSON(http.MethodGet, tokenURL, nil, "")
	if secondTokenResp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for reused authorization code, got %d", secondTokenResp.Code)
	}
}

func TestUserAndRoleManagementFlow(t *testing.T) {
	s := newIntegrationSuite(t)
	defer s.close()

	admin := s.loginAsAdmin()

	roleResp := s.doJSON(http.MethodPost, "/api/iam/roles", map[string]any{
		"code":   "operator",
		"name":   "运营角色",
		"remark": "业务运营",
	}, admin.AccessToken)
	if roleResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for create role, got %d body=%s", roleResp.Code, roleResp.Body.String())
	}

	createUserResp := s.doJSON(http.MethodPost, "/api/iam/users", map[string]any{
		"username":     "alice",
		"password":     "123456",
		"display_name": "Alice",
		"email":        "alice@example.com",
		"mobile":       "13800000000",
		"remark":       "demo user",
		"role_codes":   []string{"operator"},
	}, admin.AccessToken)
	if createUserResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for create user, got %d body=%s", createUserResp.Code, createUserResp.Body.String())
	}

	var userEnvelope responseEnvelope
	decodeJSON(t, createUserResp.Body.Bytes(), &userEnvelope)
	var createdUser model.User
	decodeJSON(t, userEnvelope.Data, &createdUser)
	if createdUser.ID == 0 {
		t.Fatal("expected created user id")
	}

	listResp := s.doJSON(http.MethodGet, "/api/iam/users?keyword=alice", nil, admin.AccessToken)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for list users, got %d", listResp.Code)
	}

	rolesResp := s.doJSON(http.MethodGet, fmt.Sprintf("/api/iam/users/%d/roles", createdUser.ID), nil, admin.AccessToken)
	if rolesResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for user roles, got %d", rolesResp.Code)
	}

	bindResp := s.doJSON(http.MethodPut, fmt.Sprintf("/api/iam/users/%d/roles", createdUser.ID), map[string]any{
		"role_codes": []string{"admin", "operator"},
	}, admin.AccessToken)
	if bindResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for bind roles, got %d", bindResp.Code)
	}

	statusResp := s.doJSON(http.MethodPut, fmt.Sprintf("/api/iam/users/%d/status", createdUser.ID), map[string]any{
		"status": 2,
	}, admin.AccessToken)
	if statusResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for update status, got %d", statusResp.Code)
	}

	passwordResp := s.doJSON(http.MethodPut, fmt.Sprintf("/api/iam/users/%d/password", createdUser.ID), map[string]any{
		"password": "new-password",
	}, admin.AccessToken)
	if passwordResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for reset password, got %d", passwordResp.Code)
	}

	loginResp := s.doJSON(http.MethodPost, "/api/iam/auth/login", map[string]any{
		"username": "alice",
		"password": "new-password",
	}, "")
	if loginResp.Code != http.StatusBadRequest {
		t.Fatalf("expected disabled user login to fail with 400, got %d", loginResp.Code)
	}
}

func newIntegrationSuite(t *testing.T) *integrationSuite {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Role{}, &model.UserRole{}, &model.AuthIdentity{}, &model.OAuthClient{}); err != nil {
		t.Fatal(err)
	}
	seedTestData(t, db)

	jwtManager := jwtpkg.NewManager("iam", "test-secret", 7200)
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	identityRepo := repository.NewAuthIdentityRepository(db)
	clientRepo := repository.NewOAuthClientRepository(db)

	authService := service.NewAuthService(userRepo, identityRepo, redisClient, jwtManager, 5, 900)
	userService := service.NewUserService(userRepo, roleRepo, identityRepo, 4)
	roleService := service.NewRoleService(roleRepo, userRepo)
	oauthService := service.NewOAuthService(clientRepo, authService, userRepo, redisClient, jwtManager, 300)
	authApplicationService := service.NewAuthApplicationService(clientRepo)

	engine := router.New(authService, redisClient, router.Handlers{
		Auth:            handler.NewAuthHandler(authService),
		User:            handler.NewUserHandler(userService),
		Role:            handler.NewRoleHandler(roleService),
		OAuth:           handler.NewOAuthHandler(oauthService, authService, redisClient, "http://localhost:5173/login"),
		AuthApplication: handler.NewAuthApplicationHandler(authApplicationService),
	})

	return &integrationSuite{t: t, engine: engine, redis: redisClient, miniRedis: mr, db: db}
}

func (s *integrationSuite) close() {
	_ = s.redis.Close()
	s.miniRedis.Close()
	if sqlDB, err := s.db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func (s *integrationSuite) loginAsAdmin() authLoginData {
	resp := s.doJSON(http.MethodPost, "/api/iam/auth/login", map[string]any{
		"username": "admin",
		"password": "123456",
	}, "")
	if resp.Code != http.StatusOK {
		s.t.Fatalf("expected 200 for admin login, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope responseEnvelope
	decodeJSON(s.t, resp.Body.Bytes(), &envelope)
	var data authLoginData
	decodeJSON(s.t, envelope.Data, &data)
	return data
}

func (s *integrationSuite) doJSON(method string, path string, payload any, token string) *httptest.ResponseRecorder {
	s.t.Helper()
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		buf, err := json.Marshal(payload)
		if err != nil {
			s.t.Fatal(err)
		}
		body = bytes.NewReader(buf)
	}
	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", token)
		req.AddCookie(&http.Cookie{Name: "iam_access_token", Value: token})
	}
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)
	return w
}

func seedTestData(t *testing.T, db *gorm.DB) {
	t.Helper()
	ctx := context.Background()
	adminRole := model.Role{Code: "admin", Name: "系统管理员", Status: 1, Remark: "default seeded role"}
	if err := db.WithContext(ctx).Create(&adminRole).Error; err != nil {
		t.Fatal(err)
	}
	adminOpenID := "ou_admin"
	adminUser := model.User{Username: "admin", OpenID: &adminOpenID, DisplayName: "系统管理员", Status: 1}
	if err := db.WithContext(ctx).Create(&adminUser).Error; err != nil {
		t.Fatal(err)
	}
	hash, err := password.Hash("123456", 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Create(&model.AuthIdentity{UserID: adminUser.ID, IdentityType: "password", Identifier: adminUser.Username, Credential: hash}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Model(&adminUser).Association("Roles").Replace([]model.Role{adminRole}); err != nil {
		t.Fatal(err)
	}
	clients := []model.OAuthClient{
		{ClientID: "system-a", ClientSecret: "system-a-secret", Name: "System A", Code: "system-a-oauth2", ResponseType: "code", RedirectURI: "http://system-a.local/callback", Status: 1, Remark: "default seeded client"},
		{ClientID: "system-b", ClientSecret: "system-b-secret", Name: "System B", Code: "system-b-oauth2", ResponseType: "code", RedirectURI: "http://system-b.local/callback", Status: 1, Remark: "default seeded client"},
	}
	for _, client := range clients {
		item := client
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func decodeJSON(t *testing.T, data []byte, out any) {
	t.Helper()
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode json failed: %v, body=%s", err, string(data))
	}
}
