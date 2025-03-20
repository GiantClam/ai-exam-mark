package services

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Min 返回两个整数中较小的那个
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CleanMarkdownCodeBlock 从Markdown代码块中提取纯JSON内容
func CleanMarkdownCodeBlock(content string) string {
	// 记录原始内容的长度以进行调试
	originalLength := len(content)

	// 打印前100个字符，用于调试
	previewLength := Min(100, originalLength)
	log.Printf("[DEBUG] 原始内容前%d个字符: %s", previewLength, content[:previewLength])

	// 先检查是否包含反引号，这可能表示是代码块
	if strings.Contains(content, "```") {
		// 更完善的正则表达式，匹配可能出现的各种代码块格式
		// 移除开始的代码块标记，包括可能的空格和换行
		startPattern := `(?s)^[\s\n]*` + "```" + `(?:json|javascript|js)?[\s\n]*`
		content = regexp.MustCompile(startPattern).ReplaceAllString(content, "")

		// 移除结束的代码块标记，包括可能的空格和换行
		endPattern := `(?s)[\s\n]*` + "```" + `[\s\n]*$`
		content = regexp.MustCompile(endPattern).ReplaceAllString(content, "")
	}

	// 处理单行代码（使用单个反引号的情况）
	if strings.Contains(content, "`") {
		content = regexp.MustCompile("`").ReplaceAllString(content, "")
	}

	// 确保内容以JSON对象开始
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "{") {
		jsonStart := strings.Index(content, "{")
		if jsonStart >= 0 {
			content = content[jsonStart:]
		}
	}

	// 确保内容以JSON对象结束
	if !strings.HasSuffix(content, "}") {
		jsonEnd := strings.LastIndex(content, "}")
		if jsonEnd >= 0 && jsonEnd < len(content)-1 {
			content = content[:jsonEnd+1]
		}
	}

	// 打印清理后的内容长度以进行调试
	cleanedLength := len(content)
	log.Printf("[DEBUG] 清理后内容长度: %d (减少了 %d 个字符)", cleanedLength, originalLength-cleanedLength)

	return content
}

// SanitizeUTF8 清理字符串中的无效UTF-8字符
func SanitizeUTF8(s string) string {
	if s == "" {
		return s
	}

	// 先检查字符串是否有效的UTF-8
	if !utf8.ValidString(s) {
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
	return s
}

// TryFixJsonFormat 尝试修复JSON格式
func TryFixJsonFormat(jsonStr string) string {
	// 移除可能存在的非JSON前缀
	jsonStartIndex := strings.Index(jsonStr, "{")
	if jsonStartIndex > 0 {
		jsonStr = jsonStr[jsonStartIndex:]
	}

	// 确保JSON结尾是正确的
	jsonEndIndex := strings.LastIndex(jsonStr, "}")
	if jsonEndIndex >= 0 && jsonEndIndex < len(jsonStr)-1 {
		jsonStr = jsonStr[:jsonEndIndex+1]
	}

	// 处理开闭括号不匹配的情况
	openCount := strings.Count(jsonStr, "{")
	closeCount := strings.Count(jsonStr, "}")

	if openCount > closeCount {
		// 补充缺少的闭括号
		jsonStr += strings.Repeat("}", openCount-closeCount)
	}

	return jsonStr
}

// EnsureValidJSON 确保字符串为有效的JSON
// 整合了多种清理和修复技术
func EnsureValidJSON(input string) string {
	if input == "" {
		return "{}"
	}

	// 第1步：清理Markdown格式
	cleaned := CleanMarkdownCodeBlock(input)

	// 第2步：清理无效的UTF-8字符
	sanitized := SanitizeUTF8(cleaned)

	// 第3步：尝试修复JSON格式
	fixed := TryFixJsonFormat(sanitized)

	// 第4步：验证JSON是否有效
	var data interface{}
	if err := json.Unmarshal([]byte(fixed), &data); err != nil {
		log.Printf("[WARN] JSON仍然无效: %v", err)
		// 返回空对象以防止进一步错误
		return "{}"
	}

	return fixed
}
