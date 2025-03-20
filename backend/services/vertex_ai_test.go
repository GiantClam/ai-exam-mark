package services

import (
	"strings"
	"testing"
)

// TestMockMode 测试模拟模式下的作业批改功能
func TestMockMode(t *testing.T) {
	// 启用模拟模式
	UseMockMode = true

	// 创建Vertex AI客户端
	client := NewVertexAIClient()

	// 测试不同类型的作业批改
	testCases := []struct {
		name         string
		homeworkType string
		expectedKey  string
	}{
		{"数学作业", "数学", "解方程步骤正确"},
		{"语文作业", "语文", "背诵正确"},
		{"英语作业", "英语", "名词复数形式错误"},
		{"未知类型", "物理", "学生整体表现良好"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 构建测试提示
			prompt := "请批改这份" + tc.homeworkType + "作业"

			// 调用模拟API
			result, err := client.GenerateContentWithFile(
				"你是一个专业的作业批改助手",
				"test_file.jpg", // 虚拟文件名
				"image/jpeg",    // 虚拟MIME类型
				prompt,
			)

			// 检查是否有错误
			if err != nil {
				t.Errorf("预期返回模拟结果，但得到错误: %v", err)
				return
			}

			// 检查结果中是否包含预期的关键字
			if !strings.Contains(result, tc.expectedKey) {
				t.Errorf("预期结果包含'%s'，但没有找到。实际结果: %s",
					tc.expectedKey, result)
			}
		})
	}
}
