package router

import (
	"context"
	"github.com/gin-gonic/gin"
	"mall-server/internal/app/models"
	"mall-server/internal/app/service"
	"mall-server/pkg/logger"
)

func App(ctx context.Context, svc *models.ServiceContext) *gin.Engine {
	r := gin.Default()

	// 使用默认配置的CORS中间件
	// 默认允许所有来源（AllowAllOrigins: true）
	//	r.Use(cors.Default())

	r.Use(CORSMiddleware())

	//curl -X POST -H "Content-Type: application/json" -d '{"username":"kane","password":"111"}' http://localhost:8080/user/save

	r.POST("/user/save", func(c *gin.Context) {
		result := service.SaveUser(ctx, c, svc)
		c.JSON(200, gin.H{
			"message": result,
		})
	})

	// 用户登录接口
	// curl -X POST -H "Content-Type: application/json" -d '{"username":"kane","password":"111"}' http://localhost:8080/user/login
	r.POST("/user/login", func(c *gin.Context) {
		service.LoginPsw(ctx, c, svc)
	})

	r.GET("/ping", func(c *gin.Context) {
		logger.WithContext(c).Info("ping invoke111")
		logger.Infof("ping invoke2222")
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	r.GET("/actuator/health/readiness", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, readiness!",
		})
	})

	r.GET("/actuator/health/liveness", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, liveness!",
		})
	})

	return r
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// http://localhost:5173  是前端vue项目的node启动的地址和端口
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
