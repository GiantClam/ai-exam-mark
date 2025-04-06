package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GiantClam/homework_marking/models"
	"github.com/GiantClam/homework_marking/services"
	"github.com/gin-gonic/gin"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Answer 表示单个答案
type Answer struct {
	QuestionNumber string `json:"questionNumber"`
	StudentAnswer  string `json:"studentAnswer"`
	IsCorrect      bool   `json:"isCorrect"`
	CorrectAnswer  string `json:"correctAnswer,omitempty"`
	Explanation    string `json:"explanation"`
}

// Student 表示学生信息
type Student struct {
	Name     string
	Content  string
	FilePath string
}

// 提示词模板
const (
	mathPromptTemplate = `请分析以下数学作业：
学生姓名：{{.name}}
作业内容：{{.content}}

请仔细分析学生的答案，特别关注：
1. 计算步骤是否正确
2. 最终答案是否准确
3. 解题思路是否清晰

请以JSON格式返回分析结果。`

	englishPromptTemplate = `请分析以下英语作业：
学生姓名：{{.name}}
作业内容：{{.content}}

请仔细分析学生的答案，特别关注：
1. 语法是否正确
2. 拼写是否准确
3. 表达是否恰当

请以JSON格式返回分析结果。`

	chinesePromptTemplate = `请分析以下语文作业：
学生姓名：{{.name}}
作业内容：{{.content}}

请仔细分析学生的答案，特别关注：
1. 内容是否完整
2. 表达是否准确
3. 思路是否清晰

请以JSON格式返回分析结果。`

	generalPromptTemplate = `请分析以下作业：
学生姓名：{{.name}}
作业内容：{{.content}}

请仔细分析学生的答案，特别关注：
1. 答案是否准确
2. 表达是否清晰
3. 思路是否合理

请以JSON格式返回分析结果。`
)

// processHomeworkImage processes the homework image with appropriate prompts based on type
func processHomeworkImage(imagePath, homeworkType string, customPrompt string) (string, error) {
	log.Printf("[DEBUG] 开始处理作业图片: %s, 类型: %s", imagePath, homeworkType)

	// 检查图片文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Printf("[ERROR] 图片文件不存在: %s", imagePath)
		return "", fmt.Errorf("图片文件不存在: %s", imagePath)
	}

	// 检查图片文件是否可读
	if _, err := os.Open(imagePath); err != nil {
		log.Printf("[ERROR] 无法打开图片文件: %v", err)
		return "", fmt.Errorf("无法打开图片文件: %v", err)
	}

	// 创建AI客户端
	client := services.NewVertexAIClient()
	log.Printf("[DEBUG] 创建AI客户端成功")

	// 根据作业类型设置系统指令
	var systemInstruction string
	switch homeworkType {
	case "english":
		systemInstruction = `
		你是一位专业的英语作业批改助手。请分析学生的英语作业图片，提取其中的手写答案。
		特别注意：
		1. 重点识别括号、下划线等位置的手写答案
		2. 区分学生的手写内容和印刷的题目内容
		3. 判断答案的语法和拼写是否正确
		4. 提供详细的解释和改正建议

		请以下面的JSON格式回复:
		{
		  "answers": [
		    {
		      "questionNumber": "题号",
		      "studentAnswer": "学生的手写答案",
		      "isCorrect": true/false,
		      "correctAnswer": "正确答案（如有错误）",
		      "explanation": "简短解释"
		    }
		  ],
		  "overallScore": "总得分",
		  "feedback": "整体评价和建议"
		}
		
		请只返回标准JSON格式数据，不要使用Markdown代码块（不要使用三个反引号或'json'标记）。
		不要添加任何其他解释性文本。确保JSON格式完整有效，字段名称与示例完全一致。`

	case "chinese":
		systemInstruction = `
		你是一位专业的语文作业批改助手。请分析学生的语文作业图片，提取其中的手写答案。
		特别注意：
		1. 重点识别括号、下划线等位置的手写答案
		2. 区分学生的手写内容和印刷的题目内容
		3. 判断答案的准确性和完整性
		4. 提供详细的解释和建议

		请以下面的JSON格式回复:
		{
		  "answers": [
		    {
		      "questionNumber": "题号",
		      "studentAnswer": "学生的手写答案",
		      "isCorrect": true/false,
		      "correctAnswer": "正确答案（如有错误）",
		      "explanation": "简短解释"
		    }
		  ],
		  "overallScore": "总得分",
		  "feedback": "整体评价和建议"
		}
		
		请只返回标准JSON格式数据，不要使用Markdown代码块（不要使用三个反引号或'json'标记）。
		不要添加任何其他解释性文本。确保JSON格式完整有效，字段名称与示例完全一致。`

	case "math":
		systemInstruction = `
		你是一位专业的数学作业批改助手。请分析学生的数学作业图片，提取其中的手写答案和计算过程。
		特别注意：
		1. 重点识别括号、方框、线条下方等位置的手写答案
		2. 区分学生的手写内容和印刷的题目内容
		3. 识别数学公式、计算过程和最终答案
		4. 判断计算步骤和最终结果是否正确

		请以下面的JSON格式回复:
		{
		  "answers": [
		    {
		      "questionNumber": "题号",
		      "studentAnswer": "学生的手写答案",
		      "isCorrect": true/false,
		      "correctSteps": "正确的解题步骤（如有错误）",
		      "explanation": "简短解释"
		    }
		  ],
		  "overallScore": "总得分",
		  "feedback": "整体评价和建议"
		}
		
		请只返回标准JSON格式数据，不要使用Markdown代码块（不要使用三个反引号或'json'标记）。
		不要添加任何其他解释性文本。确保JSON格式完整有效，字段名称与示例完全一致。`

	default:
		systemInstruction = `
		请分析学生的作业图片，提取其中的手写答案。
		特别注意：
		1. 重点识别括号、下划线、填空处等位置的手写答案
		2. 区分学生的手写内容和印刷的题目内容

		请以下面的JSON格式回复:
		{
		  "answers": [
		    {
		      "questionNumber": "题号",
		      "studentAnswer": "学生的手写答案",
		      "evaluation": "简短评价"
		    }
		  ],
		  "feedback": "整体评价和建议"
		}
		
		请只返回标准JSON格式数据，不要使用Markdown代码块（不要使用三个反引号或'json'标记）。
		不要添加任何其他解释性文本。确保JSON格式完整有效，字段名称与示例完全一致。`
	}

	log.Printf("[DEBUG] 系统指令长度: %d 字符", len(systemInstruction))

	// 设置提示词 - 使用传入的自定义提示词或者创建一个默认提示词
	textPrompt := customPrompt
	if textPrompt == "" {
		textPrompt = fmt.Sprintf("这是一份%s作业，请仔细分析图片中的手写答案，特别关注括号、下划线等位置的手写内容。", homeworkType)
	}

	log.Printf("[DEBUG] 提示词长度: %d 字符", len(textPrompt))

	// 调用Gemini 2.0 Thinking模型分析图片
	response, err := client.GenerateContentWithFile(systemInstruction, imagePath, "image/jpeg", textPrompt)
	if err != nil {
		log.Printf("[ERROR] AI服务处理失败: %v", err)
		return "", fmt.Errorf("AI服务处理失败: %v", err)
	}

	// 验证返回的JSON格式
	var jsonResult interface{}
	if err := json.Unmarshal([]byte(response), &jsonResult); err != nil {
		log.Printf("[ERROR] 解析AI返回的JSON失败: %v", err)
		return "", fmt.Errorf("解析AI返回的JSON失败: %v", err)
	}

	log.Printf("[DEBUG] 成功处理作业图片，返回结果长度: %d 字符", len(response))
	return response, nil
}

