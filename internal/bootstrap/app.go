package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"iam/internal/handler"
	"iam/internal/model"
	jwtpkg "iam/internal/pkg/jwt"
	"iam/internal/pkg/password"
	"iam/internal/repository"
	"iam/internal/router"
	"iam/internal/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	Config *Config
	Logger *slog.Logger
	DB     *gorm.DB
	Redis  *redis.Client
	Server *http.Server
}

func NewApp() (*App, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	logger := NewLogger(cfg)
	db, err := NewDB(cfg)
	if err != nil {
		return nil, err
	}
	redisClient := NewRedis(cfg)

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	// if err := autoMigrate(db); err != nil {
	// 	return nil, err
	// }
	if err := seedData(context.Background(), db, cfg); err != nil {
		return nil, err
	}

	jwtManager := jwtpkg.NewManager(cfg.JWT.Issuer, cfg.JWT.Secret, cfg.JWT.ExpireSeconds)
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	identityRepo := repository.NewAuthIdentityRepository(db)
	clientRepo := repository.NewOAuthClientRepository(db)

	authService := service.NewAuthService(userRepo, identityRepo, redisClient, jwtManager, cfg.Security.LoginFailLimit, cfg.Security.LoginFailWindowSeconds)
	userService := service.NewUserService(userRepo, roleRepo, identityRepo, cfg.Security.PasswordCost)
	roleService := service.NewRoleService(roleRepo, userRepo)
	oauthService := service.NewOAuthService(clientRepo, authService, userRepo, redisClient, jwtManager, cfg.OAuth.AuthorizeCodeExpireSeconds)
	authApplicationService := service.NewAuthApplicationService(clientRepo)

	engine := router.New(authService, redisClient, router.Handlers{
		Auth:            handler.NewAuthHandler(authService),
		User:            handler.NewUserHandler(userService),
		Role:            handler.NewRoleHandler(roleService),
		OAuth:           handler.NewOAuthHandler(oauthService, authService, redisClient, cfg.OAuth.LoginURL),
		AuthApplication: handler.NewAuthApplicationHandler(authApplicationService),
	})

	server := &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port), Handler: engine, ReadHeaderTimeout: 5 * time.Second}

	return &App{Config: cfg, Logger: logger, DB: db, Redis: redisClient, Server: server}, nil
}

func (a *App) Run() error {
	a.Logger.Info("server started", "addr", a.Server.Addr)
	err := a.Server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.Server.Shutdown(ctx); err != nil {
		return err
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
	return nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{}, &model.Role{}, &model.UserRole{}, &model.AuthIdentity{}, &model.OAuthClient{})
}

func seedData(ctx context.Context, db *gorm.DB, cfg *Config) error {
	// return seedAdmin(ctx, db, cfg)
	return nil
}

func seedAdmin(ctx context.Context, db *gorm.DB, cfg *Config) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	adminRole := model.Role{Code: "admin", Name: "系统管理员", Status: 1, Remark: "default seeded role"}
	if err := db.WithContext(ctx).Where(model.Role{Code: adminRole.Code}).FirstOrCreate(&adminRole).Error; err != nil {
		return err
	}

	admin := model.User{Username: "admin", DisplayName: "系统管理员", Status: 1}
	if err := db.WithContext(ctx).Create(&admin).Error; err != nil {
		return err
	}
	hash, err := password.Hash("123456", cfg.Security.PasswordCost)
	if err != nil {
		return err
	}
	if err := db.WithContext(ctx).Create(&model.AuthIdentity{UserID: admin.ID, IdentityType: "password", Identifier: admin.Username, Credential: hash}).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&admin).Association("Roles").Replace([]model.Role{adminRole})
}
