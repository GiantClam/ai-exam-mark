package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GiantClam/homework_marking/models"
	"github.com/GiantClam/homework_marking/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 提示词模板
const (
	mathPromptTemplate = `请分析以下数学作业：
学生姓名：{{.name}}
班级：{{.class}}
作业内容：{{.content}}

请仔细分析学生的答案，特别关注：
1. 计算是否正确
2. 思路是否清晰
3. 解题方法是否合理

请以JSON格式返回分析结果。`

	englishPromptTemplate = `请分析以下英语作业：
学生姓名：{{.name}}
班级：{{.class}}
作业内容：{{.content}}

请仔细分析学生的答案，特别关注：
1. 语法是否正确
2. 拼写是否准确
3. 表达是否恰当

请以JSON格式返回分析结果。`

	chinesePromptTemplate = `请分析以下语文作业：
学生姓名：{{.name}}
班级：{{.class}}
作业内容：{{.content}}

请仔细分析学生的答案，特别关注：
1. 内容是否完整
2. 表达是否准确
3. 思路是否清晰

请以JSON格式返回分析结果。`

	generalPromptTemplate = `请分析以下作业：
学生姓名：{{.name}}
班级：{{.class}}
作业内容：{{.content}}

请仔细分析学生的答案，特别关注：
1. 答案是否准确
2. 表达是否清晰
3. 思路是否合理

请以JSON格式返回分析结果。`
)

// HomeworkHandler handles homework related requests
type HomeworkHandler struct {
	taskQueue *services.TaskQueue
	mutex     *sync.Mutex
}

// NewHomeworkHandler creates a new homework handler
func NewHomeworkHandler(taskQueue *services.TaskQueue) *HomeworkHandler {
	return &HomeworkHandler{
		taskQueue: taskQueue,
		mutex:     &sync.Mutex{},
	}
}

// UploadHomework handles homework file upload requests
func (h *HomeworkHandler) UploadHomework(c *gin.Context) {
	// 获取文件
	file, err := c.FormFile("homework")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "请上传作业文件",
		})
		return
	}

	// 检查文件类型
	filename := file.Filename
	extension := strings.ToLower(filepath.Ext(filename))
	if extension != ".pdf" && extension != ".jpg" && extension != ".jpeg" && extension != ".png" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "只支持PDF、JPG、JPEG和PNG格式的文件",
		})
		return
	}

	// 获取作业类型
	homeworkType := c.DefaultPostForm("type", "general")

	// 获取自定义提示词
	customPrompt := c.DefaultPostForm("prompt", "")

	// 获取每个学生的页数
	pagesPerStudent := 1
	if pagesPerStudentStr := c.DefaultPostForm("pagesPerStudent", "1"); pagesPerStudentStr != "" {
		if pages, err := strconv.Atoi(pagesPerStudentStr); err == nil && pages > 0 {
			pagesPerStudent = pages
		}
	}

	// 获取布局方式
	layout := c.DefaultPostForm("layout", "single")

	// 创建唯一的文件名
	uniqueID := uuid.New().String()
	uploadDir := "uploads"
	uploadPath := filepath.Join(uploadDir, uniqueID+extension)

	// 保存文件
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		log.Printf("[ERROR] 保存文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "保存文件失败",
		})
		return
	}

	// 创建异步任务
	taskID := h.taskQueue.CreateTask("homework_processing", "正在处理文件...")

	// 立即返回任务ID
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"taskId":  taskID,
			"message": "文件上传成功，正在处理中...",
		},
	})

	// 异步处理文件
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ERROR] 处理文件时发生异常: %v", r)
				h.taskQueue.UpdateTaskStatus(taskID, "error", fmt.Sprintf("处理文件时发生异常: %v", r))
			}
		}()

		// 更新任务状态
		h.taskQueue.UpdateTaskStatus(taskID, "processing", "正在分析文件内容...")

		// 处理文件，根据文件类型选择不同的处理方式
		// var result string
		var err error

		if extension == ".pdf" {
			// PDF处理逻辑
			_, err = h.processPDFHomework(uploadPath, homeworkType, customPrompt, pagesPerStudent, layout)
		} else {
			// 图片处理逻辑
			_, err = h.processImageHomework(uploadPath, homeworkType, customPrompt)
		}

		if err != nil {
			log.Printf("[ERROR] 处理文件失败: %v", err)
			h.taskQueue.UpdateTaskStatus(taskID, "error", fmt.Sprintf("处理文件失败: %v", err))
			return
		}

		// 更新任务状态为完成
		//h.taskQueue.UpdateTaskStatus(taskID, "completed", result)
	}()
}

