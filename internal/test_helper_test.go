package internal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type apiLoginData struct {
	AccessToken string          `json:"access_token"`
	ExpiresIn   int64           `json:"expires_in"`
	User        dto.CurrentUser `json:"user"`
}

type apiSuite struct {
	t         *testing.T
	engine    *gin.Engine
	redis     *redis.Client
	miniRedis *miniredis.Miniredis
	db        *gorm.DB
}

func newAPISuite(t *testing.T) *apiSuite {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	dsn := "file:" + strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Role{}, &model.UserRole{}, &model.AuthIdentity{}, &model.OAuthClient{}); err != nil {
		t.Fatal(err)
	}
	seedAPIData(t, db)

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

	return &apiSuite{t: t, engine: engine, redis: redisClient, miniRedis: mr, db: db}
}

func (s *apiSuite) close() {
	_ = s.redis.Close()
	s.miniRedis.Close()
	if sqlDB, err := s.db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func (s *apiSuite) doJSON(method string, path string, payload any, token string) *httptest.ResponseRecorder {
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

func (s *apiSuite) loginAsAdmin() apiLoginData {
	s.t.Helper()
	resp := s.doJSON(http.MethodPost, "/api/iam/auth/login", map[string]any{"username": "admin", "password": "123456"}, "")
	if resp.Code != http.StatusOK {
		s.t.Fatalf("admin login failed: status=%d body=%s", resp.Code, resp.Body.String())
	}
	var envelope apiEnvelope
	decodeAPIJSON(s.t, resp.Body.Bytes(), &envelope)
	var data apiLoginData
	decodeAPIJSON(s.t, envelope.Data, &data)
	return data
}

func seedAPIData(t *testing.T, db *gorm.DB) {
	t.Helper()
	ctx := context.Background()
	adminRole := model.Role{Code: "admin", Name: "系统管理员", Status: 1, Remark: "default seeded role"}
	if err := db.WithContext(ctx).Create(&adminRole).Error; err != nil {
		t.Fatal(err)
	}
	adminOpenID := "ou_admin"
	adminEmail := "admin@example.com"
	adminMobile := "13900000000"
	adminUser := model.User{Username: "admin", OpenID: &adminOpenID, DisplayName: "系统管理员", Email: &adminEmail, Mobile: &adminMobile, Status: 1, Remark: "负责 IAM 平台账号、角色和接入应用的日常管理"}
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

func decodeAPIJSON(t *testing.T, data []byte, out any) {
	t.Helper()
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode json failed: %v body=%s", err, string(data))
	}
}