// MarkHomework handles homework marking requests
func MarkHomework(c *gin.Context) {
	// 获取文件
	file, err := c.FormFile("homework")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "请上传作业图片",
		})
		return
	}

	// 获取布局参数
	layout := c.DefaultPostForm("layout", "single")

	// 获取作业类型
	homeworkType := c.DefaultPostForm("type", "general")

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "无法打开文件",
		})
		return
	}
	defer src.Close()

	// 创建临时文件
	dst, err := os.CreateTemp("", "homework-*.jpg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "无法创建临时文件",
		})
		return
	}
	defer os.Remove(dst.Name())

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "无法保存文件",
		})
		return
	}

	// 根据布局参数处理图片
	var result interface{}
	if layout == "double" {
		// 双栏布局处理
		result, err = processDoubleColumnImage(dst.Name(), homeworkType)
	} else {
		// 单栏布局处理
		result, err = processSingleColumnImage(dst.Name(), homeworkType)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Result:  result,
	})
}

// 处理双栏布局图片
func processDoubleColumnImage(imagePath, homeworkType string) (interface{}, error) {
	// 直接处理整张图片，在提示词中指明这是双栏布局
	return processImageWithLayout(imagePath, homeworkType, "double")
}

// 处理单栏布局图片
func processSingleColumnImage(imagePath, homeworkType string) (interface{}, error) {
	// 直接处理整张图片，在提示词中指明这是单栏布局
	return processImageWithLayout(imagePath, homeworkType, "single")
}

// 使用特定布局处理图片
func processImageWithLayout(imagePath, homeworkType, layoutType string) (interface{}, error) {
	// 添加详细日志来追踪函数执行
	log.Printf("[DEBUG] 开始处理图片: %s, 类型: %s, 布局: %s", imagePath, homeworkType, layoutType)

	// 检查图片文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Printf("[ERROR] 图片文件不存在: %s", imagePath)
		return nil, fmt.Errorf("图片文件不存在: %s", imagePath)
	}

	// 记录环境变量信息
	log.Printf("[DEBUG] GOOGLE_APPLICATION_CREDENTIALS: %s", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	log.Printf("[DEBUG] GOOGLE_CLOUD_PROJECT: %s", os.Getenv("GOOGLE_CLOUD_PROJECT"))
	log.Printf("[DEBUG] GOOGLE_CLOUD_LOCATION: %s", os.Getenv("GOOGLE_CLOUD_LOCATION"))

	// 调用AI服务处理图片
	log.Printf("[DEBUG] 准备调用AI服务处理图片...")
	// 检查是否处于模拟模式
	log.Printf("[DEBUG] AI服务模拟模式: %v", services.UseMockMode)

	// 根据布局类型调整提示词
	var layoutPrompt string
	if layoutType == "double" {
		layoutPrompt = "这是一份试卷图片，图片是双栏布局。请先分析左半部分的内容，然后再分析右半部分的内容，从上到下处理。"
	} else {
		layoutPrompt = "这是一份普通作业图片，从上到下处理内容。"
	}

	// 完整提示词
	textPrompt := fmt.Sprintf("%s 这是一份%s作业，请仔细分析图片中的手写答案，特别关注括号、下划线等位置的手写内容。",
		layoutPrompt, homeworkType)

	// 调用大模型处理图片
	result, err := processHomeworkImage(imagePath, homeworkType, textPrompt)
	if err != nil {
		log.Printf("[ERROR] 处理图片失败: %v", err)
		return nil, fmt.Errorf("处理图片失败: %v", err)
	}

	// 记录处理结果
	log.Printf("[DEBUG] 图片处理完成，结果长度: %d 字符", len(result))

	// 安全获取结果预览，避免字符串索引超出范围错误
	previewLength := 100
	if len(result) < previewLength {
		previewLength = len(result)
	}
	log.Printf("[DEBUG] 结果预览: %s", result[:previewLength])

	// 返回结果
	return result, nil
}

// 处理单张图片
func processImage(imagePath, homeworkType string) (interface{}, error) {
	// 添加详细日志来追踪函数执行
	log.Printf("[DEBUG] 开始处理图片: %s, 类型: %s", imagePath, homeworkType)

	// 检查图片文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Printf("[ERROR] 图片文件不存在: %s", imagePath)
		return nil, fmt.Errorf("图片文件不存在: %s", imagePath)
	}

	// 检查图片文件是否可读
	if _, err := os.Open(imagePath); err != nil {
		log.Printf("[ERROR] 无法打开图片文件: %v", err)
		return nil, fmt.Errorf("无法打开图片文件: %v", err)
	}

	// 检查文件大小
	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		log.Printf("[ERROR] 获取文件信息失败: %v", err)
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 检查文件大小是否超过限制（25MB的安全限制）
	if fileInfo.Size() > 25*1024*1024 {
		log.Printf("[ERROR] 图片文件过大，超过25MB限制: %d 字节", fileInfo.Size())
		return nil, fmt.Errorf("图片文件过大，超过25MB限制: %d 字节", fileInfo.Size())
	}

	// 记录环境变量信息
	log.Printf("[DEBUG] 环境配置:")
	log.Printf("[DEBUG] - GOOGLE_APPLICATION_CREDENTIALS: %s", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	log.Printf("[DEBUG] - GOOGLE_CLOUD_PROJECT: %s", os.Getenv("GOOGLE_CLOUD_PROJECT"))
	log.Printf("[DEBUG] - GOOGLE_CLOUD_LOCATION: %s", os.Getenv("GOOGLE_CLOUD_LOCATION"))

	// 调用AI服务处理图片
	log.Printf("[DEBUG] 准备调用AI服务处理图片...")
	// 检查是否处于模拟模式
	log.Printf("[DEBUG] AI服务模拟模式: %v", services.UseMockMode)

	// 根据作业类型设置不同的提示词
	var textPrompt string
	switch homeworkType {
	case "english":
		textPrompt = "这是一份英语作业，请仔细分析图片中的手写答案，特别关注括号、下划线等位置的手写内容。"
	case "chinese":
		textPrompt = "这是一份语文作业，请仔细分析图片中的手写答案，特别关注括号、下划线等位置的手写内容。"
	case "math":
		textPrompt = "这是一份数学作业，请仔细分析图片中的手写答案，特别关注括号、方框、线条下方等位置的手写内容。"
	default:
		textPrompt = "这是一份作业，请仔细分析图片中的手写答案，特别关注括号、下划线等位置的手写内容。"
	}

	log.Printf("[DEBUG] 使用提示词: %s", textPrompt)

	// 调用大模型处理图片
	result, err := processHomeworkImage(imagePath, homeworkType, textPrompt)
	if err != nil {
		log.Printf("[ERROR] 处理图片失败: %v", err)
		return nil, fmt.Errorf("处理图片失败: %v", err)
	}

	// 记录处理结果
	log.Printf("[DEBUG] 图片处理完成，结果长度: %d 字符", len(result))

	// 安全获取结果预览，避免字符串索引超出范围错误
	previewLength := 100
	if len(result) < previewLength {
		previewLength = len(result)
	}
	log.Printf("[DEBUG] 结果预览: %s", result[:previewLength])

	// 验证返回的JSON格式
	var jsonResult interface{}
	if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
		log.Printf("[ERROR] 解析JSON结果失败: %v", err)
		return nil, fmt.Errorf("解析JSON结果失败: %v", err)
	}

	log.Printf("[DEBUG] 成功验证JSON格式")
	return result, nil
}

// PDFStudentHomework 表示单个学生的作业信息
type PDFStudentHomework struct {
	StudentName string   `json:"studentName"`
	ClassName   string   `json:"className"`
	Pages       []string `json:"pages"`    // 每页的base64图片数据
	Answers     []Answer `json:"answers"`  // 作业答案
	Score       int      `json:"score"`    // 得分
	Feedback    string   `json:"feedback"` // 总体评价
}

