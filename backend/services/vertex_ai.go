package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/vertexai/genai"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"google.golang.org/api/option"
)

// 是否使用模拟模式
var UseMockMode = false

// VertexAIClient 处理与Vertex AI的通信
type VertexAIClient struct {
	projectID string
	location  string
	model     string
	client    *genai.Client
}

// NewVertexAIClient 创建新的Vertex AI客户端
func NewVertexAIClient() *VertexAIClient {
	return &VertexAIClient{
		projectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
		location:  os.Getenv("GOOGLE_CLOUD_LOCATION"),
		model:     "gemini-2.0-flash-001", // 使用支持多模态（图像和PDF）的Gemini模型
	}
}

// 创建带代理设置的 HTTP 客户端选项
func getClientOptions(credentialsFile string) []option.ClientOption {
	// 仅返回凭证文件选项，不再设置 HTTP 客户端
	// gRPC 会自动使用环境变量中的 HTTP_PROXY/HTTPS_PROXY 设置
	return []option.ClientOption{
		option.WithCredentialsFile(credentialsFile),
	}
}

// EnsureCompleteJSON 检查并确保返回的 JSON 是完整的
func EnsureCompleteJSON(content string) string {
	log.Printf("[DEBUG] 开始检查JSON完整性")
	log.Printf("[DEBUG] 原始内容长度: %d 字符", len(content))

	// 清理内容
	content = strings.TrimSpace(content)

	// 检查是否为空
	if content == "" {
		log.Printf("[WARN] 内容为空")
		return "{}"
	}

	// 检查是否以 { 开始
	if !strings.HasPrefix(content, "{") {
		log.Printf("[WARN] 内容不以 { 开始，尝试查找第一个 {")
		if idx := strings.Index(content, "{"); idx >= 0 {
			content = content[idx:]
			log.Printf("[DEBUG] 已找到第一个 {，移除前面的内容")
		} else {
			log.Printf("[WARN] 未找到 {，返回空对象")
			return "{}"
		}
	}

	// 检查是否以 } 结束
	if !strings.HasSuffix(content, "}") {
		log.Printf("[WARN] 内容不以 } 结束，尝试查找最后一个 }")
		if idx := strings.LastIndex(content, "}"); idx >= 0 {
			content = content[:idx+1]
			log.Printf("[DEBUG] 已找到最后一个 }，移除后面的内容")
		} else {
			log.Printf("[WARN] 未找到 }，返回空对象")
			return "{}"
		}
	}

	// 计算大括号的数量
	openCount := strings.Count(content, "{")
	closeCount := strings.Count(content, "}")

	// 如果大括号数量不匹配，尝试修复
	if openCount != closeCount {
		log.Printf("[WARN] 大括号数量不匹配: { = %d, } = %d", openCount, closeCount)

		// 如果缺少右大括号，添加缺少的数量
		if openCount > closeCount {
			missing := openCount - closeCount
			content += strings.Repeat("}", missing)
			log.Printf("[DEBUG] 添加了 %d 个缺少的 }", missing)
		}
		// 如果缺少左大括号，添加到开头
		if closeCount > openCount {
			missing := closeCount - openCount
			content = strings.Repeat("{", missing) + content
			log.Printf("[DEBUG] 添加了 %d 个缺少的 {", missing)
		}
	}

	// 验证JSON格式
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(content), &jsonObj); err != nil {
		log.Printf("[ERROR] JSON格式无效: %v", err)
		// 尝试使用更宽松的方式解析
		content = CleanMarkdownCodeBlock(content)
		if err := json.Unmarshal([]byte(content), &jsonObj); err != nil {
			log.Printf("[ERROR] 清理后JSON仍然无效，返回空对象")
			return "{}"
		}
		log.Printf("[DEBUG] 清理后JSON有效")
	}

	log.Printf("[DEBUG] JSON完整性检查完成，最终内容长度: %d 字符", len(content))
	return content
}

// sanitizeUTF8 清理字符串中的无效UTF-8字符（内部函数）
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	// 创建一个新的字符串构建器
	var builder strings.Builder
	builder.Grow(len(s))

	// 遍历字符串，只保留有效的UTF-8字符
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r != utf8.RuneError || size == 1 {
			builder.WriteRune(r)
		}
		i += size
	}

	return builder.String()
}

