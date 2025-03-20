package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/vertexai/genai"
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
		model:     "gemini-2.0-flash-001", // 使用Gemini模型
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
func EnsureCompleteJSON(jsonStr string) string {
	// 检查是否是 Markdown 代码块格式，如果是，先调用 CleanMarkdownCodeBlock 清理
	if strings.Contains(jsonStr, "```") {
		log.Printf("[DEBUG] 检测到 Markdown 代码块格式，进行清理")
		jsonStr = CleanMarkdownCodeBlock(jsonStr)
	}

	// 检查是否以大括号开始和结束
	jsonStr = strings.TrimSpace(jsonStr)
	if !strings.HasPrefix(jsonStr, "{") {
		log.Printf("[WARN] JSON 响应不是以 '{' 开始的: %s", jsonStr[:Min(50, len(jsonStr))])

		// 尝试查找 JSON 开始的位置
		start := strings.Index(jsonStr, "{")
		if start >= 0 {
			log.Printf("[INFO] 找到 JSON 开始位置，截取从 '{' 开始的部分")
			jsonStr = jsonStr[start:]
		} else {
			return "{}"
		}
	}

	// 检查开闭括号是否成对
	openCount := strings.Count(jsonStr, "{")
	closeCount := strings.Count(jsonStr, "}")

	if openCount != closeCount {
		log.Printf("[WARN] JSON 响应括号不配对: %d 个 '{', %d 个 '}'", openCount, closeCount)

		// 尝试修复 JSON
		if openCount > closeCount {
			// 缺少闭括号，添加缺少的闭括号
			jsonStr += strings.Repeat("}", openCount-closeCount)
			log.Printf("[INFO] 添加了 %d 个 '}' 来修复 JSON", openCount-closeCount)
		}
	}

	// 验证 JSON 是否有效
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		log.Printf("[ERROR] JSON 无效: %v", err)

		// 进一步修复 JSON 格式
		jsonStr = TryFixJsonFormat(jsonStr)

		// 再次验证
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			log.Printf("[ERROR] 修复后的 JSON 仍然无效: %v", err)
			return "{}"
		}
	}

	return jsonStr
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

// GenerateContent 使用Vertex AI生成内容
func (c *VertexAIClient) GenerateContent(systemInstruction, prompt string) (string, error) {
	// 添加日志
	log.Printf("[DEBUG] 准备调用 Vertex AI 生成内容")
	log.Printf("[DEBUG] 项目ID: %s, 位置: %s, 模型: %s", c.projectID, c.location, c.model)
	log.Printf("[DEBUG] 凭证文件路径: %s", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	// 记录代理设置
	log.Printf("[DEBUG] HTTP_PROXY: %s", os.Getenv("HTTP_PROXY"))
	log.Printf("[DEBUG] HTTPS_PROXY: %s", os.Getenv("HTTPS_PROXY"))
	log.Printf("[DEBUG] NO_PROXY: %s", os.Getenv("NO_PROXY"))

	// 创建带有较长超时时间的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 清理输入提示中的无效UTF-8字符
	sanitizedPrompt := sanitizeUTF8(prompt)

	// 使用环境变量中的凭证文件路径
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// 检查凭证文件是否存在
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		log.Printf("[ERROR] 凭证文件不存在: %s", credentialsFile)
		return "", fmt.Errorf("凭证文件不存在: %s", credentialsFile)
	}

	log.Printf("[DEBUG] 开始创建 Vertex AI 客户端...")

	// 获取客户端选项（包含代理设置）
	opts := getClientOptions(credentialsFile)

	// 创建客户端
	client, err := genai.NewClient(ctx, c.projectID, c.location, opts...)
	if err != nil {
		log.Printf("[ERROR] 创建AI客户端失败: %v", err)
		return "", fmt.Errorf("创建AI客户端失败: %v", err)
	}
	defer client.Close()

	c.client = client
	log.Printf("[DEBUG] Vertex AI 客户端创建成功")

	// 获取模型
	model := client.GenerativeModel(c.model)

	// 设置生成参数 - 更新参数以确保更简洁的回复
	temperature := float32(0.1) // 降低温度，使输出更确定性
	topP := float32(0.7)
	topK := int32(30)
	// 减少最大输出 token 数，避免过长导致截断
	maxOutputTokens := int32(4096)

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

	log.Printf("[DEBUG] 开始向 Vertex AI 发送请求...")

	// 创建内容
	resp, err := model.GenerateContent(ctx, genai.Text(sanitizedPrompt))
	if err != nil {
		log.Printf("[ERROR] AI内容生成失败: %v，错误类型: %T", err, err)
		return "", fmt.Errorf("AI内容生成失败: %v", err)
	}

	log.Printf("[DEBUG] Vertex AI 响应接收成功")

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

	// 清理响应中的无效UTF-8字符
	sanitizedResponse := sanitizeUTF8(responseText)

	// 确保 JSON 完整
	sanitizedResponse = EnsureCompleteJSON(sanitizedResponse)

	return sanitizedResponse, nil
}