// UploadHomework 处理作业上传
func UploadHomework(c *gin.Context) {
	log.Printf("[INFO] 开始处理作业上传请求")
	startTime := time.Now()

	// 获取表单数据
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("[ERROR] 获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	log.Printf("[INFO] 收到文件: %s, 大小: %d bytes", file.Filename, file.Size)

	// 获取其他表单字段
	homeworkType := c.PostForm("type")
	if homeworkType == "" {
		homeworkType = "general"
	}

	layout := c.PostForm("layout")
	if layout == "" {
		layout = "single"
	}

	pagesPerStudent, _ := strconv.Atoi(c.PostForm("pagesPerStudent"))
	if pagesPerStudent <= 0 {
		pagesPerStudent = 1
	}

	log.Printf("[INFO] 作业参数: type=%s, layout=%s, pagesPerStudent=%d",
		homeworkType, layout, pagesPerStudent)

	// 保存文件到 uploads 目录
	savedFile, err := saveUploadedFile(file)
	if err != nil {
		log.Printf("[ERROR] 保存文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	// 关闭文件但不删除，因为它现在在 uploads 目录中
	savedFile.Close()
	filePath := savedFile.Name()

	log.Printf("[INFO] 文件已保存到: %s", filePath)

	// 根据文件类型和布局选择处理方法
	var result string
	fileExt := strings.ToLower(filepath.Ext(file.Filename))

	if fileExt == ".pdf" && pagesPerStudent > 0 {
		// PDF 文件按学生页数处理
		result, err = processPDFWithStudents(filePath, homeworkType, pagesPerStudent)
	} else {
		// 图片文件或单页 PDF 直接处理
		result, err = processFileDirectly(filePath, homeworkType, layout)
	}

	if err != nil {
		log.Printf("[ERROR] 处理文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("处理文件失败: %v", err)})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"message":  "作业处理成功",
		"results":  result,
		"filePath": filePath, // 返回文件路径，便于前端引用
	})

	duration := time.Since(startTime)
	log.Printf("[INFO] 作业上传处理完成，耗时: %v", duration)
}

// processPDFWithStudents 处理多页PDF，每个学生占用指定页数
func processPDFWithStudents(filePath, homeworkType string, pagesPerStudent int) (string, error) {
	log.Printf("[INFO] 开始处理PDF文件，每个学生 %d 页", pagesPerStudent)

	// 分割PDF
	splitDir := filepath.Join("uploads", "split")
	studentPDFs, err := splitPDF(filePath, pagesPerStudent, splitDir)
	if err != nil {
		return "", fmt.Errorf("分割PDF失败: %v", err)
	}

	log.Printf("[INFO] 成功分割PDF为 %d 个学生作业", len(studentPDFs))

	// 处理每个学生的PDF
	var results []string
	for i, pdfFile := range studentPDFs {
		log.Printf("[INFO] 开始处理第 %d 个学生的作业", i+1)

		// 直接处理PDF文件，不提取内容
		result, err := processFileWithAIDirectly(pdfFile, homeworkType, fmt.Sprintf("学生 %d", i+1))
		if err != nil {
			log.Printf("[ERROR] 处理学生 %d 的PDF失败: %v", i+1, err)
			continue
		}

		results = append(results, result)
		log.Printf("[INFO] 学生 %d 的作业处理成功", i+1)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("没有成功处理任何学生作业")
	}

	// 组合结果
	combinedResult, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("组合结果失败: %v", err)
	}

	return string(combinedResult), nil
}

// processFileDirectly 直接处理单个文件（图片或PDF）
func processFileDirectly(filePath, homeworkType, layout string) (string, error) {
	log.Printf("[INFO] 开始直接处理文件: %s, 类型: %s, 布局: %s", filePath, homeworkType, layout)

	// 如果是双栏布局，在提示词中说明这一点
	var studentName string
	if layout == "double" {
		studentName = "学生（双栏布局）"
	} else {
		studentName = "学生"
	}

	// 直接处理文件，不尝试提取PDF内容
	result, err := processFileWithAIDirectly(filePath, homeworkType, studentName)
	if err != nil {
		return "", fmt.Errorf("处理文件失败: %v", err)
	}

	return result, nil
}

// processFileWithAI 直接使用AI处理文件
func processFileWithAI(filePath, homeworkType, studentName string) (string, error) {
	log.Printf("[INFO] 使用AI处理文件: %s, 类型: %s, 学生: %s", filePath, homeworkType, studentName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("[ERROR] 文件不存在: %s", filePath)
		return "", fmt.Errorf("文件不存在: %s", filePath)
	}

	// 获取文件类型
	fileType := "application/pdf"
	if strings.HasSuffix(filePath, ".jpg") || strings.HasSuffix(filePath, ".jpeg") {
		fileType = "image/jpeg"
	} else if strings.HasSuffix(filePath, ".png") {
		fileType = "image/png"
	}

	// 检查PDF是否有多页
	isMultiPagePDF := false
	pdfPageCount := 1
	fileName := filepath.Base(filePath)
	var pdfAllPagesContent string
	if strings.HasSuffix(filePath, ".pdf") {
		// 获取PDF页数
		pageCount, err := api.PageCountFile(filePath)
		if err != nil {
			log.Printf("[WARN] 获取PDF页数失败: %v", err)
		} else {
			pdfPageCount = pageCount
			if pageCount > 1 {
				isMultiPagePDF = true
				log.Printf("[INFO] 检测到多页PDF: %s, 共 %d 页", fileName, pageCount)

				// 提取所有页面的内容，确保大模型能分析全部内容
				conf := model.NewDefaultConfiguration()
				var allContentBuilder strings.Builder
				extractedContentCount := 0

				for page := 1; page <= pageCount; page++ {
					log.Printf("[INFO] 正在提取PDF第 %d/%d 页内容", page, pageCount)
					pageContent, err := extractSinglePageContent(filePath, page, conf)
					if err != nil {
						log.Printf("[WARN] 提取第 %d 页内容失败: %v", page, err)
						allContentBuilder.WriteString(fmt.Sprintf("[页面 %d 内容提取失败]\n\n", page))
					} else {
						allContentBuilder.WriteString(fmt.Sprintf("--- 页面 %d ---\n", page))
						allContentBuilder.WriteString(pageContent)
						allContentBuilder.WriteString("\n\n")
						extractedContentCount++
					}
				}

				pdfAllPagesContent = allContentBuilder.String()
				log.Printf("[INFO] 成功提取 %d/%d 页PDF内容，总长度: %d 字符",
					extractedContentCount, pageCount, len(pdfAllPagesContent))
			}
		}
	}

	// 根据作业类型创建系统提示词
	var systemPrompt string
	multiPageInstruction := ""
	if isMultiPagePDF {
		multiPageInstruction = fmt.Sprintf(`
注意：这是一个包含 %d 页的PDF文件。请务必仔细分析所有页面的内容，不要只关注第一页。
每个页面都包含重要信息，请完整处理所有页面内容后再给出评分和反馈。
下面是我已经提取的各页内容：

%s`, pdfPageCount, pdfAllPagesContent)
	}

	switch homeworkType {
	case "english":
		systemPrompt = fmt.Sprintf(`
你是一位专业的英语作业批改助手。请分析 %s 的英语作业，仔细提取并评价作业中的内容。
%s

请以下面的JSON格式回复:
{
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的答案",
      "isCorrect": true/false,
      "correctAnswer": "正确答案（如有错误）",
      "explanation": "简短解释"
    }
  ],
  "overallScore": "总得分",
  "feedback": "整体评价和建议"
}

请确保输出为有效的JSON格式，不要包含额外的文本。`, studentName, multiPageInstruction)

	case "chinese":
		systemPrompt = fmt.Sprintf(`
你是一位专业的语文作业批改助手。请分析 %s 的语文作业，仔细提取并评价作业中的内容。
%s

请以下面的JSON格式回复:
{
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的答案",
      "isCorrect": true/false,
      "correctAnswer": "正确答案（如有错误）",
      "explanation": "简短解释"
    }
  ],
  "overallScore": "总得分",
  "feedback": "整体评价和建议"
}

请确保输出为有效的JSON格式，不要包含额外的文本。`, studentName, multiPageInstruction)

	case "math":
		systemPrompt = fmt.Sprintf(`
你是一位专业的数学作业批改助手。请分析 %s 的数学作业，仔细提取并评价作业中的内容。
%s

请以下面的JSON格式回复:
{
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的答案",
      "isCorrect": true/false,
      "correctSteps": "正确的解题步骤（如有错误）",
      "explanation": "简短解释"
    }
  ],
  "overallScore": "总得分",
  "feedback": "整体评价和建议"
}

请确保输出为有效的JSON格式，不要包含额外的文本。`, studentName, multiPageInstruction)

	default:
		systemPrompt = fmt.Sprintf(`
请分析 %s 的作业，仔细提取并评价作业中的内容。
%s

请以下面的JSON格式回复:
{
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的答案",
      "evaluation": "简短评价"
    }
  ],
  "feedback": "整体评价和建议"
}

请确保输出为有效的JSON格式，不要包含额外的文本。`, studentName, multiPageInstruction)
	}

	// 创建AI客户端
	client := services.NewVertexAIClient()
	log.Printf("[DEBUG] 创建AI客户端成功")

	// 判断是否需要发送文件或已提取的内容
	var result string
	var err error

	if isMultiPagePDF && pdfAllPagesContent != "" {
		// 如果是多页PDF且我们已经提取了所有内容，直接发送文本而不是文件
		log.Printf("[INFO] 使用提取的多页PDF内容调用AI服务，而不是直接发送文件")

		// 设置文本提示词
		textPrompt := fmt.Sprintf("请分析这份%s作业，提取并评价内容。这是已经提取的 %d 页PDF文本内容，请务必分析所有页面。",
			homeworkType, pdfPageCount)

		// 调用大模型处理文本内容
		result, err = client.GenerateContent(systemPrompt, textPrompt+"\n\n"+pdfAllPagesContent)
	} else {
		// 对于其他情况，直接使用文件
		log.Printf("[INFO] 直接发送文件给AI服务进行分析")

		// 设置提示词
		var textPrompt string
		if isMultiPagePDF {
			textPrompt = fmt.Sprintf("请分析这份%s作业，提取并评价内容。这是一份包含 %d 页的PDF文件，请务必分析所有页面内容，不要只关注第一页。",
				homeworkType, pdfPageCount)
		} else {
			textPrompt = fmt.Sprintf("请分析这份%s作业，提取并评价内容。", homeworkType)
		}

		// 调用大模型处理文件
		result, err = client.GenerateContentWithFile(systemPrompt, filePath, fileType, textPrompt)
	}

	// 添加重试机制
	maxRetries := 2
	retryCount := 0
	for err != nil && retryCount < maxRetries {
		retryCount++
		log.Printf("[INFO] 第 %d 次重试调用AI服务", retryCount)

		// 等待一段时间后重试
		time.Sleep(2 * time.Second)

		if isMultiPagePDF && pdfAllPagesContent != "" {
			textPrompt := fmt.Sprintf("重试: 请分析这份%s作业，提取并评价内容。", homeworkType)
			result, err = client.GenerateContent(systemPrompt, textPrompt+"\n\n"+pdfAllPagesContent)
		} else {
			textPrompt := fmt.Sprintf("重试: 请分析这份%s作业，提取并评价内容。", homeworkType)
			result, err = client.GenerateContentWithFile(systemPrompt, filePath, fileType, textPrompt)
		}
	}

	if err != nil {
		log.Printf("[ERROR] 多次尝试后AI服务仍处理失败: %v", err)
		return "", fmt.Errorf("AI服务处理失败: %v", err)
	}

	// 验证返回的JSON格式
	var jsonResult interface{}
	if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
		log.Printf("[ERROR] 解析AI返回的JSON失败: %v", err)

		// 尝试清理JSON
		cleanedResult := services.EnsureCompleteJSON(result)
		if err := json.Unmarshal([]byte(cleanedResult), &jsonResult); err != nil {
			log.Printf("[ERROR] 清理后JSON仍然无效: %v", err)
			return "", fmt.Errorf("解析AI返回的JSON失败: %v", err)
		}

		log.Printf("[INFO] 清理后JSON有效，使用清理后的结果")
		result = cleanedResult
	}

	log.Printf("[INFO] AI服务处理成功，结果长度: %d 字符", len(result))
	return result, nil
}

// processFileWithAIDirectly 直接使用大模型处理文件，不尝试提取内容
func processFileWithAIDirectly(filePath, homeworkType, studentName string) (string, error) {
	log.Printf("[INFO] 直接使用大模型处理文件: %s, 类型: %s, 学生: %s", filePath, homeworkType, studentName)

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("[ERROR] 文件不存在或无法访问: %s, 错误: %v", filePath, err)
		return "", fmt.Errorf("文件不存在或无法访问: %s", filePath)
	}

	log.Printf("[INFO] 文件大小: %d 字节", fileInfo.Size())

	// 获取文件类型
	fileExt := strings.ToLower(filepath.Ext(filePath))
	var fileType string
	switch fileExt {
	case ".pdf":
		fileType = "application/pdf"
	case ".jpg", ".jpeg":
		fileType = "image/jpeg"
	case ".png":
		fileType = "image/png"
	default:
		log.Printf("[ERROR] 不支持的文件类型: %s", fileExt)
		return "", fmt.Errorf("不支持的文件类型: %s", fileExt)
	}

	// 检查PDF页数以在提示词中说明
	pdfPageNote := ""
	if fileExt == ".pdf" {
		pageCount, err := api.PageCountFile(filePath)
		if err != nil {
			log.Printf("[WARN] 获取PDF页数失败: %v", err)
		} else if pageCount > 1 {
			pdfPageNote = fmt.Sprintf(" (这是一个包含%d页的PDF文件，请务必分析所有页面)", pageCount)
			log.Printf("[INFO] 检测到多页PDF: %s, 共 %d 页", filepath.Base(filePath), pageCount)
		}
	}

	// 根据作业类型创建系统提示词
	systemPrompt := getPromptForHomeworkType(homeworkType)

	// 创建AI客户端
	client := services.NewVertexAIClient()
	log.Printf("[INFO] 创建AI客户端成功")

	// 设置提示词
	textPrompt := fmt.Sprintf("请分析这份%s作业%s，提取并评价内容。请详细分析其中的问题和答案，提供评分和建议。",
		homeworkType, pdfPageNote)

	// 增强重试机制
	maxRetries := 5 // 增加重试次数
	var lastError error
	var result string

	for retryCount := 0; retryCount <= maxRetries; retryCount++ {
		// 如果是重试，等待一段时间后再进行
		if retryCount > 0 {
			// 使用指数退避策略，但添加一些随机性避免所有请求同时进行
			backoffSeconds := (1 << uint(retryCount)) + rand.Intn(5)
			if backoffSeconds > 60 {
				backoffSeconds = 60 // 最大等待60秒
			}
			backoffDuration := time.Duration(backoffSeconds) * time.Second

			log.Printf("[INFO] 第 %d/%d 次重试，等待 %v 后进行", retryCount, maxRetries, backoffDuration)
			time.Sleep(backoffDuration)

			// 在每次重试时检查文件是否仍然存在和可访问
			if _, err := os.Stat(filePath); err != nil {
				log.Printf("[ERROR] 重试前检查文件失败: %v", err)
				return "", fmt.Errorf("重试前检查文件失败: %v", err)
			}
		}

		// 修改提示词，增加重试信息
		currentPrompt := textPrompt
		if retryCount > 0 {
			currentPrompt = fmt.Sprintf("重试(%d/%d): 请分析这份%s作业%s，提取并评价内容。请详细分析其中的问题和答案，提供评分和建议。",
				retryCount, maxRetries, homeworkType, pdfPageNote)
		}

		// 记录开始/重试信息
		infoPrefix := "开始"
		if retryCount > 0 {
			infoPrefix = "重试"
		}
		log.Printf("[INFO] %s调用AI服务处理文件", infoPrefix)

		// 调用AI服务
		result, err = client.GenerateContentWithFile(systemPrompt, filePath, fileType, currentPrompt)

		// 如果成功获取结果，跳出循环
		if err == nil && result != "" {
			log.Printf("[INFO] AI服务处理成功，跳出重试循环")
			break
		}

		// 记录错误，作为最后一次错误
		lastError = err

		if err != nil {
			// 详细记录错误信息
			log.Printf("[ERROR] 第 %d/%d 次调用AI服务失败: %v", retryCount+1, maxRetries+1, err)

			// 检查特定类型的错误
			if strings.Contains(err.Error(), "内容被安全策略限制") {
				log.Printf("[ERROR] 内容因安全策略被拒绝，不再重试")
				return "", err // 这种错误不需要重试
			}
		} else if result == "" {
			log.Printf("[ERROR] 第 %d/%d 次调用AI服务未返回有效内容", retryCount+1, maxRetries+1)
		}

		// 如果是最后一次重试，退出循环
		if retryCount == maxRetries {
			log.Printf("[ERROR] 达到最大重试次数 %d，放弃处理", maxRetries)
		}
	}

	// 如果所有重试都失败
	if result == "" {
		if lastError != nil {
			log.Printf("[ERROR] 多次尝试后AI服务仍失败: %v", lastError)
			return "", fmt.Errorf("AI服务处理失败: %v", lastError)
		}
		log.Printf("[ERROR] 多次尝试后AI服务未返回有效内容")
		return "", fmt.Errorf("AI服务未返回有效内容")
	}

	// 验证返回的JSON格式
	var jsonResult interface{}
	if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
		log.Printf("[ERROR] 解析AI返回的JSON失败: %v", err)

		// 尝试清理JSON
		cleanedResult := services.EnsureCompleteJSON(result)
		if err := json.Unmarshal([]byte(cleanedResult), &jsonResult); err != nil {
			log.Printf("[ERROR] 清理后JSON仍然无效: %v", err)
			return "", fmt.Errorf("解析AI返回的JSON失败: %v", err)
		}

		log.Printf("[INFO] 清理后JSON有效，使用清理后的结果")
		result = cleanedResult
	}

	log.Printf("[INFO] AI服务处理成功，结果长度: %d 字符", len(result))
	return result, nil
}