// GenerateContent 使用文本内容生成AI回复，不需要附加文件
func (c *VertexAIClient) GenerateContent(systemInstruction, textPrompt string) (string, error) {
	log.Printf("[INFO] 开始生成AI内容...")
	log.Printf("[DEBUG] 系统指令长度: %d 字符", len(systemInstruction))
	log.Printf("[DEBUG] 文本提示词长度: %d 字符", len(textPrompt))

	// 创建一个上下文，可以在必要时取消请求
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	// 创建客户端
	client, err := genai.NewClient(ctx, c.projectID, c.location, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Printf("[ERROR] 创建AI客户端失败: %v", err)
		return "", fmt.Errorf("创建AI客户端失败: %v", err)
	}
	defer client.Close()

	// 选择模型
	model := client.GenerativeModel(c.model)

	// 设置模型参数
	temperature := float32(0.2)
	topP := float32(0.8)
	topK := int32(40)
	maxOutputTokens := int32(8192)
	model.Temperature = &temperature
	model.TopP = &topP
	model.TopK = &topK
	model.MaxOutputTokens = &maxOutputTokens

	// 如果系统指令不为空，设置系统指令
	if systemInstruction != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(systemInstruction)},
			Role:  "system",
		}
	}

	log.Printf("[INFO] 发送请求到Gemini模型，预期等待时间10-30秒...")

	// 发送请求 - 直接传递文本作为输入
	resp, err := model.GenerateContent(ctx, genai.Text(textPrompt))

	// 处理错误
	if err != nil {
		log.Printf("[ERROR] AI请求失败: %v", err)

		// 检查是否为安全策略限制错误
		if strings.Contains(err.Error(), "safety") || strings.Contains(err.Error(), "blocked") {
			return "", fmt.Errorf("内容被安全策略限制，无法处理该请求。请尝试不同的内容或描述方式")
		}

		return "", fmt.Errorf("AI服务请求失败: %v", err)
	}

	// 检查是否有候选结果
	if len(resp.Candidates) == 0 {
		log.Printf("[ERROR] AI未返回任何候选结果")
		return "", fmt.Errorf("AI未返回任何候选结果")
	}

	// 检查是否存在封锁内容原因
	if resp.Candidates[0].FinishReason == genai.FinishReasonSafety {
		log.Printf("[ERROR] 内容被安全策略限制")
		return "", fmt.Errorf("内容被安全策略限制")
	}

	// 检查是否有内容部分
	if len(resp.Candidates[0].Content.Parts) == 0 {
		log.Printf("[ERROR] AI返回的候选结果没有内容部分")
		return "", fmt.Errorf("AI返回的候选结果没有内容部分")
	}

	// 提取响应文本
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	if responseText == "" {
		log.Printf("[ERROR] AI未返回文本内容")
		return "", fmt.Errorf("AI未返回文本内容")
	}

	// 记录响应的预览
	if len(responseText) > 100 {
		log.Printf("[DEBUG] AI响应文本前100个字符: %s", responseText[:100])
	} else {
		log.Printf("[DEBUG] AI响应文本: %s", responseText)
	}

	// 处理JSON格式
	// 如果响应文本看起来是JSON格式，尝试清理和验证
	if strings.Contains(responseText, "{") || strings.Contains(responseText, "[") {
		log.Printf("[INFO] 响应看起来包含JSON，尝试处理和验证")
		// 使用增强版的JSON处理函数
		validJSON := EnsureValidJSON(responseText)
		return validJSON, nil
	}

	// 如果不是JSON，直接返回原始文本
	log.Printf("[INFO] 响应不包含JSON结构，返回原始文本")
	return responseText, nil
}

