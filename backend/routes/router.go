package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/GiantClam/homework_marking/handlers"
	"github.com/GiantClam/homework_marking/services"
)

// SetupRouter 设置API路由
func SetupRouter(geminiService *services.GeminiService) *gin.Engine {
	// 创建任务队列
	taskQueue := services.NewTaskQueue(5) // 5个工作协程

	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // 当AllowOrigins为*时，必须设为false
		MaxAge:           12 * time.Hour,
	}))

	// 添加请求日志中间件
	r.Use(func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqURI := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		log.Printf("[GIN] %s | %3d | %13v | %15s | %s",
			reqMethod,
			statusCode,
			latencyTime,
			c.ClientIP(),
			reqURI,
		)
	})

	// 创建处理器
	homeworkHandler := handlers.NewHomeworkHandler(taskQueue)
	taskHandler := handlers.NewTaskHandler(taskQueue)

	// 上传文件API
	api := r.Group("/api")
	{
		upload := api.Group("/upload")
		{
			upload.POST("/homework", homeworkHandler.UploadHomework)
		}

		// 作业批改API
		marking := api.Group("/marking")
		{
			marking.POST("/homework", homeworkHandler.MarkHomework)
		}
		
		// 任务API
		tasks := api.Group("/tasks")
		{
			tasks.GET("/:taskId", taskHandler.GetTaskStatus)
			tasks.GET("", taskHandler.GetAllTasks)
		}

		// 测试API
		api.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "API工作正常",
				"time":    time.Now(),
			})
		})
	}

	// 添加根路径处理
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "online",
			"message": "AI作业批改API",
			"version": "1.0",
		})
	})

	// 添加健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 无法找到路由时的处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("未找到路由: %s %s", c.Request.Method, c.Request.URL.Path),
		})
	})

	return r
}