// 添加一个辅助函数，获取作业类型对应的提示词
func getPromptForHomeworkType(homeworkType string) string {
	switch homeworkType {
	case "english":
		return `你是一位专业的英语作业批改助手。请分析学生的英语作业，仔细提取并评价作业中的内容。

请以下面的JSON格式回复:
{
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的答案",
      "isCorrect": true/false,
      "correctAnswer": "正确答案（如有错误）",
      "explanation": "简短解释"
    }
  ],
  "overallScore": "总得分",
  "feedback": "整体评价和建议"
}

请确保输出为有效的JSON格式，不要包含额外的文本。不要使用Markdown的代码块标记。`

	case "chinese":
		return `你是一位专业的语文作业批改助手。请分析学生的语文作业，仔细提取并评价作业中的内容。

请以下面的JSON格式回复:
{
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的答案",
      "isCorrect": true/false,
      "correctAnswer": "正确答案（如有错误）",
      "explanation": "简短解释"
    }
  ],
  "overallScore": "总得分",
  "feedback": "整体评价和建议"
}

请确保输出为有效的JSON格式，不要包含额外的文本。不要使用Markdown的代码块标记。`

	case "math":
		return `你是一位专业的数学作业批改助手。请分析学生的数学作业，仔细提取并评价作业中的内容。

请以下面的JSON格式回复:
{
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的答案",
      "isCorrect": true/false,
      "correctSteps": "正确的解题步骤（如有错误）",
      "explanation": "简短解释"
    }
  ],
  "overallScore": "总得分",
  "feedback": "整体评价和建议"
}

请确保输出为有效的JSON格式，不要包含额外的文本。不要使用Markdown的代码块标记。`

	default:
		return `请分析学生的作业，仔细提取并评价作业中的内容。

请以下面的JSON格式回复:
{
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的答案",
      "evaluation": "简短评价"
    }
  ],
  "feedback": "整体评价和建议"
}

请确保输出为有效的JSON格式，不要包含额外的文本。不要使用Markdown的代码块标记。`
	}
}