// GenerateContentWithFile 使用文件内容生成AI回复
func (c *VertexAIClient) GenerateContentWithFile(systemInstruction, filePath, mimeType, textPrompt string) (string, error) {
	log.Printf("[INFO] 开始生成带文件的AI内容...")
	log.Printf("[DEBUG] 系统指令长度: %d 字符", len(systemInstruction))
	log.Printf("[DEBUG] 文件路径: %s", filePath)
	log.Printf("[DEBUG] MIME类型: %s", mimeType)
	log.Printf("[DEBUG] 文本提示词长度: %d 字符", len(textPrompt))

	// 检查模拟模式
	if UseMockMode {
		log.Printf("[INFO] 使用模拟模式，将生成模拟响应")
		return GenerateMockHomeworkResult(filePath, textPrompt)
	}

	// 检查文件是否存在和可访问
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("[ERROR] 文件检查失败: %v", err)
		return "", fmt.Errorf("文件检查失败: %v", err)
	}

	// 获取文件名
	fileName := filepath.Base(filePath)
	log.Printf("[INFO] 文件名: %s, 大小: %d 字节", fileName, fileInfo.Size())

	// 检查API凭证
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credFile == "" {
		log.Printf("[ERROR] 未设置GOOGLE_APPLICATION_CREDENTIALS环境变量")
		return "", fmt.Errorf("未设置GOOGLE_APPLICATION_CREDENTIALS环境变量")
	}

	// 检查凭证文件是否存在
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		log.Printf("[ERROR] API凭证文件不存在: %s", credFile)
		return "", fmt.Errorf("API凭证文件不存在: %s", credFile)
	}

	log.Printf("[INFO] 使用API凭证文件: %s", credFile)

	// 尝试读取文件内容，确保可读
	fileContent, readErr := os.ReadFile(filePath)
	if readErr != nil {
		log.Printf("[ERROR] 无法读取文件内容: %v", readErr)
		return "", fmt.Errorf("无法读取文件内容: %v", readErr)
	}

	log.Printf("[INFO] 成功读取文件内容，大小: %d 字节", len(fileContent))

	// 如果是PDF，检查文件是否有效
	if strings.HasSuffix(fileName, ".pdf") {
		// 尝试获取页数作为验证PDF有效性的方式
		pageCount, err := api.PageCountFile(filePath)
		if err != nil {
			log.Printf("[WARN] PDF验证失败，可能不是有效PDF文件: %v", err)
		} else {
			log.Printf("[INFO] PDF文件有效，页数: %d", pageCount)
			if pageCount > 1 {
				// 如果PDF有多页，修改提示词，指示大模型分析所有页面
				log.Printf("[INFO] 检测到多页PDF，修改提示词以分析所有 %d 页", pageCount)
				textPrompt = fmt.Sprintf("%s 注意：这是一个包含 %d 页的PDF文件，请务必分析所有页面内容。",
					textPrompt, pageCount)
			}
		}
	}

	// 设置重试次数和超时时间
	maxRetries := 3
	for retryCount := 0; retryCount <= maxRetries; retryCount++ {
		if retryCount > 0 {
			backoffTime := time.Duration(retryCount*5) * time.Second
			log.Printf("[INFO] 第 %d 次重试，等待 %v 后进行...", retryCount, backoffTime)
			time.Sleep(backoffTime)
		}

		// 每次尝试创建新的上下文，增加超时时间
		timeoutSeconds := 300 + retryCount*60 // 每次重试多增加60秒的超时时间
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
		defer cancel()

		// 创建客户端
		client, err := genai.NewClient(ctx, c.projectID, c.location, option.WithCredentialsFile(credFile))
		if err != nil {
			log.Printf("[ERROR] 创建AI客户端失败: %v", err)
			continue // 重试
		}
		defer client.Close()

		// 选择模型
		model := client.GenerativeModel(c.model)

		// 设置模型参数
		temperature := float32(0.2)
		topP := float32(0.8)
		topK := int32(40)
		maxOutputTokens := int32(8192)

		model.Temperature = &temperature
		model.TopP = &topP
		model.TopK = &topK
		model.MaxOutputTokens = &maxOutputTokens

		// 如果系统指令不为空，设置系统指令
		if systemInstruction != "" {
			model.SystemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(systemInstruction)},
				Role:  "system",
			}
		}

		// 如果重试，修改提示词
		actualPrompt := textPrompt
		if retryCount > 0 {
			actualPrompt = fmt.Sprintf("重试(%d/%d): %s", retryCount, maxRetries, textPrompt)
		}

		log.Printf("[INFO] 发送带文件的请求到Gemini模型，超时时间: %d秒...", timeoutSeconds)

		// 创建文件blob
		fileBlob := genai.Blob{
			MIMEType: mimeType,
			Data:     fileContent,
		}

		// 发送请求 - 使用多个输入参数
		resp, err := model.GenerateContent(ctx, genai.Text(actualPrompt), fileBlob)

		// 处理错误
		if err != nil {
			log.Printf("[ERROR] AI请求失败 (尝试 %d/%d): %v", retryCount+1, maxRetries+1, err)

			// 详细记录错误类型
			if strings.Contains(err.Error(), "EOF") ||
				strings.Contains(err.Error(), "Unavailable") ||
				strings.Contains(err.Error(), "DeadlineExceeded") ||
				strings.Contains(err.Error(), "timeout") {
				log.Printf("[ERROR] 连接超时或中断，将在稍后重试")
				continue // 继续重试
			}

			// 检查是否为安全策略限制错误
			if strings.Contains(err.Error(), "safety") || strings.Contains(err.Error(), "blocked") {
				return "", fmt.Errorf("内容被安全策略限制，无法处理该请求。请尝试不同的文件或描述方式")
			}

			// 如果是文件解析错误，给出更有用的错误信息
			if strings.Contains(err.Error(), "unsupported") ||
				strings.Contains(err.Error(), "invalid") ||
				strings.Contains(err.Error(), "parse") {
				log.Printf("[WARN] 可能是文件格式问题，尝试使用仅文本方式")

				// 尝试使用纯文本方式重新请求
				retryPrompt := actualPrompt + "\n\n[注意：文件处理失败，但我会尽力分析您的请求]"

				resp, err = model.GenerateContent(ctx, genai.Text(retryPrompt))
				if err != nil {
					if retryCount < maxRetries {
						continue // 继续重试
					}
					return "", fmt.Errorf("文件处理失败，且文本备用方式也失败: %v", err)
				}
			} else {
				if retryCount < maxRetries {
					continue // 继续重试
				}
				return "", fmt.Errorf("AI服务请求失败: %v", err)
			}
		}

		// 检查是否有候选结果
		if len(resp.Candidates) == 0 {
			log.Printf("[ERROR] AI未返回任何候选结果 (尝试 %d/%d)", retryCount+1, maxRetries+1)
			if retryCount < maxRetries {
				continue // 继续重试
			}
			return "", fmt.Errorf("AI未返回任何候选结果")
		}

		// 检查是否存在封锁内容原因
		if resp.Candidates[0].FinishReason == genai.FinishReasonSafety {
			log.Printf("[ERROR] 内容被安全策略限制")
			return "", fmt.Errorf("内容被安全策略限制")
		}

		// 检查是否有内容部分
		if len(resp.Candidates[0].Content.Parts) == 0 {
			log.Printf("[ERROR] AI返回的候选结果没有内容部分 (尝试 %d/%d)", retryCount+1, maxRetries+1)
			if retryCount < maxRetries {
				continue // 继续重试
			}
			return "", fmt.Errorf("AI返回的候选结果没有内容部分")
		}

		// 提取响应文本
		responseText := ""
		for _, part := range resp.Candidates[0].Content.Parts {
			if text, ok := part.(genai.Text); ok {
				responseText += string(text)
			}
		}

		if responseText == "" {
			log.Printf("[ERROR] AI未返回文本内容 (尝试 %d/%d)", retryCount+1, maxRetries+1)
			if retryCount < maxRetries {
				continue // 继续重试
			}
			return "", fmt.Errorf("AI未返回文本内容")
		}

		// 记录响应的预览
		if len(responseText) > 100 {
			log.Printf("[DEBUG] AI响应文本前100个字符: %s", responseText[:100])
		} else {
			log.Printf("[DEBUG] AI响应文本: %s", responseText)
		}

		// 处理JSON格式
		// 如果响应文本看起来是JSON格式，尝试清理和验证
		if strings.Contains(responseText, "{") || strings.Contains(responseText, "[") {
			log.Printf("[INFO] 响应看起来包含JSON，尝试处理和验证")
			// 使用增强版的JSON处理函数
			sanitizedResponse := sanitizeUTF8(responseText)
			validJSON := EnsureValidJSON(sanitizedResponse)
			return validJSON, nil
		}

		// 如果不是JSON，直接返回原始文本
		log.Printf("[INFO] 响应不包含JSON结构，返回原始文本")
		return sanitizeUTF8(responseText), nil
	}

	// 如果所有重试都失败
	return "", fmt.Errorf("多次尝试后AI服务仍未返回有效响应")
}

