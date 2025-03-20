package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"image"

	"github.com/GiantClam/homework_marking/models"
	"github.com/GiantClam/homework_marking/services"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

// processHomeworkImage processes the homework image with appropriate prompts based on type
func processHomeworkImage(imagePath, homeworkType string) (string, error) {
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
		
		请只返回JSON格式数据，不要添加任何其他文本。确保JSON格式完整有效。`

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
		
		请只返回JSON格式数据，不要添加任何其他文本。确保JSON格式完整有效。`

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
		
		请只返回JSON格式数据，不要添加任何其他文本。确保JSON格式完整有效。`

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
		
		请只返回JSON格式数据，不要添加任何其他文本。确保JSON格式完整有效。`
	}

	// 设置提示词
	textPrompt := fmt.Sprintf("这是一份%s作业，请仔细分析图片中的手写答案，特别关注括号、下划线等位置的手写内容。", homeworkType)

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

// 处理单栏布局图片
func processSingleColumnImage(imagePath, homeworkType string) (interface{}, error) {
	// 这里实现单栏布局的处理逻辑
	// 直接从上到下处理整个图片
	return processImage(imagePath, homeworkType)
}

// 处理双栏布局图片
func processDoubleColumnImage(imagePath, homeworkType string) (interface{}, error) {
	// 这里实现双栏布局的处理逻辑
	// 1. 将图片分成左右两栏
	// 2. 先处理左栏（从上到下）
	// 3. 再处理右栏（从上到下）
	// 4. 合并结果

	// 读取图片
	img, err := imaging.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开图片: %v", err)
	}

	// 获取图片尺寸
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// 分割图片为左右两栏
	leftImage := imaging.Crop(img, image.Rect(0, 0, width/2, height))
	rightImage := imaging.Crop(img, image.Rect(width/2, 0, width, height))

	// 保存临时文件
	leftPath := imagePath + "-left.jpg"
	rightPath := imagePath + "-right.jpg"

	err = imaging.Save(leftImage, leftPath)
	if err != nil {
		return nil, fmt.Errorf("无法保存左栏图片: %v", err)
	}
	defer os.Remove(leftPath)

	err = imaging.Save(rightImage, rightPath)
	if err != nil {
		return nil, fmt.Errorf("无法保存右栏图片: %v", err)
	}
	defer os.Remove(rightPath)

	// 处理左栏
	leftResult, err := processImage(leftPath, homeworkType)
	if err != nil {
		return nil, fmt.Errorf("处理左栏失败: %v", err)
	}

	// 处理右栏
	rightResult, err := processImage(rightPath, homeworkType)
	if err != nil {
		return nil, fmt.Errorf("处理右栏失败: %v", err)
	}

	// 合并结果
	// 这里需要根据具体的返回类型进行合并
	// 假设返回的是 []Answer 类型
	leftAnswers := leftResult.([]models.Answer)
	rightAnswers := rightResult.([]models.Answer)

	// 合并答案，保持左栏在前，右栏在后
	mergedAnswers := append(leftAnswers, rightAnswers...)

	return mergedAnswers, nil
}

// 处理单张图片
func processImage(imagePath, homeworkType string) (interface{}, error) {
	// 这里实现具体的图片处理逻辑
	// 可以调用 Gemini API 或其他 OCR 服务
	// 返回处理结果
	return nil, nil
}