// splitPDF 分割PDF文件
func splitPDF(inputFile string, pagesPerStudent int, outputDir string) ([]string, error) {
	log.Printf("[INFO] 开始分割PDF文件: %s, 每个学生 %d 页", inputFile, pagesPerStudent)

	// 检查输入文件是否存在
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		log.Printf("[ERROR] 输入PDF文件不存在: %s", inputFile)
		return nil, fmt.Errorf("输入PDF文件不存在: %s", inputFile)
	}

	// 创建分割文件的目录: uploads/split
	splitDir := filepath.Join("uploads", "split")
	if err := os.MkdirAll(splitDir, 0755); err != nil {
		log.Printf("[ERROR] 创建分割文件目录失败: %s, 错误: %v", splitDir, err)
		return nil, fmt.Errorf("创建分割文件目录失败: %v", err)
	}

	log.Printf("[DEBUG] 已创建分割文件目录: %s", splitDir)

	// 创建配置
	conf := api.LoadConfiguration()

	// 获取PDF页数
	pageCount, err := api.PageCountFile(inputFile)
	if err != nil {
		log.Printf("[ERROR] 获取PDF页数失败: %v", err)
		return nil, fmt.Errorf("获取PDF页数失败: %v", err)
	}

	log.Printf("[INFO] PDF文件总页数: %d, 每个学生 %d 页", pageCount, pagesPerStudent)

	// 检查页数是否为零
	if pageCount <= 0 {
		log.Printf("[ERROR] PDF文件页数为零或无效")
		return nil, fmt.Errorf("PDF文件页数为零或无效")
	}

	// 计算需要分割的文件数
	numFiles := (pageCount + pagesPerStudent - 1) / pagesPerStudent
	var outputFiles []string

	// 获取输入文件的基本名称（不含路径和扩展名）
	baseFileName := filepath.Base(inputFile)
	fileExt := filepath.Ext(baseFileName)
	baseFileName = strings.TrimSuffix(baseFileName, fileExt)

	log.Printf("[DEBUG] 基本文件名: %s, 扩展名: %s", baseFileName, fileExt)

	// 为这次分割创建唯一的时间戳目录
	timestamp := time.Now().Format("20060102_150405")
	splitSessionDir := filepath.Join(splitDir, fmt.Sprintf("%s_%s", baseFileName, timestamp))

	if err := os.MkdirAll(splitSessionDir, 0755); err != nil {
		log.Printf("[ERROR] 创建分割会话目录失败: %s, 错误: %v", splitSessionDir, err)
		return nil, fmt.Errorf("创建分割会话目录失败: %v", err)
	}

	log.Printf("[INFO] 创建分割会话目录: %s", splitSessionDir)

	// 分割PDF - 为每个学生创建一个包含多页的PDF
	for i := 0; i < numFiles; i++ {
		startPage := i*pagesPerStudent + 1
		endPage := min((i+1)*pagesPerStudent, pageCount)

		// 创建最终输出文件名
		studentPDFName := fmt.Sprintf("student_%d.pdf", i+1)
		studentPDFFile := filepath.Join(splitSessionDir, studentPDFName)

		log.Printf("[INFO] 处理学生 %d, 页面范围: %d-%d, 输出文件: %s",
			i+1, startPage, endPage, studentPDFFile)

		// 如果之前的文件存在，先删除
		if fileExists(studentPDFFile) {
			if err := os.Remove(studentPDFFile); err != nil {
				log.Printf("[WARN] 无法删除已存在的文件: %s, 错误: %v", studentPDFFile, err)
			}
		}

		// 创建临时提取目录
		tempExtractDir := filepath.Join(splitSessionDir, fmt.Sprintf("temp_extract_%d", i+1))
		if err := os.MkdirAll(tempExtractDir, 0755); err != nil {
			return nil, fmt.Errorf("创建临时提取目录失败: %v", err)
		}

		// 构建页面范围
		var pageRanges []string
		for page := startPage; page <= endPage; page++ {
			pageRanges = append(pageRanges, fmt.Sprintf("%d", page))
		}

		log.Printf("[DEBUG] 提取页面范围: %v 到临时目录: %s", pageRanges, tempExtractDir)

		// 提取页面到临时目录 - ExtractPagesFile默认会生成带有page_*后缀的文件
		err = api.ExtractPagesFile(inputFile, tempExtractDir, pageRanges, conf)
		if err != nil {
			log.Printf("[ERROR] 提取页面 %d-%d 失败: %v", startPage, endPage, err)
			os.RemoveAll(tempExtractDir) // 清理临时目录

			// 添加临时调试日志
			log.Printf("[DEBUG] 当前工作目录: %s", getCurrentDirectory())
			log.Printf("[DEBUG] 输入文件存在: %v", fileExists(inputFile))
			log.Printf("[DEBUG] 输出目录存在: %v", fileExists(splitSessionDir))
			log.Printf("[DEBUG] 输出目录可写: %v", directoryWritable(splitSessionDir))

			// 尝试备选方法: 单独提取每一页
			log.Printf("[INFO] 尝试使用备选方法提取页面范围...")

			// 创建临时目录存放各个页面
			tempDir, err := os.MkdirTemp("", "pdf-pages-*")
			if err != nil {
				return nil, fmt.Errorf("创建临时目录失败: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// 逐页提取
			var pageFiles []string
			for page := startPage; page <= endPage; page++ {
				singlePageFile := filepath.Join(tempDir, fmt.Sprintf("page_%d.pdf", page))
				log.Printf("[DEBUG] 提取单页 %d 到 %s", page, singlePageFile)

				// 提取单页
				pageRange := []string{fmt.Sprintf("%d", page)}
				if err := api.ExtractPagesFile(inputFile, singlePageFile, pageRange, conf); err != nil {
					log.Printf("[ERROR] 提取单页 %d 失败: %v", page, err)
					continue
				}

				// 检查生成的文件名可能有 page_ 前缀
				if !fileExists(singlePageFile) {
					// 尝试查找真实生成的文件
					expectedPattern := fmt.Sprintf("*page_%s.pdf", pageRange[0])
					matches, _ := filepath.Glob(filepath.Join(tempDir, expectedPattern))
					if len(matches) > 0 {
						// 使用找到的文件
						singlePageFile = matches[0]
					} else {
						log.Printf("[WARN] 找不到提取的页面文件，跳过")
						continue
					}
				}

				if fileExists(singlePageFile) {
					pageFiles = append(pageFiles, singlePageFile)
				}
			}

			// 检查是否成功提取了页面
			if len(pageFiles) == 0 {
				return nil, fmt.Errorf("无法提取任何页面")
			}

			// 对于只有一页的情况，直接复制
			if len(pageFiles) == 1 {
				if err := copyFile(pageFiles[0], studentPDFFile); err != nil {
					return nil, fmt.Errorf("复制单页文件失败: %v", err)
				}
			} else {
				// 对于多页，尝试使用 pdfcpu 的合并功能
				log.Printf("[INFO] 合并 %d 个页面到 %s", len(pageFiles), studentPDFFile)

				// 使用PDFCPU的合并功能 (参数：输入文件数组，输出文件，保留书签布尔值，配置)
				err = api.MergeCreateFile(pageFiles, studentPDFFile, false, conf)
				if err != nil {
					log.Printf("[ERROR] 合并页面失败: %v", err)
					// 如果合并失败，至少保留第一页
					if err := copyFile(pageFiles[0], studentPDFFile); err != nil {
						return nil, fmt.Errorf("无法保存任何页面: %v", err)
					}
					log.Printf("[WARN] 合并失败，仅保留第一页")
				}
			}
		} else {
			// 提取成功，处理临时目录中的文件
			log.Printf("[INFO] 页面提取成功，查找提取的文件")

			// 查找所有生成的PDF文件
			extractedFiles, _ := filepath.Glob(filepath.Join(tempExtractDir, "*.pdf"))
			if len(extractedFiles) == 0 {
				log.Printf("[ERROR] 临时目录中未找到提取的PDF文件")
				os.RemoveAll(tempExtractDir) // 清理临时目录
				return nil, fmt.Errorf("未找到提取的PDF文件")
			}

			// 按修改时间排序文件，确保页面顺序正确
			sortFilesByModTime(extractedFiles)

			// 如果只有一个文件，直接复制
			if len(extractedFiles) == 1 {
				log.Printf("[INFO] 只有一个提取文件，直接复制到 %s", studentPDFFile)
				if err := copyFile(extractedFiles[0], studentPDFFile); err != nil {
					os.RemoveAll(tempExtractDir) // 清理临时目录
					return nil, fmt.Errorf("复制文件失败: %v", err)
				}
			} else if len(extractedFiles) > 1 {
				// 如果有多个文件，需要合并
				log.Printf("[INFO] 合并 %d 个提取文件到 %s", len(extractedFiles), studentPDFFile)
				err = api.MergeCreateFile(extractedFiles, studentPDFFile, false, conf)
				if err != nil {
					log.Printf("[ERROR] 合并提取文件失败: %v", err)
					// 如果合并失败，至少保留第一个文件
					if err := copyFile(extractedFiles[0], studentPDFFile); err != nil {
						os.RemoveAll(tempExtractDir) // 清理临时目录
						return nil, fmt.Errorf("无法保存任何页面: %v", err)
					}
					log.Printf("[WARN] 合并失败，仅保留第一个文件")
				}
			}

			// 清理临时目录
			os.RemoveAll(tempExtractDir)
		}

		// 验证输出文件存在
		if !fileExists(studentPDFFile) {
			log.Printf("[ERROR] 输出文件未创建: %s", studentPDFFile)
			return nil, fmt.Errorf("输出文件未创建: %s", studentPDFFile)
		}

		// 验证输出文件大小
		fileInfo, err := os.Stat(studentPDFFile)
		if err != nil {
			log.Printf("[WARN] 无法获取输出文件信息: %s, 错误: %v", studentPDFFile, err)
		} else if fileInfo.Size() == 0 {
			log.Printf("[ERROR] 输出文件大小为零: %s", studentPDFFile)
			return nil, fmt.Errorf("输出文件大小为零: %s", studentPDFFile)
		} else {
			log.Printf("[DEBUG] 输出文件创建成功: %s, 大小: %d 字节", studentPDFFile, fileInfo.Size())
		}

		// 确认确实提取了正确数量的页面
		extractedPageCount, err := api.PageCountFile(studentPDFFile)
		if err != nil {
			log.Printf("[WARN] 无法获取提取文件的页数: %v", err)
		} else {
			expectedPageCount := endPage - startPage + 1
			log.Printf("[INFO] 提取的文件页数: %d, 预期页数: %d", extractedPageCount, expectedPageCount)

			if extractedPageCount != expectedPageCount {
				log.Printf("[WARN] 提取的页数与预期不符！")
			}
		}

		outputFiles = append(outputFiles, studentPDFFile)
		log.Printf("[INFO] 成功创建学生 %d 的PDF文件", i+1)
	}

	log.Printf("[INFO] PDF分割完成，共创建 %d 个文件, 保存在目录: %s", len(outputFiles), splitSessionDir)
	return outputFiles, nil
}

// 获取当前工作目录（辅助函数）
func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("获取工作目录失败: %v", err)
	}
	return dir
}