// 生成模拟的作业批改结果
func GenerateMockHomeworkResult(filePath, textPrompt string) (string, error) {
	fileName := filepath.Base(filePath)
	homeworkType := "未知"

	// 从提示词中提取作业类型
	if strings.Contains(textPrompt, "英语") {
		homeworkType = "英语"
	} else if strings.Contains(textPrompt, "数学") {
		homeworkType = "数学"
	} else if strings.Contains(textPrompt, "语文") {
		homeworkType = "语文"
	}

	log.Printf("[INFO] 生成模拟作业批改结果，文件: %s, 类型: %s", fileName, homeworkType)

	// 生成随机数量的学生结果（1-3名学生）
	numStudents := rand.Intn(3) + 1

	// 创建学生结果数组
	studentResults := make([]string, numStudents)

	// 基础名字列表
	baseNames := []string{"张三", "李四", "王五", "赵六", "钱七"}

	// 为每个学生生成模拟结果
	for i := 0; i < numStudents; i++ {
		// 获取学生姓名
		studentName := baseNames[rand.Intn(len(baseNames))]

		// 根据作业类型生成不同的模拟结果
		var studentResult string

		switch homeworkType {
		case "英语":
			// 生成随机分数 - 确保是0-100之间的整数
			score := 70 + rand.Intn(30)

			// 生成答案数组
			correctCount := 2 + rand.Intn(2)   // 2-3个正确答案
			totalQuestions := 4 + rand.Intn(2) // 4-5个总题目

			answers := []string{}
			for j := 1; j <= totalQuestions; j++ {
				isCorrect := j <= correctCount // 前N题为正确答案

				if isCorrect {
					answer := fmt.Sprintf(`
    {
      "questionNumber": "%d",
      "studentAnswer": "This is a correct answer for question %d",
      "isCorrect": true,
      "explanation": "答案正确，表达流畅"
    }`, j, j)
					answers = append(answers, answer)
				} else {
					answer := fmt.Sprintf(`
    {
      "questionNumber": "%d",
      "studentAnswer": "This is an incorrect answer for question %d",
      "isCorrect": false,
      "correctAnswer": "The correct answer for question %d",
      "explanation": "答案有误，需要改进"
    }`, j, j, j)
					answers = append(answers, answer)
				}
			}

			// 合并成完整的学生结果 - 分数直接使用整数
			studentResult = fmt.Sprintf(`{
  "name": "%s",
  "answers": [%s
  ],
  "overallScore": "%d",
  "feedback": "英语作业完成良好，发音和语法有待提高。继续练习！"
}`, studentName, strings.Join(answers, ","), score)

		case "数学":
			// 生成随机分数 - 确保是0-100之间的整数
			score := 75 + rand.Intn(25)

			// 生成答案数组
			correctCount := 3 + rand.Intn(3)   // 3-5个正确答案
			totalQuestions := 6 + rand.Intn(3) // 6-8个总题目

			answers := []string{}
			for j := 1; j <= totalQuestions; j++ {
				isCorrect := j <= correctCount // 前N题为正确答案

				if isCorrect {
					answer := fmt.Sprintf(`
    {
      "questionNumber": "%d",
      "studentAnswer": "x = %d",
      "isCorrect": true,
      "explanation": "计算正确，方法得当"
    }`, j, 2*j)
					answers = append(answers, answer)
				} else {
					correctValue := 2 * j
					wrongValue := correctValue + 1 + rand.Intn(3)
					answer := fmt.Sprintf(`
    {
      "questionNumber": "%d",
      "studentAnswer": "x = %d",
      "isCorrect": false,
      "correctAnswer": "x = %d",
      "explanation": "计算有误，应为%d"
    }`, j, wrongValue, correctValue, correctValue)
					answers = append(answers, answer)
				}
			}

			// 合并成完整的学生结果 - 分数直接使用整数
			studentResult = fmt.Sprintf(`{
  "name": "%s",
  "answers": [%s
  ],
  "overallScore": "%d",
  "feedback": "数学作业基本掌握了概念，但计算需要更加仔细。多做练习，提高准确性。"
}`, studentName, strings.Join(answers, ","), score)

		case "语文":
			// 生成随机分数 - 确保是0-100之间的整数
			score := 80 + rand.Intn(20)

			// 生成答案数组
			correctCount := 2 + rand.Intn(2)   // 2-3个正确答案
			totalQuestions := 4 + rand.Intn(2) // 4-5个总题目

			answers := []string{}
			for j := 1; j <= totalQuestions; j++ {
				if j <= correctCount {
					answer := fmt.Sprintf(`
    {
      "questionNumber": "%d",
      "studentAnswer": "语文题目%d的正确回答",
      "isCorrect": true,
      "evaluation": "理解深刻，表达流畅",
      "suggestion": "可以再增加一些文学性的表达"
    }`, j, j)
					answers = append(answers, answer)
				} else {
					answer := fmt.Sprintf(`
    {
      "questionNumber": "%d",
      "studentAnswer": "语文题目%d的不完全正确回答",
      "isCorrect": false,
      "correctAnswer": "语文题目%d的标准答案",
      "evaluation": "基本理解了题意，但表达不够准确",
      "suggestion": "需要注意遣词造句，提高表达准确性"
    }`, j, j, j)
					answers = append(answers, answer)
				}
			}

			// 合并成完整的学生结果 - 分数直接使用整数
			studentResult = fmt.Sprintf(`{
  "name": "%s",
  "answers": [%s
  ],
  "overallScore": "%d",
  "feedback": "语文作业整体表现良好，对文章主旨理解到位，但词句表达有待提高。"
}`, studentName, strings.Join(answers, ","), score)

		default:
			// 通用题目
			score := 75 + rand.Intn(25)
			studentResult = fmt.Sprintf(`{
  "name": "%s",
  "answers": [
    {
      "questionNumber": "1",
      "studentAnswer": "这是学生的第一个回答",
      "isCorrect": true,
      "evaluation": "回答基本正确"
    },
    {
      "questionNumber": "2",
      "studentAnswer": "这是学生的第二个回答",
      "isCorrect": false,
      "correctAnswer": "标准答案",
      "evaluation": "有一些小错误需要修正"
    },
    {
      "questionNumber": "3",
      "studentAnswer": "这是学生的第三个回答",
      "isCorrect": true,
      "evaluation": "回答完全正确"
    }
  ],
  "overallScore": "%d",
  "feedback": "学生整体表现良好，需要在细节上多加注意。"
}`, studentName, score)
		}

		studentResults[i] = studentResult
	}

	// 生成整体结果：如果只有一个学生，直接返回；如果有多个学生，放在数组中
	if numStudents == 1 {
		log.Printf("[INFO] 生成单个学生的作业批改结果")
		return studentResults[0], nil
	} else {
		finalResult := fmt.Sprintf("[%s]", strings.Join(studentResults, ",\n"))
		log.Printf("[INFO] 生成了 %d 名学生的作业批改结果", numStudents)
		return finalResult, nil
	}
}

