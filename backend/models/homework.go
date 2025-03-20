package models

import "encoding/json"

// HomeworkAnswer 代表单个作业题目的答案
type HomeworkAnswer struct {
	QuestionNumber string `json:"questionNumber"`
	StudentAnswer  string `json:"studentAnswer"`
	IsCorrect      *bool  `json:"isCorrect,omitempty"`
	CorrectAnswer  string `json:"correctAnswer,omitempty"`
	CorrectSteps   string `json:"correctSteps,omitempty"`
	Explanation    string `json:"explanation,omitempty"`
	Evaluation     string `json:"evaluation,omitempty"`
	Suggestion     string `json:"suggestion,omitempty"`
}

// HomeworkResult 代表作业批改结果
type HomeworkResult struct {
	Answers      []HomeworkAnswer `json:"answers"`
	OverallScore string           `json:"overallScore,omitempty"`
	Feedback     string           `json:"feedback,omitempty"`
}

// APIResponse 表示API的通用响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

// Answer 表示一个答案的结构
type Answer struct {
	QuestionNumber string `json:"questionNumber"`
	StudentAnswer  string `json:"studentAnswer"`
}

// HomeworkResponse 表示作业批改的响应
type HomeworkResponse struct {
	Success bool            `json:"success"`
	Type    string          `json:"type"`
	Result  json.RawMessage `json:"result"`
}