// 检查文件是否存在（辅助函数）
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// 检查目录是否可写（辅助函数）
func directoryWritable(dir string) bool {
	// 尝试在目录中创建一个临时文件
	tempFile, err := os.CreateTemp(dir, "write-test-*.tmp")
	if err != nil {
		return false
	}

	// 关闭并删除临时文件
	tempFile.Close()
	os.Remove(tempFile.Name())
	return true
}

// extractStudentInfo 从图片中提取学生信息
func extractStudentInfo(image string) (struct {
	Name  string
	Class string
}, error) {
	// 使用Gemini模型识别学生信息
	// 这里需要实现具体的识别逻辑
	return struct {
		Name  string
		Class string
	}{}, nil
}

// calculateScore 计算作业得分
func calculateScore(answers []Answer) int {
	// 实现得分计算逻辑
	return 0
}

// generateFeedback 生成作业评价
func generateFeedback(answers []Answer, score int) string {
	// 实现评价生成逻辑
	return ""
}

// saveUploadedFile 保存上传的文件
func saveUploadedFile(file *multipart.FileHeader) (*os.File, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 确保 uploads 目录存在
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建 uploads 目录失败: %v", err)
	}

	// 生成唯一文件名
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filePath := filepath.Join(uploadsDir, fileName)

	log.Printf("[INFO] 保存上传文件到: %s", filePath)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %v", err)
	}

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(filePath)
		return nil, fmt.Errorf("保存文件内容失败: %v", err)
	}

	// 重新将文件指针定位到文件开头，以便后续读取
	if _, err := dst.Seek(0, 0); err != nil {
		dst.Close()
		return nil, fmt.Errorf("重置文件指针失败: %v", err)
	}

	return dst, nil
}