// GenerateContentWithFile 使用Vertex AI分析文件内容
func (c *VertexAIClient) GenerateContentWithFile(systemInstruction string, filePath string, mimeType string, textPrompt string) (string, error) {
	// 如果启用了模拟模式，返回模拟数据
	if UseMockMode {
		log.Printf("[INFO] 使用模拟模式返回作业批改结果")
		mockResult, err := generateMockHomeworkResult(filePath, textPrompt)
		if err != nil {
			return "", err
		}
		// 确保模拟结果也是有效的JSON
		cleanResult := CleanMarkdownCodeBlock(mockResult)
		sanitizedResult := sanitizeUTF8(cleanResult)
		return sanitizedResult, nil
	}

	ctx := context.Background()

	// 使用环境变量中的凭证文件路径
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// 检查凭证文件是否存在
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		log.Printf("[WARN] 凭证文件不存在或无法访问: %s，启用模拟模式", credentialsFile)
		UseMockMode = true
		return generateMockHomeworkResult(filePath, textPrompt)
	}

	// 创建客户端
	client, err := genai.NewClient(ctx, c.projectID, c.location, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Printf("[WARN] 创建AI客户端失败: %v，启用模拟模式", err)
		UseMockMode = true
		return generateMockHomeworkResult(filePath, textPrompt)
	}
	defer client.Close()

	c.client = client

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

	// 读取文件内容
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[WARN] 无法打开文件: %v，启用模拟模式", err)
		UseMockMode = true
		return generateMockHomeworkResult(filePath, textPrompt)
	}
	defer file.Close()

	// 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[WARN] 无法读取文件数据: %v，启用模拟模式", err)
		UseMockMode = true
		return generateMockHomeworkResult(filePath, textPrompt)
	}

	// 获取文件名
	fileName := filepath.Base(filePath)
	log.Printf("处理文件: %s, 大小: %d 字节, MIME类型: %s", fileName, len(fileData), mimeType)

	// 构建提示文本，包含有关文件的信息
	filePrompt := fmt.Sprintf("请分析以下作业图片（文件名: %s）：", fileName)
	combinedPrompt := filePrompt
	if textPrompt != "" {
		sanitizedPrompt := sanitizeUTF8(textPrompt)
		combinedPrompt += "\n\n" + sanitizedPrompt
	}

	// 直接将原始文件数据作为请求的一部分（二进制数据）
	resp, err := model.GenerateContent(ctx, genai.Blob{
		MIMEType: mimeType,
		Data:     fileData,
	}, genai.Text(combinedPrompt))

	if err != nil {
		log.Printf("[WARN] AI内容生成失败: %v，启用模拟模式", err)
		UseMockMode = true
		return generateMockHomeworkResult(filePath, textPrompt)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Printf("[WARN] AI未返回有效内容，启用模拟模式")
		UseMockMode = true
		return generateMockHomeworkResult(filePath, textPrompt)
	}

	// 获取响应文本
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	if responseText == "" {
		log.Printf("[WARN] AI未返回文本内容，启用模拟模式")
		UseMockMode = true
		return generateMockHomeworkResult(filePath, textPrompt)
	}

	// 清理响应中的无效UTF-8字符
	sanitizedResponse := sanitizeUTF8(responseText)

	return sanitizedResponse, nil
}