// 处理PDF作业
func (h *HomeworkHandler) processPDFHomework(pdfPath, homeworkType, customPrompt string, pagesPerStudent int, layout string) (string, error) {
	// 实现PDF处理逻辑
	log.Printf("[INFO] 处理PDF作业: %s, 类型: %s", pdfPath, homeworkType)

	// 创建任务记录
	taskID := h.taskQueue.CreateTask("pdf_processing", "正在处理PDF文件...")

	// 检查文件是否存在
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("PDF文件不存在: %s", pdfPath)
		log.Printf("[ERROR] %s", errMsg)
		h.taskQueue.FailTask(taskID, errMsg)
		return "", fmt.Errorf(errMsg)
	}

	// 创建AI客户端
	client := services.NewVertexAIClient()

	// 获取系统指令
	systemInstruction := getSystemInstructionByType(homeworkType)

	// 创建临时目录用于分割的PDF文件
	splitDir := filepath.Join("uploads", "split")

	// 按照学生页数拆分PDF
	studentPDFs, err := services.SplitPDF(pdfPath, pagesPerStudent, splitDir)
	if err != nil {
		errMsg := fmt.Sprintf("拆分PDF失败: %v", err)
		log.Printf("[ERROR] %s", errMsg)
		h.taskQueue.FailTask(taskID, errMsg)
		return "", fmt.Errorf(errMsg)
	}

	// 更新任务状态
	totalStudents := len(studentPDFs)
	h.taskQueue.UpdateTaskStatus(taskID, "processing", fmt.Sprintf("正在处理，总共%d个学生", totalStudents))
	h.taskQueue.UpdateTaskTotalStudents(taskID, totalStudents)

	// 用于保存每个学生的处理结果
	results := make([]string, totalStudents)

	// 创建一个等待组来同步所有goroutine
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex

	// 处理每个学生的PDF - 按照索引顺序处理
	for studentIdx, studentPDF := range studentPDFs {
		wg.Add(1)

		// 为每个学生创建AI分析任务
		go func(studentIdx int, pdfPath string) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[ERROR] 处理学生作业时发生panic: %v", r)
				}
			}()

			// 处理单个学生的作业
			log.Printf("[INFO] 开始处理学生 %d 的作业: %s", studentIdx+1, pdfPath)

			// 设置提示词
			textPrompt := customPrompt
			if textPrompt == "" {
				textPrompt = fmt.Sprintf("这是一份%s作业，请分析PDF中的内容。这是学生%d的作业。请从上到下处理，整理所有答案。",
					homeworkType, studentIdx+1)
			}

			// 调用AI模型分析PDF（添加重试机制）
			var response string
			var err error
			maxRetries := 3

			for attempt := 0; attempt <= maxRetries; attempt++ {
				if attempt > 0 {
					log.Printf("[INFO] 第%d次重试调用大模型处理学生%d的PDF...",
						attempt, studentIdx+1)
					// 指数退避策略
					backoffTime := time.Duration(attempt*attempt) * time.Second
					time.Sleep(backoffTime)
				}

				// 模拟模式下返回测试数据
				if services.UseMockMode {
					log.Printf("[INFO] 模拟模式: 生成模拟结果代替调用大模型")
					mockData := fmt.Sprintf(`{
  "answers": [
    {
								"questionNumber": "1",
								"studentAnswer": "模拟答案 - 学生%d PDF作业",
								"isCorrect": true,
								"correctAnswer": "标准答案",
								"explanation": "这是一个测试解释"
							}
						],
						"overallScore": "90",
						"feedback": "模拟反馈: 学生%d PDF作业表现良好"
					}`, studentIdx+1, studentIdx+1)
					response = mockData
					break
				} else {
					// 调用大模型API处理PDF文件
					response, err = services.GenerateContentWithPDF(client, systemInstruction, pdfPath, textPrompt)
				}

				if err == nil {
					log.Printf("[INFO] 成功获取学生%d的大模型分析结果", studentIdx+1)
					break // 成功获取响应，退出重试循环
				}

				log.Printf("[ERROR] 调用大模型处理PDF失败 (尝试%d/%d): %v",
					attempt+1, maxRetries+1, err)

				if attempt == maxRetries {
					log.Printf("[WARN] 达到最大重试次数，无法处理学生%d的作业",
						studentIdx+1)
				}
			}

			// 如果获取到响应
			if err == nil && response != "" {
				log.Printf("[INFO] 成功处理学生 %d 的作业", studentIdx+1)

				// 解析JSON字符串为对象
				var responseObj map[string]interface{}
				err := json.Unmarshal([]byte(response), &responseObj)
				if err == nil {
					// 移除 "uploads/split/" 路径前缀
					cleanPath := pdfPath
					if strings.HasPrefix(cleanPath, "uploads/split/") {
						cleanPath = strings.TrimPrefix(cleanPath, "uploads/split/")
					}

					// 添加PDF文件路径到响应对象
					responseObj["pdfUrl"] = cleanPath

					// 将对象转换回JSON字符串
					updatedResponse, jsonErr := json.Marshal(responseObj)
					if jsonErr == nil {
						response = string(updatedResponse)
					} else {
						log.Printf("[ERROR] 将更新后的响应转换为JSON失败: %v", jsonErr)
					}
				} else {
					log.Printf("[ERROR] 解析学生 %d 的响应JSON失败: %v", studentIdx+1, err)
				}

				// 更新处理计数
				h.taskQueue.IncrementProcessedCount(taskID)

				// 锁定添加结果
				resultsMutex.Lock()
				// 保存结果到正确的索引位置
				results[studentIdx] = response
				resultsMutex.Unlock()
			} else {
				log.Printf("[ERROR] 处理学生 %d 作业失败: %v", studentIdx+1, err)
				// 即使处理失败，也在结果数组中保留位置
				resultsMutex.Lock()
				results[studentIdx] = ""
				resultsMutex.Unlock()
			}
		}(studentIdx, studentPDF)
	}

	// 等待所有处理完成
	wg.Wait()

	// 合并结果 - 按原始索引顺序合并
	combinedResults := "["
	validResultCount := 0

	for i, result := range results {
		// 如果某个位置为空（处理失败），跳过它
		if result == "" {
			continue
		}

		validResultCount++

		// 去除首尾可能存在的中括号和多余空格
		result = strings.TrimSpace(result)
		if strings.HasPrefix(result, "[") {
			result = strings.TrimPrefix(result, "[")
		}
		if strings.HasSuffix(result, "]") {
			result = strings.TrimSuffix(result, "]")
		}

		// 添加到合并结果中
		combinedResults += result

		// 如果不是最后一个结果，添加逗号分隔
		if i < len(results)-1 {
			// 检查后面是否还有非空结果
			hasMoreResults := false
			for j := i + 1; j < len(results); j++ {
				if results[j] != "" {
					hasMoreResults = true
					break
				}
			}
			if hasMoreResults {
				combinedResults += ","
			}
		}
	}
	combinedResults += "]"

	// 完成任务
	h.taskQueue.CompleteTask(taskID, combinedResults)

	// 返回任务ID，前端可以轮询任务状态
	return taskID, nil
}

