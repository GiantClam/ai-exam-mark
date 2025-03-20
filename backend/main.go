package main

import (
	"fmt"
	"log"
	"os"

	"github.com/GiantClam/homework_marking/routes"
	"github.com/GiantClam/homework_marking/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// 同步代理环境变量（确保大小写变量值一致）
func syncProxyEnvVars() {
	// 检查大写环境变量是否已设置
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")
	noProxy := os.Getenv("NO_PROXY")

	// 检查小写环境变量是否已设置
	httpProxyLower := os.Getenv("http_proxy")
	httpsProxyLower := os.Getenv("https_proxy")
	noProxyLower := os.Getenv("no_proxy")

	// 根据优先级设置环境变量
	// 优先使用大写变量值
	if httpProxy != "" && httpProxyLower == "" {
		os.Setenv("http_proxy", httpProxy)
		log.Printf("已设置 http_proxy = %s", httpProxy)
	} else if httpProxyLower != "" && httpProxy == "" {
		os.Setenv("HTTP_PROXY", httpProxyLower)
		log.Printf("已设置 HTTP_PROXY = %s", httpProxyLower)
	}

	if httpsProxy != "" && httpsProxyLower == "" {
		os.Setenv("https_proxy", httpsProxy)
		log.Printf("已设置 https_proxy = %s", httpsProxy)
	} else if httpsProxyLower != "" && httpsProxy == "" {
		os.Setenv("HTTPS_PROXY", httpsProxyLower)
		log.Printf("已设置 HTTPS_PROXY = %s", httpsProxyLower)
	}

	if noProxy != "" && noProxyLower == "" {
		os.Setenv("no_proxy", noProxy)
		log.Printf("已设置 no_proxy = %s", noProxy)
	} else if noProxyLower != "" && noProxy == "" {
		os.Setenv("NO_PROXY", noProxyLower)
		log.Printf("已设置 NO_PROXY = %s", noProxyLower)
	}
}

func main() {
	// 同步代理环境变量
	syncProxyEnvVars()

	// 加载.env文件中的环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("警告: 未找到.env文件或无法加载")
	}

	// 检查凭证文件是否存在，如果不存在则启用模拟模式
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsFile == "" {
		log.Println("警告: 未设置GOOGLE_APPLICATION_CREDENTIALS环境变量，将启用模拟模式")
		services.UseMockMode = true
	} else if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		log.Printf("警告: 凭证文件不存在: %s，将启用模拟模式", credentialsFile)
		services.UseMockMode = true
	} else {
		log.Printf("已找到凭证文件: %s", credentialsFile)
	}

	// 确保上传目录存在
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// 设置Gin模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 使用路由模块配置路由
	r := routes.SetupRouter()

	// 确定端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("作业批改服务器启动在 http://localhost%s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