// processFile 处理上传的文件
func processFile(filePath, layout string, pagesPerStudent int) ([]Student, error) {
	log.Printf("[INFO] 开始处理文件: %s", filePath)

	// 检查文件类型
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 根据文件类型处理
	if strings.HasSuffix(filePath, ".pdf") {
		return processPDFFile(filePath, pagesPerStudent)
	} else if strings.HasSuffix(filePath, ".jpg") || strings.HasSuffix(filePath, ".jpeg") || strings.HasSuffix(filePath, ".png") {
		return processImageFile(filePath, layout)
	} else {
		return nil, fmt.Errorf("不支持的文件类型")
	}
}

// processPDFFile 处理PDF文件
func processPDFFile(filePath string, pagesPerStudent int) ([]Student, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "pdf-split-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	log.Printf("[INFO] 开始处理PDF文件，每个学生 %d 页", pagesPerStudent)

	// 分割PDF
	studentPDFs, err := splitPDF(filePath, pagesPerStudent, tempDir)
	if err != nil {
		return nil, fmt.Errorf("分割PDF失败: %v", err)
	}

	log.Printf("[INFO] 成功分割PDF为 %d 个学生文件", len(studentPDFs))

	// 处理每个学生的PDF
	var students []Student
	for i, pdfFile := range studentPDFs {
		log.Printf("[INFO] 开始读取学生 %d 的PDF: %s", i+1, pdfFile)

		// 读取PDF内容
		content, err := readPDFContent(pdfFile)
		if err != nil {
			log.Printf("[ERROR] 读取学生 %d 的PDF失败: %v", i+1, err)
			continue
		}

		students = append(students, Student{
			Name:    fmt.Sprintf("学生 %d", i+1),
			Content: content,
		})

		log.Printf("[INFO] 成功读取学生 %d 的PDF内容，长度: %d", i+1, len(content))
	}

	return students, nil
}

// processImageFile 处理图片文件
func processImageFile(filePath, layout string) ([]Student, error) {
	// 读取图片内容
	content, err := readImageContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取图片失败: %v", err)
	}

	// 根据布局处理
	if layout == "double" {
		// 双栏布局，分割为两个学生的作业
		return []Student{
			{Name: "学生 1", Content: content[:len(content)/2]},
			{Name: "学生 2", Content: content[len(content)/2:]},
		}, nil
	} else {
		// 单栏布局，作为一个学生的作业
		return []Student{
			{Name: "学生 1", Content: content},
		}, nil
	}
}

// readPDFContent 读取PDF文件内容
func readPDFContent(filePath string) (string, error) {
	log.Printf("[INFO] 开始读取PDF内容: %s", filePath)

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("PDF文件不存在: %s", filePath)
	}

	// 打开PDF文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开PDF文件失败: %v", err)
	}
	defer file.Close()

	// 创建配置
	conf := model.NewDefaultConfiguration()

	// 获取PDF页数
	pageCount, err := api.PageCountFile(filePath)
	if err != nil {
		return "", fmt.Errorf("获取PDF页数失败: %v", err)
	}

	log.Printf("[INFO] PDF文件 %s 共有 %d 页", filePath, pageCount)

	// 使用临时文件存储提取的文本
	tempFile, err := os.CreateTemp("", "pdf-text-*.txt")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 提取所有页面的文本内容
	err = api.ExtractContentFile(filePath, tempFile.Name(), nil, conf)
	if err != nil {
		log.Printf("[ERROR] 提取PDF内容失败: %v", err)

		// 尝试使用替代方法：逐页提取
		var allContent strings.Builder
		for page := 1; page <= pageCount; page++ {
			log.Printf("[INFO] 尝试提取第 %d 页内容", page)
			pageContent, pageErr := extractSinglePageContent(filePath, page, conf)
			if pageErr != nil {
				log.Printf("[WARN] 提取第 %d 页失败: %v", page, pageErr)
				continue
			}
			if pageContent != "" {
				allContent.WriteString(fmt.Sprintf("--- 第 %d 页 ---\n", page))
				allContent.WriteString(pageContent)
				allContent.WriteString("\n\n")
			}
		}

		if allContent.Len() > 0 {
			content := allContent.String()
			log.Printf("[INFO] 成功使用替代方法提取PDF内容，长度: %d", len(content))
			return content, nil
		}

		return "", fmt.Errorf("提取PDF内容失败: %v", err)
	}

	// 读取提取的文本内容
	data, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("读取提取的文本内容失败: %v", err)
	}

	content := string(data)

	// 如果内容为空，尝试使用OCR或其他方法
	if strings.TrimSpace(content) == "" {
		log.Printf("[WARN] 提取的PDF文本内容为空，可能是扫描PDF或图片PDF")
		// 在这里可以添加OCR处理逻辑
		// 暂时返回提示信息
		return "[这是一个扫描PDF或图片PDF，无法直接提取文本。系统将尝试使用图像识别方式处理]", nil
	}

	log.Printf("[INFO] PDF内容读取成功，长度: %d", len(content))
	return content, nil
}