// 根据作业类型获取系统指令
func getSystemInstructionByType(homeworkType string) string {
	var systemInstruction string
	switch homeworkType {
	case "english":
		systemInstruction = `
		你是一位专业的英语老师。
		特别注意：
		1. 重点识别模拟题
		2. 区分学生的手写内容和印刷的题目内容
		3. 判断答案的英语作业图片，提取其中的手写答案。
		4. 从作业中提取学生姓名和班级信息（通常在作业右上角或左上角）
		
		请以下面的JSON格式回答：
{
  "name": "学生姓名（如果能识别）",
  "class": "班级（如果能识别）",
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的手写答案",
      "isCorrect": true/false,
      "correctAnswer": "正确答案（可双栏布局）",
      "explanation": "简短答案解释"
    }
  ],
  "overallScore": "总得分（必须使用百分制，0-100之间的数字，不要带百分号）",
  "feedback": "整体评价和建议"
}

		请只返回标准JSON格式数据，不要使用Markdown代码块；不要分析右半部分的内容，从上到下处理。`
	case "math":
		systemInstruction = `
		你是一位专业的数学老师。
		特别注意：
		1. 重点识别模拟题
		2. 区分学生的手写内容和印刷的题目内容
		3. 识别数学符号
		4. 判断数学公式
		5. 从作业中提取学生姓名和班级信息（通常在作业右上角或左上角）
		
		请以下面的JSON格式回答：
{
  "name": "学生姓名（如果能识别）",
  "class": "班级（如果能识别）",
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的手写答案",
      "isCorrect": true/false,
      "correctAnswer": "正确答案",
      "explanation": "简短答案解释"
    }
  ],
  "overallScore": "总得分（必须使用百分制，0-100之间的数字，不要带百分号）",
  "feedback": "整体评价和建议"
		}`
	case "chinese":
		systemInstruction = `
		你是一位专业的语文老师。
		特别注意：
		1. 重点识别文章和练习题
		2. 区分学生的手写内容和印刷的题目内容
		3. 评判文字表达是否准确得体
		4. 从作业中提取学生姓名和班级信息（通常在作业右上角或左上角）
		
		请以下面的JSON格式回答：
{
  "name": "学生姓名（如果能识别）",
  "class": "班级（如果能识别）",
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的手写答案",
      "isCorrect": true/false,
      "correctAnswer": "正确答案",
      "explanation": "简短答案解释"
    }
  ],
  "overallScore": "总得分（必须使用百分制，0-100之间的数字，不要带百分号）",
  "feedback": "整体评价和建议"
		}`
	default:
		systemInstruction = `
		请分析学生的作业图片，提取其中的内容。
		特别注意：
		1. 重点识别模拟题
		2. 区分学生的手写内容和印刷的题目内容
		3. 从作业中提取学生姓名和班级信息（通常在作业右上角或左上角）
		
		请以下面的JSON格式回答：
{
  "name": "学生姓名（如果能识别）",
  "class": "班级（如果能识别）",
  "answers": [
    {
      "questionNumber": "题号",
      "studentAnswer": "学生的手写答案",
      "isCorrect": true/false,
      "correctAnswer": "正确答案",
      "explanation": "简短答案解释"
    }
  ],
  "overallScore": "总得分（必须使用百分制，0-100之间的数字，不要带百分号）",
  "feedback": "整体评价和建议"
		}`
	}
	return systemInstruction
}