// 生成模拟的作业批改结果
func generateMockHomeworkResult(filePath, textPrompt string) (string, error) {
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

	// 根据作业类型生成不同的模拟结果
	var mockResult string
	switch homeworkType {
	case "英语":
		mockResult = `{
  "answers": [
    {
      "questionNumber": "1",
      "studentAnswer": "He is a doctor",
      "isCorrect": true,
      "explanation": "正确回答了问题，语法正确"
    },
    {
      "questionNumber": "2",
      "studentAnswer": "They are student",
      "isCorrect": false,
      "correctAnswer": "They are students",
      "explanation": "名词复数形式错误，应该是students"
    },
    {
      "questionNumber": "3",
      "studentAnswer": "I am 12 years old",
      "isCorrect": true,
      "explanation": "年龄表达正确"
    }
  ],
  "overallScore": "85分",
  "feedback": "大部分题目回答正确，但需要注意名词的单复数形式。继续加油！"
}`
	case "数学":
		mockResult = `{
  "answers": [
    {
      "questionNumber": "1",
      "studentAnswer": "x = 5",
      "isCorrect": true,
      "explanation": "解方程步骤正确，得到了正确答案"
    },
    {
      "questionNumber": "2",
      "studentAnswer": "30平方厘米",
      "isCorrect": false,
      "correctAnswer": "28平方厘米",
      "correctSteps": "长4厘米，宽7厘米，面积=4×7=28平方厘米",
      "explanation": "计算错误，4×7=28，不是30"
    },
    {
      "questionNumber": "3",
      "studentAnswer": "64",
      "isCorrect": true,
      "explanation": "计算8的平方正确"
    }
  ],
  "overallScore": "80分",
  "feedback": "基本掌握了数学概念，但计算时需要更加仔细，避免计算错误。"
}`
	case "语文":
		mockResult = `{
  "answers": [
    {
      "questionNumber": "1",
      "studentAnswer": "春风又绿江南岸，明月何时照我还",
      "evaluation": "背诵正确，没有错别字",
      "suggestion": "可以理解诗句中流露的思乡之情"
    },
    {
      "questionNumber": "2",
      "studentAnswer": "欲穷千里目，更上一层天",
      "evaluation": "有错别字，'天'应为'楼'",
      "suggestion": "注意背诵准确性，理解'更上一层楼'意境"
    },
    {
      "questionNumber": "3",
      "studentAnswer": "这篇文章主要讲述了作者童年的回忆",
      "evaluation": "理解文章主题正确，但表达不够具体",
      "suggestion": "可以加入具体事例，指出文章中的关键情节"
    }
  ],
  "overallScore": "88分",
  "feedback": "整体表现不错，对文学作品有一定理解，但背诵古诗词需要更加准确，作答时可以再具体一些。"
}`
	default:
		mockResult = `{
  "answers": [
    {
      "questionNumber": "1",
      "studentAnswer": "这是学生的第一个回答",
      "evaluation": "回答基本正确"
    },
    {
      "questionNumber": "2",
      "studentAnswer": "这是学生的第二个回答",
      "evaluation": "有一些小错误需要修正"
    },
    {
      "questionNumber": "3",
      "studentAnswer": "这是学生的第三个回答",
      "evaluation": "回答完全正确"
    }
  ],
  "feedback": "学生整体表现良好，需要在细节上多加注意。"
}`
	}

	log.Printf("[INFO] 生成模拟作业批改结果，文件: %s, 类型: %s", fileName, homeworkType)
	return mockResult, nil
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
	maxOutputTokens := int32(4096)

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

// BuildResumeScreeningPrompt 构建简历筛选提示
func BuildResumeScreeningPrompt(jobRequirements, industry string, resumeContents []string, resumeNames []string) (string, string) {
	systemInstruction := fmt.Sprintf(`
		你是一个%s行业的高级招聘专家，精通人才筛选。请基于以下招聘要求和行业，评估简历。

		招聘要求:
		%s

		请分析以下简历，判断它是否符合招聘要求，并简要说明你的判断理由:

		请注意以下要求：
		1. 筛选理由必须简洁，每个理由控制在50字以内
		2. 使用要点式回答，便于快速阅读
		3. 确保生成完整的JSON且不会因长度过长而被截断

		请以下面的JSON格式回复:
		{
		"passed": [
			{"name": "简历文件名", "reason": "通过原因(50字以内)"}
		],
		"failed": [
			{"name": "简历文件名", "reason": "不通过原因(50字以内)"}
		]
		}

		注意：每份简历只能出现在passed或failed其中一个数组中，不能同时出现在两个数组中。请直接返回JSON，不要使用Markdown代码块，不要添加任何额外的解释。确保在返回的JSON中，name字段使用我提供的原始文件名。
		
		`, industry, jobRequirements)

	prompt := ""
	for i, content := range resumeContents {
		filename := "未知文件名"
		if i < len(resumeNames) {
			filename = resumeNames[i]
		}
		prompt += fmt.Sprintf("\n简历内容 (文件名: %s):\n%s\n", filename, content)
	}

	return systemInstruction, prompt
}

// BuildInterviewQuestionsPrompt 构建面试题目生成提示
func BuildInterviewQuestionsPrompt(jobRequirements, industry, resumeContent string, industryKeywords string) (string, string) {
	industryKeywordsContent := "无特殊行业特性"
	if industryKeywords != "" {
		industryKeywordsContent = industryKeywords
	}

	systemInstruction := fmt.Sprintf(`
	你是一个经验丰富的%s行业面试官。请根据以下招聘要求、行业特性和候选人简历，生成20个高质量的针对性面试问题，并提供简洁的参考答案。
	
	行业特性:
	%s
	
	招聘要求:
	%s
	
	请注意以下要求：
	1. 答案必须简洁，每个答案控制在100-150字以内
	2. 只提供关键点，避免冗长解释
	3. 使用要点式回答，便于面试官快速参考
	4. 确保生成完整的JSON且不会因长度过长而被截断
	
	请以下面的JSON格式回复:
	{
	  "questions": [
		{"category": "问题类别", "question": "问题内容", "answer": "简洁的参考答案"},
		...
	  ]
	}
	
	记住：直接返回JSON，不要使用Markdown代码块，不要添加任何额外的解释。确保JSON格式完整有效。
	`, industry, industryKeywordsContent, jobRequirements)

	prompt := fmt.Sprintf("\n候选人简历:\n%s\n", resumeContent)
	return systemInstruction, prompt
}

// BuildInterviewSummaryPrompt 构建面试总结提示
func BuildInterviewSummaryPrompt(jobRequirements, industry, interviewNotes string, industryKeywords string) (string, string) {
	industryKeywordsContent := "无特殊行业特性"
	if industryKeywords != "" {
		industryKeywordsContent = industryKeywords
	}

	systemInstruction := fmt.Sprintf(`
		你是一个经验丰富的%s行业的面试官。参考行业特性，通过面试记录，生成一份简洁的面试总结，提炼候选人的优劣势，给出是否录用的评价。
		
		行业特性:
		%s
		
		请注意以下要求：
		1. 所有评价必须简洁，避免冗长解释
		2. 每项内容控制在50字以内
		3. 使用要点式描述，便于快速阅读
		4. 确保生成完整的JSON且不会因长度过长而被截断
		
		请以下面的JSON格式回复:
		{
		"overall": "总体评价(50字以内)",
		"strengths": ["优势1", "优势2", ...],
		"weaknesses": ["不足1", "不足2", ...],
		"recommendation": "是否推荐录用及简要原因(50字以内)",
		"furtherQuestions": ["需要进一步了解的问题1", "需要进一步了解的问题2", ...],
		"riskPoints": ["风险点1", "风险点2", ...], 
		"suggestions": ["建议1", "建议2", ...]
		}

		记住：直接返回JSON，不要使用Markdown代码块，不要添加任何额外的解释。确保JSON格式完整有效。
		`, industry, industryKeywordsContent)

	prompt := fmt.Sprintf("\n面试记录:\n%s\n", interviewNotes)
	return systemInstruction, prompt
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
