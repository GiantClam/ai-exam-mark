package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/GiantClam/homework_marking/models"
	"github.com/GiantClam/homework_marking/services"
	"github.com/gin-gonic/gin"
)

// processHomeworkImage processes the homework image with appropriate prompts based on type
func processHomeworkImage(imagePath, homeworkType string, customPrompt string) (string, error) {
	client := services.NewVertexAIClient()

	var systemInstruction string
	switch homeworkType {
	case "english":
		systemInstruction = `
		你是一位专业的英语作业批改助手。请分析学生的英语作业图片，提取其中的手写答案。
		特别注意：
		1. 重点识别括号、下划线、填空处等位置的手写答案
		2. 区分学生的手写内容和印刷的题目内容
		3. 分析答案的正确性，并给出改进建议
		4. 提供最终的得分评估

		请以下面的JSON格式回复:
		{
		  "answers": [
		    {
		      "questionNumber": "题号",
		      "studentAnswer": "学生的手写答案",
		      "isCorrect": true/false,
		      "correctAnswer": "正确答案（如果学生答错）",
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
		1. 重点识别括号、下划线、填空处等位置的手写答案
		2. 区分学生的手写内容和印刷的题目内容
		3. 分析答案的语法、用词和表达是否正确
		4. 评价答案的内容深度和创意性

		请以下面的JSON格式回复:
		{
		  "answers": [
		    {
		      "questionNumber": "题号",
		      "studentAnswer": "学生的手写答案",
		      "evaluation": "答案评价",
		      "suggestion": "改进建议"
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

	// 设置提示词 - 使用传入的自定义提示词或者创建一个默认提示词
	textPrompt := customPrompt
	if textPrompt == "" {
		textPrompt = fmt.Sprintf("这是一份%s作业，请仔细分析图片中的手写答案，特别关注括号、下划线等位置的手写内容。", homeworkType)
	}

	// 调用Gemini 2.0 Thinking模型分析图片
	response, err := client.GenerateContentWithFile(systemInstruction, imagePath, "image/jpeg", textPrompt)
	if err != nil {
		return "", err
	}

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

	// 记录环境变量信息
	log.Printf("[DEBUG] GOOGLE_APPLICATION_CREDENTIALS: %s", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	log.Printf("[DEBUG] GOOGLE_CLOUD_PROJECT: %s", os.Getenv("GOOGLE_CLOUD_PROJECT"))
	log.Printf("[DEBUG] GOOGLE_CLOUD_LOCATION: %s", os.Getenv("GOOGLE_CLOUD_LOCATION"))

	// 调用AI服务处理图片
	log.Printf("[DEBUG] 准备调用AI服务处理图片...")
	// 检查是否处于模拟模式
	log.Printf("[DEBUG] AI服务模拟模式: %v", services.UseMockMode)

	// 调用大模型处理图片
	result, err := processHomeworkImage(imagePath, homeworkType, "")
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