// 计算总分
func calculateOverallScore(answers []map[string]interface{}) string {
	if len(answers) == 0 {
		return "0"
	}

	correctCount := 0
	for _, answer := range answers {
		if isCorrect, ok := answer["isCorrect"].(bool); ok && isCorrect {
			correctCount++
		}
	}

	// 计算百分比
	percentage := float64(correctCount) / float64(len(answers)) * 100
	return fmt.Sprintf("%.1f", percentage)
}

// 生成反馈
func generateFeedback(answers []map[string]interface{}) string {
	if len(answers) == 0 {
		return "未检测到答案内容"
	}

	correctCount := 0
	for _, answer := range answers {
		if isCorrect, ok := answer["isCorrect"].(bool); ok && isCorrect {
			correctCount++
		}
	}

	// 计算正确率
	correctRate := float64(correctCount) / float64(len(answers))

	if correctRate >= 0.8 {
		return "整体表现优秀，继续保持！"
	} else if correctRate >= 0.6 {
		return "整体表现良好，但仍有提升空间。"
	} else {
		return "需要更多练习，建议重点复习错题。"
	}
}

// 创建模拟图片文件（仅用于演示）
func createMockImageFile(imagePath string) error {
	// 确保目录存在
	dir := filepath.Dir(imagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建一个简单的文本文件模拟图片
	// 实际使用中，这里应该是从PDF转换得到的真实图片
	return os.WriteFile(imagePath, []byte("Mock image file for testing"), 0644)
}

// 将PDF分割为图片
func splitPDFToImages(pdfPath, outputDir string) ([]string, error) {
	// 这是一个占位实现，实际项目中需要使用PDF库
	// 如UniDoc、pdfcpu、GhostScript或其他方式分割PDF

	log.Printf("[INFO] 分割PDF到图片: %s -> %s", pdfPath, outputDir)

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 检查PDF文件是否存在
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("PDF文件不存在: %s", pdfPath)
	}

	// 实际实现中，这里应该使用PDF库读取文件并计算页数
	// 然后将每页转换为图片并保存

	// 这里仅返回模拟的3页图片路径作为示例
	imageFiles := []string{
		filepath.Join(outputDir, "page1.jpg"),
		filepath.Join(outputDir, "page2.jpg"),
		filepath.Join(outputDir, "page3.jpg"),
	}

	// 创建模拟图片文件
	for _, path := range imageFiles {
		if err := createMockImageFile(path); err != nil {
			return nil, fmt.Errorf("创建图片文件失败: %v", err)
		}
	}

	log.Printf("[INFO] PDF分割完成，生成了 %d 张图片", len(imageFiles))
	return imageFiles, nil
}

