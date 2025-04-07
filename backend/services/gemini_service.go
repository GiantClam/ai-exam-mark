package services

import (
	"log"
)

// GeminiService 是对Vertex AI的封装
type GeminiService struct {
	vertexClient *VertexAIClient
}

// NewGeminiService 创建新的Gemini服务
func NewGeminiService() (*GeminiService, error) {
	log.Println("初始化Gemini服务...")
	
	// 创建Vertex AI客户端
	vertexClient := NewVertexAIClient()
	
	service := &GeminiService{
		vertexClient: vertexClient,
	}
	
	log.Println("Gemini服务初始化完成")
	
	return service, nil
}

// GenerateContent 生成内容
func (s *GeminiService) GenerateContent(systemInstruction, prompt string) (string, error) {
	return s.vertexClient.GenerateContent(systemInstruction, prompt)
}

// GenerateContentWithFile 使用文件生成内容
func (s *GeminiService) GenerateContentWithFile(systemInstruction, filePath, mimeType, prompt string) (string, error) {
	return s.vertexClient.GenerateContentWithFile(systemInstruction, filePath, mimeType, prompt)
}

// BuildHomeworkAnalysisPrompt 构建作业分析提示词
func (s *GeminiService) BuildHomeworkAnalysisPrompt(homeworkType string, imageContent string) string {
	return BuildHomeworkAnalysisPrompt(homeworkType, imageContent)
} 