// GenerateContentStream 使用Vertex AI流式生成内容
func (c *VertexAIClient) GenerateContentStream(ctx context.Context, systemInstruction, prompt string) (*genai.GenerateContentResponseIterator, error) {
	// 添加调试日志
	log.Printf("[DEBUG] 准备调用 Vertex AI 流式生成内容")
	log.Printf("[DEBUG] 项目ID: %s, 位置: %s, 模型: %s", c.projectID, c.location, c.model)
	log.Printf("[DEBUG] 凭证文件路径: %s", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	// 记录代理设置
	log.Printf("[DEBUG] HTTP_PROXY: %s", os.Getenv("HTTP_PROXY"))
	log.Printf("[DEBUG] HTTPS_PROXY: %s", os.Getenv("HTTPS_PROXY"))
	log.Printf("[DEBUG] NO_PROXY: %s", os.Getenv("NO_PROXY"))

	// 如果ctx没有超时，创建一个带有超时的新上下文
	_, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		log.Printf("[DEBUG] 已创建60秒超时的上下文")
	}

	// 清理输入提示中的无效UTF-8字符
	sanitizedPrompt := sanitizeUTF8(prompt)

	// 使用环境变量中的凭证文件路径
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// 检查凭证文件是否存在
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		log.Printf("[ERROR] 凭证文件不存在: %s", credentialsFile)
		return nil, fmt.Errorf("凭证文件不存在: %s", credentialsFile)
	}

	log.Printf("[DEBUG] 开始创建 Vertex AI 客户端...")

	// 获取客户端选项（包含代理设置）
	opts := getClientOptions(credentialsFile)

	// 创建客户端
	client, err := genai.NewClient(ctx, c.projectID, c.location, opts...)
	if err != nil {
		log.Printf("[ERROR] 创建AI客户端失败: %v", err)
		return nil, fmt.Errorf("创建AI客户端失败: %v", err)
	}

	// 保存客户端引用以便稍后关闭
	c.client = client
	log.Printf("[DEBUG] Vertex AI 客户端创建成功")

	// 获取模型
	model := client.GenerativeModel(c.model)

	// 设置生成参数 - 更新参数以确保更简洁的回复
	temperature := float32(0.1) // 降低温度，使输出更确定性
	topP := float32(0.7)
	topK := int32(30)
	// 减少最大输出 token 数，避免过长导致截断
	maxOutputTokens := int32(8192)

	// 直接在模型上设置参数
	model.Temperature = &temperature
	model.TopP = &topP
	model.TopK = &topK
	model.MaxOutputTokens = &maxOutputTokens

	// 正确设置SystemInstruction为genai.Content类型
	sysContent := genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
		Role:  "system",
	}
	model.SystemInstruction = &sysContent

	log.Printf("[DEBUG] 开始向 Vertex AI 发送流式请求...")

	// 流式生成内容
	iter := model.GenerateContentStream(ctx, genai.Text(sanitizedPrompt))
	return iter, nil
}