// 处理图片作业
func (h *HomeworkHandler) processImageHomework(imagePath, homeworkType, customPrompt string) (string, error) {
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
	systemInstruction := getSystemInstructionByType(homeworkType)
	log.Printf("[DEBUG] 系统指令长度: %d 字符", len(systemInstruction))

	// 设置提示词 - 使用传入的自定义提示词或者创建一个基本提示词
	textPrompt := customPrompt
	if textPrompt == "" {
		textPrompt = fmt.Sprintf("这是一份%s作业，请分析图片中的内容，从上到下处理。", homeworkType)
	}

	log.Printf("[DEBUG] 提示词长度: %d 字符", len(textPrompt))

	// 调用Gemini模型分析图片（添加重试机制）
	var response string
	var err error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[INFO] 第%d次重试调用大模型处理图片...", attempt)
			// 指数退避策略
			backoffTime := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoffTime)
		}

		// 调用大模型API
		response, err = client.GenerateContentWithFile(systemInstruction, imagePath, "image/jpeg", textPrompt)

		if err == nil {
			log.Printf("[INFO] 成功获取大模型分析结果")
			break // 成功获取响应，退出重试循环
		}

		log.Printf("[ERROR] 调用大模型处理图片失败 (尝试%d/%d): %v",
			attempt+1, maxRetries+1, err)

		if attempt == maxRetries {
			log.Printf("[ERROR] 达到最大重试次数，处理图片失败")
			return "", fmt.Errorf("AI服务处理失败: %v", err)
		}
	}

	// 验证返回的JSON格式
	var jsonResult interface{}
	if err := json.Unmarshal([]byte(response), &jsonResult); err != nil {
		log.Printf("[ERROR] 解析AI返回的JSON失败: %v", err)
		// 记录部分原始响应以便调试
		if len(response) > 200 {
			log.Printf("[DEBUG] 原始响应前200字符: %s", response[:200])
		} else {
			log.Printf("[DEBUG] 原始响应: %s", response)
		}
		return "", fmt.Errorf("解析AI返回的JSON失败: %v", err)
	}

	log.Printf("[DEBUG] 成功处理作业图片，返回结果长度: %d 字符", len(response))
	return response, nil
}

// MarkHomework handles homework marking requests
func (h *HomeworkHandler) MarkHomework(c *gin.Context) {
	// 获取文件
	file, err := c.FormFile("homework")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "请上传作业文件",
		})
		return
	}

	// 获取布局参数（记录但本例中未使用）
	_ = c.DefaultPostForm("layout", "single")

	// 获取作业类型（记录但本例中未使用）
	_ = c.DefaultPostForm("type", "general")

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

	// 这里简化实现，直接返回成功响应
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Result:  "作业已接收，正在处理中",
	})
}

// 检查同一学生的文件是否已在处理中
func (h *HomeworkHandler) isProcessing(studentName string) bool {
	// 通过获取所有任务来检查是否存在处理中的相同学生名称的任务
	allTasks := h.getAllTasks()
	for _, task := range allTasks {
		// 只检查处理中的任务
		if task.Status == services.TaskStatusProcessing || task.Status == services.TaskStatusPending {
			// 检查任务参数中是否包含相同的学生名称
			if params, ok := task.Params["studentName"]; ok {
				if studentNameVal, ok := params.(string); ok && studentNameVal == studentName {
					return true
				}
			}
		}
	}
	return false
}

// 辅助方法：获取所有任务
func (h *HomeworkHandler) getAllTasks() []*services.HomeworkTask {
	return h.taskQueue.GetAllTasks()
}
