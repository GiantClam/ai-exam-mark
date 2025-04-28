package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GiantClam/homework_marking/handlers"
	"github.com/GiantClam/homework_marking/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

		// 添加文件服务API
		files := api.Group("/files")
		{
			// 用于获取分割后的PDF文件
			files.GET("/:path/:filename", func(c *gin.Context) {
				path := c.Param("path")
				filename := c.Param("filename")

				// 安全检查，确保文件名不包含路径遍历
				if filepath.Ext(filename) != ".pdf" {
					c.JSON(http.StatusBadRequest, gin.H{
						"status":  "error",
						"message": "只支持PDF文件",
					})
					return
				}

				// 构建文件路径
				filePath := filepath.Join("uploads/split", path, filename)

				// 检查文件是否存在
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					log.Printf("[ERROR] 文件不存在: %s", filePath)

					// 尝试直接从 uploads/split 目录获取文件
					directPath := filepath.Join("uploads/split", filename)
					if _, err := os.Stat(directPath); os.IsNotExist(err) {
						c.JSON(http.StatusNotFound, gin.H{
							"status":  "error",
							"message": "文件不存在",
						})
						return
					}

					// 使用直接路径
					filePath = directPath
					log.Printf("[INFO] 使用直接路径提供文件: %s", filePath)
				}

				// 提供文件
				log.Printf("[INFO] 正在提供文件: %s", filePath)
				c.File(filePath)
			})

			// 增加新路由处理直接从 split 目录获取文件的情况
			files.GET("/split/:path/:filename", func(c *gin.Context) {
				filename := c.Param("filename")

				// 安全检查，确保文件名不包含路径遍历
				if filepath.Ext(filename) != ".pdf" {
					c.JSON(http.StatusBadRequest, gin.H{
						"status":  "error",
						"message": "只支持PDF文件",
					})
					return
				}

				// 构建文件路径
				filePath := filepath.Join("uploads/split", filename)

				// 检查文件是否存在
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					log.Printf("[ERROR] 文件不存在: %s", filePath)
					c.JSON(http.StatusNotFound, gin.H{
						"status":  "error",
						"message": "文件不存在",
					})
					return
				}

				// 提供文件
				log.Printf("[INFO] 正在提供文件: %s", filePath)
				c.File(filePath)
			})
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
