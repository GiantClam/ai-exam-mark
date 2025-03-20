package routes

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GiantClam/homework_marking/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置API路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS配置 - 使用cors中间件允许所有源访问
	config := cors.DefaultConfig()

	// 允许特定的源或所有源
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin != "" {
		config.AllowOrigins = []string{corsOrigin}
	} else {
		// 开发环境下允许常见的本地开发源
		config.AllowOrigins = []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:8000",
		}
	}

	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
	}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	r.Use(func(c *gin.Context) {
		// 请求开始前记录
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		method := c.Request.Method
		log.Printf("收到请求[%s]: 路径=%s, IP=%s, 头=%v", method, path, clientIP, c.Request.Header)

		// 执行下一个处理器
		c.Next()

		// 响应后记录
		status := c.Writer.Status()
		log.Printf("请求完成: 路径=%s, 状态=%d, 错误=%v", path, status, c.Errors)
	})

	// 作业批改API
	r.POST("/api/homework/mark", handlers.MarkHomework)

	// 测试API端点
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "测试API连接成功",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 根路径处理 - 返回API状态信息
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "作业批改助手API",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	return r
}
