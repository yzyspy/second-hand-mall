package router

import (
	"mall-server/pkg/jwtx"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 鉴权中间件
// 检测请求头中的 Authorization 字段是否包含合法的 token
// 格式: Authorization: Bearer <token>
// 如果没有 token 或 token 不合法，返回 401 错误
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "未登录，请先登录",
			})
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "token 格式错误，格式为: Bearer <token>",
			})
			return
		}

		// 解析并验证 token
		claims, err := jwtx.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "token 无效或已过期",
			})
			return
		}

		// 将用户信息存入上下文，后续 handler 可通过 c.Get 获取
		c.Set("user_id", claims.UserID)
		c.Set("user_name", claims.UserName)

		c.Next()
	}
}
