package main

import (
	"fmt"
	"log"
	"os"
	"unicode/utf16"
	"bytes"
	"io/ioutil"

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

// 检测文件是否为UTF-16编码，并转换为UTF-8
func convertUTF16ToUTF8(filename string) (string, error) {
	// 读取文件内容
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// 检查是否包含BOM(字节顺序标记)
	if len(content) >= 2 {
		// UTF-16LE BOM: 0xFF 0xFE
		if content[0] == 0xFF && content[1] == 0xFE {
			log.Printf("检测到UTF-16LE编码的文件: %s", filename)
			
			// 创建临时文件名
			tmpFilename := filename + ".utf8"
			
			// 移除BOM，并确保字节数是偶数
			content = content[2:]
			if len(content)%2 != 0 {
				content = content[:len(content)-1]
			}
			
			// 将UTF-16LE转换为UTF-8
			u16s := make([]uint16, len(content)/2)
			for i := 0; i < len(u16s); i++ {
				u16s[i] = uint16(content[i*2]) + (uint16(content[i*2+1]) << 8)
			}
			
			// 转换回UTF-8
			var buf bytes.Buffer
			for _, r := range utf16.Decode(u16s) {
				buf.WriteRune(r)
			}
			
			// 写入转换后的文件
			if err := ioutil.WriteFile(tmpFilename, buf.Bytes(), 0644); err != nil {
				return "", fmt.Errorf("无法写入转换后的UTF-8文件: %v", err)
			}
			
			log.Printf("已将文件转换为UTF-8: %s", tmpFilename)
			return tmpFilename, nil
		}
	}
	
	// 不是UTF-16，返回原始文件名
	return filename, nil
}

// 加载环境变量文件
func loadEnvFile(filename string) error {
	// 首先尝试转换文件编码
	utf8Filename, err := convertUTF16ToUTF8(filename)
	if err != nil {
		return fmt.Errorf("转换文件编码失败: %v", err)
	}
	
	// 使用转换后的文件加载环境变量
	return godotenv.Load(utf8Filename)
}

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("启动作业批改服务...")

	// 首先同步代理环境变量
	syncProxyEnvVars()

	// 检查并优先加载.env文件中的环境变量
	log.Println("加载环境变量配置...")
	
	// 定义可能的环境文件列表，按优先级排序
	envFiles := []string{".env", ".env.production", ".env.example"}
	
	var loadedFile string
	var err error
	
	// 检查文件是否存在
	for _, file := range envFiles {
		if _, statErr := os.Stat(file); statErr == nil {
			// 文件存在，尝试加载
			if loadErr := loadEnvFile(file); loadErr == nil {
				loadedFile = file
				log.Printf("成功加载环境变量文件: %s", file)
				break
			} else {
				err = loadErr
				log.Printf("警告: 无法加载环境变量文件 %s: %v", file, loadErr)
			}
		}
	}
	
	// 如果没有成功加载任何文件
	if loadedFile == "" {
		log.Println("警告: 未找到有效的环境变量文件(.env, .env.production, .env.example)或无法加载，使用系统环境变量")
	}

	// 记录是否启用模拟模式 (加载.env后检查)
	services.UseMockMode = os.Getenv("USE_MOCK_MODE") == "true"
	log.Printf("模拟模式: %v", services.UseMockMode)

	// 检查关键环境变量
	log.Printf("环境变量配置:")
	log.Printf("- GOOGLE_APPLICATION_CREDENTIALS: %s", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	log.Printf("- GOOGLE_CLOUD_PROJECT: %s", os.Getenv("GOOGLE_CLOUD_PROJECT"))
	log.Printf("- GOOGLE_CLOUD_LOCATION: %s", os.Getenv("GOOGLE_CLOUD_LOCATION"))
	log.Printf("- PORT: %s", os.Getenv("PORT"))

	// 检查凭证文件是否存在
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credFile != "" {
		if _, err := os.Stat(credFile); os.IsNotExist(err) {
			log.Printf("警告: 凭证文件不存在: %s", credFile)
		} else {
			log.Printf("凭证文件存在: %s", credFile)
		}
	} else {
		log.Printf("警告: 未设置GOOGLE_APPLICATION_CREDENTIALS环境变量")
	}

	// 如果未设置项目ID或location，设置默认值
	if os.Getenv("GOOGLE_CLOUD_PROJECT") == "" {
		log.Printf("警告: 未设置GOOGLE_CLOUD_PROJECT环境变量，某些功能可能不可用")
		// 可能需要设置默认值
	}

	if os.Getenv("GOOGLE_CLOUD_LOCATION") == "" {
		log.Printf("警告: 未设置GOOGLE_CLOUD_LOCATION环境变量，某些功能可能不可用")
		// 可能需要设置默认值
	}

	// 确保上传目录存在
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}
	log.Println("[INFO] 上传目录已就绪: uploads")
	
	// 确保分割文件目录存在
	splitDir := "uploads/split"
	if err := os.MkdirAll(splitDir, 0755); err != nil {
		log.Fatalf("创建分割文件目录失败: %v", err)
	}
	log.Println("[INFO] 分割文件目录已就绪: uploads\\split")

	// 设置Gin模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化Gemini服务
	log.Println("初始化Gemini服务...")
	geminiService, err := services.NewGeminiService()
	if err != nil {
		log.Fatalf("创建Gemini服务失败: %v", err)
	}
	log.Println("Gemini服务初始化成功")

	// 使用路由模块配置路由
	r := routes.SetupRouter(geminiService)

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