// BuildHomeworkAnalysisPrompt 构建作业分析提示
func BuildHomeworkAnalysisPrompt(homeworkType string, imageContent string) string {
	var prompt string
	switch homeworkType {
	case "english":
		prompt = fmt.Sprintf(`请分析这张英语作业图片，提取所有问题和答案。要求：
1. 按照从左到右、从上到下的顺序识别
2. 保持原始格式和编号
3. 返回JSON格式，包含以下字段：
   - questions: 问题列表
   - answers: 答案列表
   - feedback: 总体评价

图片内容：
%s`, imageContent)
	case "chinese":
		prompt = fmt.Sprintf(`请分析这张语文作业图片，提取所有问题和答案。要求：
1. 按照从左到右、从上到下的顺序识别
2. 保持原始格式和编号
3. 返回JSON格式，包含以下字段：
   - questions: 问题列表
   - answers: 答案列表
   - feedback: 总体评价

图片内容：
%s`, imageContent)
	case "math":
		prompt = fmt.Sprintf(`请分析这张数学作业图片，提取所有问题和答案。要求：
1. 按照从左到右、从上到下的顺序识别
2. 保持原始格式和编号
3. 返回JSON格式，包含以下字段：
   - questions: 问题列表
   - answers: 答案列表
   - feedback: 总体评价

图片内容：
%s`, imageContent)
	default:
		prompt = fmt.Sprintf(`请分析这张作业图片，提取所有问题和答案。要求：
1. 按照从左到右、从上到下的顺序识别
2. 保持原始格式和编号
3. 返回JSON格式，包含以下字段：
   - questions: 问题列表
   - answers: 答案列表
   - feedback: 总体评价

图片内容：
%s`, imageContent)
	}
	return prompt
}