// extractSinglePageContent 提取单页PDF内容
func extractSinglePageContent(filePath string, pageNum int, conf *model.Configuration) (string, error) {
	log.Printf("[INFO] 开始提取PDF单页内容: %s, 页码: %d", filePath, pageNum)

	// 检查输入文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("[ERROR] 输入PDF文件不存在: %s", filePath)
		return "", fmt.Errorf("输入PDF文件不存在: %s", filePath)
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "pdf-page-*")
	if err != nil {
		log.Printf("[ERROR] 创建临时目录失败: %v", err)
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer func() {
		log.Printf("[DEBUG] 清理临时目录: %s", tempDir)
		os.RemoveAll(tempDir)
	}()

	log.Printf("[DEBUG] 已创建临时目录: %s", tempDir)

	// 获取输入文件的基本名称（不含路径和扩展名）
	baseFileName := filepath.Base(filePath)
	fileExt := filepath.Ext(baseFileName)
	baseFileName = strings.TrimSuffix(baseFileName, fileExt)

	// 页面范围
	pageRange := []string{fmt.Sprintf("%d", pageNum)}

	log.Printf("[DEBUG] 提取单页PDF, 输入: %s, 输出目录: %s, 页面: %v",
		filePath, tempDir, pageRange)

	// 尝试提取单页，传入输出目录
	err = api.ExtractPagesFile(filePath, tempDir, pageRange, conf)
	if err != nil {
		log.Printf("[ERROR] 提取单页失败: %v", err)
		return "", fmt.Errorf("提取单页失败: %v", err)
	}

	// 查找生成的文件 - 根据具体的命名格式
	expectedFileName := fmt.Sprintf("%s_page_%s.pdf", baseFileName, pageRange[0])
	expectedFilePath := filepath.Join(tempDir, expectedFileName)

	var singlePageFile string

	// 如果精确的文件存在，直接使用
	if fileExists(expectedFilePath) {
		singlePageFile = expectedFilePath
		log.Printf("[DEBUG] 找到预期的页面文件: %s", singlePageFile)
	} else {
		// 否则，尝试查找任何PDF文件
		pageFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.pdf"))
		if len(pageFiles) == 0 {
			log.Printf("[ERROR] 找不到提取的PDF页面文件")
			return "", fmt.Errorf("找不到提取的PDF页面文件")
		}

		// 使用找到的第一个文件（通常只会有一个）
		singlePageFile = pageFiles[0]
		log.Printf("[DEBUG] 找到生成的页面文件: %s", singlePageFile)
	}

	// 验证输出文件存在
	if _, err := os.Stat(singlePageFile); os.IsNotExist(err) {
		log.Printf("[ERROR] 提取的单页文件不存在: %s", singlePageFile)
		return "", fmt.Errorf("提取的单页文件不存在: %s", singlePageFile)
	}

	// 创建临时文本文件
	tempFile, err := os.CreateTemp(tempDir, "page-text-*.txt")
	if err != nil {
		log.Printf("[ERROR] 创建临时文本文件失败: %v", err)
		return "", fmt.Errorf("创建临时文本文件失败: %v", err)
	}
	defer func() {
		tempFile.Close()
		log.Printf("[DEBUG] 清理临时文本文件: %s", tempFile.Name())
		os.Remove(tempFile.Name())
	}()

	log.Printf("[DEBUG] 已创建临时文本文件: %s", tempFile.Name())

	// 提取文本
	err = api.ExtractContentFile(singlePageFile, tempFile.Name(), nil, conf)
	if err != nil {
		log.Printf("[ERROR] 提取页面文本失败: %v", err)

		// 如果提取失败，尝试使用替代方法
		log.Printf("[INFO] 尝试使用替代方法读取PDF页面内容")

		// 这里可以添加PDF内容提取的替代方法
		// 如使用不同的库或OCR处理

		return fmt.Sprintf("[页面 %d 内容无法提取，这可能是扫描PDF或图片]", pageNum), nil
	}

	// 读取文本
	data, err := os.ReadFile(tempFile.Name())
	if err != nil {
		log.Printf("[ERROR] 读取页面文本文件失败: %v", err)
		return "", fmt.Errorf("读取页面文本文件失败: %v", err)
	}

	content := string(data)
	log.Printf("[INFO] 成功提取页面 %d 内容，长度: %d 字符", pageNum, len(content))

	// 如果内容为空，返回占位文本
	if strings.TrimSpace(content) == "" {
		log.Printf("[WARN] 提取的页面内容为空")
		return fmt.Sprintf("[页面 %d 未提取到文本内容]", pageNum), nil
	}

	return content, nil
}

// readImageContent 读取图片文件内容
func readImageContent(filePath string) (string, error) {
	log.Printf("[INFO] 开始读取图片内容: %s", filePath)

	// 打开图片文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开图片文件失败: %v", err)
	}
	defer file.Close()

	// 读取图片数据
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取图片数据失败: %v", err)
	}

	// 将图片数据转换为base64字符串
	content := base64.StdEncoding.EncodeToString(data)
	log.Printf("[INFO] 图片内容读取成功，长度: %d", len(content))
	return content, nil
}

// 初始化上传目录
func init() {
	// 确保上传目录存在
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("[ERROR] 无法创建上传目录 %s: %v", uploadsDir, err)
	} else {
		log.Printf("[INFO] 上传目录已就绪: %s", uploadsDir)
	}

	// 确保分割文件目录存在
	splitDir := filepath.Join(uploadsDir, "split")
	if err := os.MkdirAll(splitDir, 0755); err != nil {
		log.Printf("[ERROR] 无法创建分割文件目录 %s: %v", splitDir, err)
	} else {
		log.Printf("[INFO] 分割文件目录已就绪: %s", splitDir)
	}

	// 检查上传目录权限
	if !directoryWritable(uploadsDir) {
		log.Printf("[WARN] 上传目录 %s 不可写", uploadsDir)
	}

	// 检查分割文件目录权限
	if !directoryWritable(splitDir) {
		log.Printf("[WARN] 分割文件目录 %s 不可写", splitDir)
	}
}

// HandleUpload 处理文件上传请求
func HandleUpload(c *gin.Context) {
	startTime := time.Now()
	log.Printf("[INFO] 收到文件上传请求 %s", c.Request.URL.Path)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("[ERROR] 获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请上传一个有效的文件",
		})
		return
	}

	log.Printf("[INFO] 收到文件: %s, 大小: %d 字节", file.Filename, file.Size)

	// 检查文件大小
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if file.Size > maxFileSize {
		log.Printf("[ERROR] 文件过大: %d 字节 (最大允许 %d 字节)", file.Size, maxFileSize)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "文件大小超过限制",
		})
		return
	}

	// 获取其他参数
	homeworkType := c.PostForm("type")
	if homeworkType == "" {
		homeworkType = "general" // 默认类型
	}

	// 获取页面布局
	layout := c.PostForm("layout")
	if layout == "" {
		layout = "single" // 默认单栏布局
	}

	// 获取每个学生的页数
	pagesPerStudentStr := c.PostForm("pagesPerStudent")
	pagesPerStudent := 1 // 默认每个学生1页
	if pagesPerStudentStr != "" {
		if num, err := strconv.Atoi(pagesPerStudentStr); err == nil && num > 0 {
			pagesPerStudent = num
		}
	}

	log.Printf("[INFO] 处理参数: type=%s, layout=%s, pagesPerStudent=%d",
		homeworkType, layout, pagesPerStudent)

	// 保存上传的文件
	savedFile, err := saveUploadedFile(file)
	if err != nil {
		log.Printf("[ERROR] 保存文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存文件失败",
		})
		return
	}

	// 关闭文件但不删除它
	savedFile.Close()
	filePath := savedFile.Name()

	log.Printf("[INFO] 文件已保存到: %s", filePath)

	// 处理学生作业
	result, err := processStudentHomework(filePath, homeworkType, layout, pagesPerStudent)
	if err != nil {
		log.Printf("[ERROR] 处理作业失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 计算处理时间
	duration := time.Since(startTime)
	log.Printf("[INFO] 文件处理完成，耗时: %v", duration)

	// 返回成功结果
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "文件上传并处理成功",
		"result":   result,
		"filePath": filePath,
	})
}

// processStudentHomework 处理学生作业
func processStudentHomework(filePath, homeworkType, layout string, pagesPerStudent int) (interface{}, error) {
	log.Printf("[INFO] 开始处理学生作业: %s, 类型: %s, 布局: %s, 每学生页数: %d",
		filePath, homeworkType, layout, pagesPerStudent)

	// 获取文件扩展名
	fileExt := strings.ToLower(filepath.Ext(filePath))

	// 根据文件类型和参数选择处理方法
	if fileExt == ".pdf" && pagesPerStudent > 0 {
		// 多页PDF文件按学生页数处理
		return processPDFWithStudents(filePath, homeworkType, pagesPerStudent)
	} else if fileExt == ".pdf" || fileExt == ".jpg" || fileExt == ".jpeg" || fileExt == ".png" {
		// 单一文件直接处理
		return processFileDirectly(filePath, homeworkType, layout)
	} else {
		log.Printf("[ERROR] 不支持的文件类型: %s", fileExt)
		return nil, fmt.Errorf("不支持的文件类型: %s", fileExt)
	}
}

// isFileNewer 检查file1是否比file2更新
func isFileNewer(file1, file2 string) bool {
	info1, err1 := os.Stat(file1)
	info2, err2 := os.Stat(file2)

	if err1 != nil || err2 != nil {
		// 如果无法获取文件信息，默认返回false
		return false
	}

	// 比较修改时间
	return info1.ModTime().After(info2.ModTime())
}

// sortFilesByModTime 按修改时间排序文件
func sortFilesByModTime(files []string) {
	sort.Slice(files, func(i, j int) bool {
		info1, _ := os.Stat(files[i])
		info2, _ := os.Stat(files[j])
		return info1.ModTime().Before(info2.ModTime())
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
