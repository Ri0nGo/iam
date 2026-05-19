package middleware

import (
	"strings"

	jwtpkg "iam/internal/pkg/jwt"
	"iam/internal/pkg/resp"
	"iam/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Auth(authService service.AuthService, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if token == "" {
			cookie, err := c.Cookie("iam_access_token")
			if err == nil {
				token = cookie
			}
		}

		if token == "" {
			resp.Fail(c, 401, "missing access token")
			c.Abort()
			return
		}

		claims, err := authService.ParseToken(token)
		if err != nil {
			resp.Fail(c, 401, "invalid token")
			c.Abort()
			return
		}

		if ok, _ := redisClient.Exists(c.Request.Context(), "iam:token:blacklist:"+claims.ID).Result(); ok > 0 {
			resp.Fail(c, 401, "token already revoked")
			c.Abort()
			return
		}
		if claims.TokenUse != jwtpkg.TokenUseConsole {
			resp.Fail(c, 401, "invalid token use")
			c.Abort()
			return
		}

		userID, err := service.ParseUserID(claims.Subject)
		if err != nil {
			resp.Fail(c, 401, "invalid token subject")
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("tokenID", claims.ID)
		c.Next()
	}
}