// GenerateContentWithBinaryFile 使用Vertex AI分析二进制文件内容
func (c *VertexAIClient) GenerateContentWithBinaryFile(systemInstruction string, fileContent string, mimeType string, textPrompt string) (string, error) {
	ctx := context.Background()

	// 使用环境变量中的凭证文件路径
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// 创建客户端
	client, err := genai.NewClient(ctx, c.projectID, c.location, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return "", fmt.Errorf("创建AI客户端失败: %v", err)
	}
	defer client.Close()

	c.client = client

	log.Printf("使用模型: %s, 项目: %s, 位置: %s", c.model, c.projectID, c.location)

	// 获取模型
	model := client.GenerativeModel(c.model)

	// 设置生成参数
	temperature := float32(0.2)
	topP := float32(0.8)
	topK := int32(40)
	maxOutputTokens := int32(8192)

	// 直接在模型上设置参数
	model.Temperature = &temperature
	model.TopP = &topP
	model.TopK = &topK
	model.MaxOutputTokens = &maxOutputTokens

	// 正确设置SystemInstruction为genai.Content类型
	sysContent := genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
		Role:  "system",
	}
	model.SystemInstruction = &sysContent

	// 将字符串内容转换为字节数组
	fileData := []byte(fileContent)
	fileSize := len(fileData)
	log.Printf("处理文件内容大小: %d 字节, MIME类型: %s", fileSize, mimeType)

	// 检查文件大小是否超过限制（25MB的安全限制）
	if fileSize > 25*1024*1024 {
		return "", fmt.Errorf("文件过大，超过25MB限制: %d 字节", fileSize)
	}

	// 构建提示文本
	filePrompt := "请分析以下简历文件："
	combinedPrompt := filePrompt
	if textPrompt != "" {
		// 清理提示词中的无效字符
		sanitizedPrompt := strings.TrimSpace(textPrompt)
		// 先检查字符串是否有效的UTF-8
		if !utf8.ValidString(sanitizedPrompt) {
			var builder strings.Builder
			builder.Grow(len(sanitizedPrompt))

			// 遍历字符串，只保留有效的UTF-8字符
			for i := 0; i < len(sanitizedPrompt); {
				r, size := utf8.DecodeRuneInString(sanitizedPrompt[i:])
				if r != utf8.RuneError || size == 1 {
					builder.WriteRune(r)
				}
				i += size
			}
			sanitizedPrompt = builder.String()
		}
		combinedPrompt += "\n\n" + sanitizedPrompt
	}

	log.Printf("准备发送文件内容到Gemini API, 提示词长度: %d 字符", len(combinedPrompt))

	// 尝试使用二进制格式发送请求
	var resp *genai.GenerateContentResponse
	var responseErr error

	// 直接将原始文件数据作为请求的一部分（二进制数据）
	resp, responseErr = model.GenerateContent(ctx, genai.Blob{
		MIMEType: mimeType,
		Data:     fileData,
	}, genai.Text(combinedPrompt))

	if responseErr != nil {
		log.Printf("AI内容生成失败: %v", responseErr)
		// 如果文件格式错误或解析失败，尝试仅使用文本分析
		if strings.Contains(responseErr.Error(), "unsupported") ||
			strings.Contains(responseErr.Error(), "cannot parse") ||
			strings.Contains(responseErr.Error(), "invalid") {

			log.Printf("文件解析失败，尝试使用纯文本方式重新分析")

			// 构建替代提示
			alternativePrompt := fmt.Sprintf(
				"无法解析文件。这可能是格式问题或文件损坏。请根据职位要求生成一个通用的简历评估，说明由于技术原因无法分析此简历。",
			)

			// 尝试纯文本请求
			resp, responseErr = model.GenerateContent(ctx, genai.Text(alternativePrompt))
			if responseErr != nil {
				return "", fmt.Errorf("备用分析也失败: %v", responseErr)
			}
		} else {
			return "", fmt.Errorf("AI内容生成失败: %v", responseErr)
		}
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("AI未返回有效内容")
	}

	// 获取响应文本
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	if responseText == "" {
		return "", fmt.Errorf("AI未返回文本内容")
	}

	log.Printf("成功收到回复，长度: %d 字符", len(responseText))

	// 清理响应中的无效UTF-8字符
	sanitizedResponse := sanitizeUTF8(responseText)

	// 确保 JSON 完整
	sanitizedResponse = EnsureCompleteJSON(sanitizedResponse)

	return sanitizedResponse, nil
}

// UpdatePrompt 更新提示词
func (c *VertexAIClient) UpdatePrompt(prompt string) string {
	// 清理提示词中的无效字符
	sanitizedPrompt := strings.TrimSpace(prompt)
	// 先检查字符串是否有效的UTF-8
	if !utf8.ValidString(sanitizedPrompt) {
		var builder strings.Builder
		builder.Grow(len(sanitizedPrompt))

		// 遍历字符串，只保留有效的UTF-8字符
		for i := 0; i < len(sanitizedPrompt); {
			r, size := utf8.DecodeRuneInString(sanitizedPrompt[i:])
			if r != utf8.RuneError || size == 1 {
				builder.WriteRune(r)
			}
			i += size
		}
		sanitizedPrompt = builder.String()
	}

	return sanitizedPrompt
}
