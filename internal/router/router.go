package router

import (
	"iam/internal/handler"
	"iam/internal/middleware"
	"iam/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handlers struct {
	Auth            *handler.AuthHandler
	User            *handler.UserHandler
	Role            *handler.RoleHandler
	OAuth           *handler.OAuthHandler
	AuthApplication *handler.AuthApplicationHandler
}

func New(authService service.AuthService, redisClient *redis.Client, handlers Handlers) *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Auth.Login)
			auth.Use(middleware.Auth(authService, redisClient))
			auth.POST("/logout", handlers.Auth.Logout)
			auth.GET("/me", handlers.Auth.Me)
		}

		oauth := api.Group("/oauth")
		{
			oauth.GET("/authorize", handlers.OAuth.Authorize)
			oauth.POST("/token", handlers.OAuth.Token)
			oauth.GET("/userinfo", handlers.OAuth.UserInfo)
		}

		secured := api.Group("")
		secured.Use(middleware.Auth(authService, redisClient))
		{
			secured.POST("/auth-applications", handlers.AuthApplication.Create)
			secured.GET("/auth-applications", handlers.AuthApplication.List)
			secured.GET("/auth-applications/:id", handlers.AuthApplication.Get)
			secured.PUT("/auth-applications/:id", handlers.AuthApplication.Update)
			secured.DELETE("/auth-applications/:id", handlers.AuthApplication.Delete)
			secured.POST("/users", handlers.User.Create)
			secured.GET("/users", handlers.User.List)
			secured.GET("/users/:id", handlers.User.Get)
			secured.PUT("/users/:id/status", handlers.User.UpdateStatus)
			secured.DELETE("/users/:id", handlers.User.Delete)
			secured.PUT("/users/:id/password", handlers.User.ResetPassword)
			secured.POST("/roles", handlers.Role.Create)
			secured.GET("/roles", handlers.Role.List)
			secured.PUT("/users/:id/roles", handlers.Role.BindUserRoles)
			secured.GET("/users/:id/roles", handlers.Role.GetUserRoles)
		}
	}
	return r
}
